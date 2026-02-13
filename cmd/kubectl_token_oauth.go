package cmd

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	apiv3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const (
	oauthCodeFlowTimeout          = 5 * time.Minute
	oauthCodeExchangeTimeout      = 30 * time.Second
	callbackServerShutdownTimeout = time.Second
)

// oauthAuth dispatches the OAuth authentication flow based on the auth flow type.
func oauthAuth(client *http.Client, input *LoginInput, provider TypedProvider, useV1Public bool) (*managementClient.Token, error) {
	if input.authFlow == "" { // The flag has precedence over the env variable.
		input.authFlow = os.Getenv("CATTLE_OAUTH_AUTH_FLOW")
	}
	input.authFlow = strings.ToLower(input.authFlow)

	switch input.authFlow {
	case authCodeFlow:
		return oauthAuthCodeAuth(client, input, provider, oauthCodeFlowTimeout, useV1Public)
	case deviceAuthFlow, "": // Default to device code flow if not specified.
		return oauthDeviceCodeAuth(client, input, provider, useV1Public)
	default:
		return nil, fmt.Errorf("invalid auth-flow value: %s", input.authFlow)
	}
}

// oauthAuthCodeAuth implements the authorization code flow for OAuth authentication.
func oauthAuthCodeAuth(client *http.Client, input *LoginInput, provider TypedProvider, timeoutAfter time.Duration, useV1Public bool) (*managementClient.Token, error) {
	oauthConfig, err := newOauthConfig(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create oauth config: %w", err)
	}

	// Generate PKCE verifier (43-128 chars, cryptographically random).
	verifier := oauth2.GenerateVerifier()

	var callbackPort int
	if v := os.Getenv("CATTLE_OAUTH_CALLBACK_PORT"); v != "" {
		callbackPort, err = strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid callback port value: %w", err)
		}
		if callbackPort < 0 || callbackPort > 65535 {
			return nil, errors.New("callback port value must be between 0 and 65535")
		}
		if callbackPort > 0 && callbackPort < 1024 {
			logrus.Warnf("Using privileged port %d may require elevated permissions", callbackPort)
		}
	}

	// Start a local callback server on a random port.
	// Note: RFC 8252 Section 7.3 explicitly allows HTTP for localhost
	// https://datatracker.ietf.org/doc/html/rfc8252#section-7.3
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", callbackPort))
	if err != nil {
		return nil, fmt.Errorf("failed to start local callback server on port %d: %w", callbackPort, err)
	}
	defer listener.Close()

	if callbackPort == 0 {
		callbackPort = listener.Addr().(*net.TCPAddr).Port
	}

	oauthConfig.RedirectURL = fmt.Sprintf("http://localhost:%d", callbackPort)

	// Generate state for CSRF protection.
	state, err := generateState()
	if err != nil {
		return nil, err
	}

	// Build the authorization URL with PKCE challenge.
	authURL := oauthConfig.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))

	// Start the callback server.
	resultCh := make(chan callbackResult, 1)
	srv := startCallbackServer(listener, state, resultCh)
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), callbackServerShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			srv.Close() // Force close if graceful shutdown fails.
		}
	}()

	// Open the user's browser.
	customPrint("\nOpening browser for authentication...\n")
	if err := openBrowser(authURL); err != nil {
		logrus.Debugf("Failed to open browser: %v", err)
		customPrint(fmt.Sprintf("Failed to open browser automatically. Please open the following URL manually:\n%s\n", authURL))
	}

	// Wait for the callback with the authorization code.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer signal.Stop(interrupt)

	timeout := time.NewTimer(timeoutAfter)
	defer timeout.Stop()

	select {
	case result := <-resultCh:
		if result.err != nil {
			return nil, result.err
		}

		// Exchange the authorization code for tokens using the PKCE verifier.
		ctx, cancel := context.WithTimeout(context.Background(), oauthCodeExchangeTimeout)
		defer cancel()
		ctx = context.WithValue(ctx, oauth2.HTTPClient, client)

		oauthToken, err := oauthConfig.Exchange(ctx, result.code, oauth2.VerifierOption(verifier))
		if err != nil {
			return nil, fmt.Errorf("failed to exchange authorization code for token: %w", err)
		}

		// Send the id_token to Rancher to get a Rancher token.
		return rancherLogin(client, input, oauthToken, useV1Public)
	case <-timeout.C:
		return nil, errors.New("timed out waiting for browser authentication")
	case <-interrupt:
		return nil, errors.New("authentication interrupted by user")
	}
}

// generateState creates a random string to be used as the OAuth state parameter for CSRF protection.
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// openBrowser attempts to open the specified URL in the user's default browser, with support for Windows, macOS, and Linux.
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", "", url}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return exec.Command(cmd, args...).Start()
}

// callbackResult is used to communicate the result of the OAuth callback handling back to the main authentication flow.
type callbackResult struct {
	code  string
	state string
	err   error
}

// startCallbackServer starts an HTTP server to listen for the OAuth callback and validates the state parameter for CSRF protection.
func startCallbackServer(listener net.Listener, expectedState string, resultCh chan<- callbackResult) *http.Server {
	mux := http.NewServeMux()
	srv := &http.Server{
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	var once sync.Once

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				resultCh <- callbackResult{err: fmt.Errorf("panic in callback handler: %v", r)}
			}
		}()

		once.Do(func() {
			query := r.URL.Query()

			// Check for error response from the IdP.
			if errCode := query.Get("error"); errCode != "" {
				errDesc := query.Get("error_description")
				fmt.Fprintf(w, "<html><body><h1>Authentication Failed</h1><p>%s: %s</p><p>You can close this window.</p></body></html>", errCode, errDesc)
				resultCh <- callbackResult{err: fmt.Errorf("authentication error: %s: %s", errCode, errDesc)}
				return
			}

			// Validate state for CSRF protection.
			state := query.Get("state")
			if state != expectedState {
				http.Error(w, "Invalid state parameter", http.StatusBadRequest)
				resultCh <- callbackResult{err: errors.New("invalid state parameter in callback (possible CSRF attack)")}
				return
			}

			code := query.Get("code")
			if code == "" {
				http.Error(w, "Missing authorization code", http.StatusBadRequest)
				resultCh <- callbackResult{err: errors.New("missing authorization code in callback")}
				return
			}

			fmt.Fprint(w, "<html><body><h1>Authentication Successful</h1><p>You can close this window and return to the terminal.</p></body></html>")
			resultCh <- callbackResult{code: code, state: state}
		})
	})

	go srv.Serve(listener)

	return srv
}

// oauthDeviceCodeAuth implements the device code flow for OAuth authentication.
func oauthDeviceCodeAuth(client *http.Client, input *LoginInput, provider TypedProvider, useV1Public bool) (*managementClient.Token, error) {
	oauthConfig, err := newOauthConfig(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create oauth config: %w", err)
	}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, client) // Set the custom HTTP client.

	deviceAuthResp, err := oauthConfig.DeviceAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate device authorization: %w", err)
	}

	customPrint(fmt.Sprintf(
		"\nTo sign in, use a web browser to open the page %s and enter the code %s to authenticate.\n",
		deviceAuthResp.VerificationURI,
		deviceAuthResp.UserCode,
	))

	oauthToken, err := oauthConfig.DeviceAccessToken(ctx, deviceAuthResp)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve access token: %w", err)
	}

	token, err := rancherLogin(client, input, oauthToken, useV1Public)
	if err != nil {
		return nil, fmt.Errorf("error during rancher login: %w", err)
	}

	return token, nil
}

func newOauthConfig(provider TypedProvider) (*oauth2.Config, error) {
	var oauthProvider apiv3.OAuthProvider

	switch p := provider.(type) {
	case *apiv3.AzureADProvider:
		oauthProvider = p.OAuthProvider
	default:
		return nil, fmt.Errorf("provider %s is not a supported OAuth provider", provider.GetType())
	}

	return &oauth2.Config{
		ClientID: oauthProvider.ClientID,
		Scopes:   oauthProvider.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:       oauthProvider.AuthURL,
			DeviceAuthURL: oauthProvider.DeviceAuthURL,
			TokenURL:      oauthProvider.TokenURL,
		},
	}, nil
}

// rancherLogin sends the obtained OAuth token to Rancher to exchange it for a Rancher token that can be used for API authentication.
func rancherLogin(client *http.Client, input *LoginInput, oauthToken *oauth2.Token, useV1Public bool) (*managementClient.Token, error) {
	reqURL := fmt.Sprintf(loginURL, input.server)
	if !useV1Public {
		providerName := strings.ToLower(strings.TrimSuffix(input.authProvider, "Provider"))
		reqURL = fmt.Sprintf(loginURLv3, input.server, input.authProvider, providerName)
	}

	responseType := "kubeconfig"
	if input.clusterID != "" {
		responseType = fmt.Sprintf("%s_%s", responseType, input.clusterID)
	}

	reqBody, err := json.Marshal(map[string]any{
		"type":         input.authProvider,
		"responseType": responseType,
		"id_token":     oauthToken.Extra("id_token"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, respBody, err := doRequest(client, req)
	if err == nil && resp.StatusCode != http.StatusCreated {
		err = fmt.Errorf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	if err != nil {
		return nil, err
	}

	token := &managementClient.Token{}
	err = json.Unmarshal(respBody, token)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling login response: %w", err)
	}

	return token, nil
}

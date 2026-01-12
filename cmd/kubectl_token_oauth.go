package cmd

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	apiv3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"golang.org/x/oauth2"
)

const (
	// OAuth flow types
	OAuthFlowDevice   = "device"
	OAuthFlowAuthCode = "authcode"

	// Timeouts
	AuthTimeout           = 5 * time.Minute
	ServerShutdownTimeout = 5 * time.Second
)

func oauthAuth(client *http.Client, input *LoginInput, provider TypedProvider) (*managementClient.Token, error) {
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, client) // Set the custom HTTP client.

	var oauthToken *oauth2.Token
	var err error

	// Determine which OAuth flow to use
	switch input.oauthFlow {
	case OAuthFlowDevice:
		oauthToken, err = performDeviceCodeFlow(ctx, provider)
	case OAuthFlowAuthCode:
		oauthToken, err = performAuthCodeFlow(ctx, provider, client, input.oauthCallbackPort)
	default:
		return nil, fmt.Errorf("invalid oauth flow: %s. Valid values are '%s' or '%s'", input.oauthFlow, OAuthFlowAuthCode, OAuthFlowDevice)
	}

	if err != nil {
		return nil, err
	}

	token, err := rancherLogin(client, input, provider, oauthToken)
	if err != nil {
		return nil, fmt.Errorf("error during rancher login: %w", err)
	}

	return token, nil
}

// performDeviceCodeFlow implements the device code OAuth flow
func performDeviceCodeFlow(ctx context.Context, provider TypedProvider) (*oauth2.Token, error) {
	// For device flow, the port doesn't matter, so we use 0 as a placeholder
	oauthConfig, err := newOauthConfig(provider, 0)
	if err != nil {
		return nil, err
	}

	deviceAuthResp, err := oauthConfig.DeviceAuth(ctx)
	if err != nil {
		return nil, err
	}

	customPrint(fmt.Sprintf(
		"\nTo sign in, use a web browser to open the page %s and enter the code %s to authenticate.\n",
		deviceAuthResp.VerificationURI,
		deviceAuthResp.UserCode,
	))

	oauthToken, err := oauthConfig.DeviceAccessToken(ctx, deviceAuthResp)
	if err != nil {
		return nil, err
	}

	return oauthToken, nil
}

// performAuthCodeFlow implements the authorization code OAuth flow with PKCE
func performAuthCodeFlow(ctx context.Context, provider TypedProvider, client *http.Client, callbackPort int) (*oauth2.Token, error) {
	oauthConfig, err := newOauthConfig(provider, callbackPort)
	if err != nil {
		return nil, err
	}

	// Generate PKCE code verifier and challenge
	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}
	codeChallenge := generateCodeChallenge(codeVerifier)

	// Start local callback server
	authCode, err := performAuthCodeCallbackFlow(ctx, oauthConfig, codeChallenge, callbackPort)
	if err != nil {
		return nil, fmt.Errorf("failed to perform auth code flow: %w", err)
	}

	// Exchange authorization code for token
	oauthToken, err := exchangeCodeForToken(ctx, oauthConfig, authCode, codeVerifier)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return oauthToken, nil
}

func newOauthConfig(provider TypedProvider, callbackPort int) (*oauth2.Config, error) {
	var oauthProvider apiv3.OAuthProvider

	switch p := provider.(type) {
	case *apiv3.AzureADProvider:
		oauthProvider = p.OAuthProvider
	default:
		return nil, fmt.Errorf("provider %s is not a supported OAuth provider", provider.GetType())
	}

	// Use the specified port for the local callback server
	redirectURI := fmt.Sprintf("http://localhost:%d/callback", callbackPort)

	return &oauth2.Config{
		ClientID:    oauthProvider.ClientID,
		Scopes:      oauthProvider.Scopes,
		RedirectURL: redirectURI,
		Endpoint: oauth2.Endpoint{
			AuthURL:       oauthProvider.AuthURL,
			TokenURL:      oauthProvider.TokenURL,
			DeviceAuthURL: oauthProvider.DeviceAuthURL,
		},
	}, nil
}

func rancherLogin(client *http.Client, input *LoginInput, provider TypedProvider, oauthToken *oauth2.Token) (*managementClient.Token, error) {
	// login with id_token
	providerName := strings.ToLower(strings.TrimSuffix(input.authProvider, "Provider"))
	url := fmt.Sprintf("%s/v3-public/%ss/%s?action=login", input.server, provider.GetType(), providerName)

	responseType := "kubeconfig"
	if input.clusterID != "" {
		responseType = fmt.Sprintf("%s_%s", responseType, input.clusterID)
	}

	jsonBody, err := json.Marshal(map[string]interface{}{
		"responseType": responseType,
		"id_token":     oauthToken.Extra("id_token"),
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, respBody, err := doRequest(client, req)
	if err == nil && resp.StatusCode != http.StatusCreated {
		err = fmt.Errorf("unexpected http status code %d", resp.StatusCode)
	}
	if err != nil {
		return nil, err
	}

	token := &managementClient.Token{}
	err = json.Unmarshal(respBody, token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// generateCodeVerifier generates a random code verifier for PKCE
func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// generateCodeChallenge generates the code challenge from the verifier using S256
func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// performAuthCodeCallbackFlow starts the local callback server and opens the browser
func performAuthCodeCallbackFlow(ctx context.Context, config *oauth2.Config, codeChallenge string, callbackPort int) (string, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", callbackPort))
	if err != nil {
		return "", fmt.Errorf("failed to start local server on port %d: %w", callbackPort, err)
	}
	defer listener.Close()

	// Channel to receive the authorization code
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Create HTTP server for callback
	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/callback" {
				http.NotFound(w, r)
				return
			}

			code := r.URL.Query().Get("code")
			if code == "" {
				errMsg := r.URL.Query().Get("error")
				if errMsg == "" {
					errMsg = "authorization code not found"
				}
				errChan <- fmt.Errorf("authentication failed: %s", errMsg)
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "<html><body><h1>Authentication Failed</h1><p>%s</p><p>You can close this window.</p></body></html>", errMsg)
				return
			}

			codeChan <- code
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "<html><body><h1>Authentication Successful</h1><p>You can close this window and return to the CLI.</p></body></html>")
		}),
	}

	// Start server in background
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("callback server error: %w", err)
		}
	}()

	// Build authorization URL with PKCE
	authURL := config.AuthCodeURL("state",
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"))

	customPrint(fmt.Sprintf("\nOpening browser for authentication at:\n%s\n", authURL))
	customPrint("If the browser doesn't open automatically, please visit the URL above.\n")

	// Open browser
	if err := openBrowser(authURL); err != nil {
		customPrint(fmt.Sprintf("Failed to open browser automatically: %v\n", err))
		customPrint(fmt.Sprintf("Please open the following URL manually:\n%s\n", authURL))
	}

	// Wait for callback with timeout
	timeout := time.After(AuthTimeout)
	var authCode string

	select {
	case authCode = <-codeChan:
		// Success
	case err := <-errChan:
		return "", err
	case <-timeout:
		return "", fmt.Errorf("timeout waiting for authentication")
	case <-ctx.Done():
		return "", ctx.Err()
	}

	// Shutdown server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), ServerShutdownTimeout)
	defer cancel()
	server.Shutdown(shutdownCtx)

	return authCode, nil
}

// exchangeCodeForToken exchanges the authorization code for an access token
func exchangeCodeForToken(ctx context.Context, config *oauth2.Config, code, codeVerifier string) (*oauth2.Token, error) {
	// Build token request
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", config.RedirectURL)
	data.Set("client_id", config.ClientID)
	data.Set("code_verifier", codeVerifier)

	req, err := http.NewRequestWithContext(ctx, "POST", config.Endpoint.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Get HTTP client from context if available
	client := http.DefaultClient
	if ctxClient, ok := ctx.Value(oauth2.HTTPClient).(*http.Client); ok {
		client = ctxClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	token := &oauth2.Token{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		RefreshToken: tokenResp.RefreshToken,
	}

	if tokenResp.ExpiresIn > 0 {
		token.Expiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	}

	// Add id_token as extra field
	if tokenResp.IDToken != "" {
		token = token.WithExtra(map[string]interface{}{
			"id_token": tokenResp.IDToken,
		})
	}

	return token, nil
}

// openBrowser opens the default browser with the given URL
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

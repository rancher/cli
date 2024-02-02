package cmd

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rancher/cli/pkce"
	apiv3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"

	"github.com/skratchdot/open-golang/open"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/authhandler"
)

func oauthAuth(input *LoginInput, provider TypedProvider) (managementClient.Token, error) {
	token := managementClient.Token{}

	// channel where send the final URL coming from the authorization flow
	ch := make(chan authorizationResult)
	authorizationHandler, redirectURL, err := newAuthorizationHandler(ch, input.prompt)

	oauthConfig, err := newOauthConfig(provider, redirectURL)
	if err != nil {
		return token, err
	}

	state, err := initState(provider.GetType())
	if err != nil {
		return token, err
	}

	pkceParams, err := initPKCE()
	if err != nil {
		return token, err
	}

	tokenSource := authhandler.TokenSourceWithPKCE(
		context.Background(),
		oauthConfig,
		state,
		authorizationHandler,
		pkceParams,
	)

	oauthToken, err := tokenSource.Token()
	if err != nil {
		return token, err
	}

	// login with id_token
	providerName := strings.ToLower(strings.TrimSuffix(input.authProvider, "Provider"))
	url := fmt.Sprintf("%s/v3-public/%ss/%s?action=login", input.server, provider.GetType(), providerName)

	jsonBody := fmt.Sprintf(`{"responseType":"kubeconfig","token":"%s"}`, oauthToken.Extra("id_token"))
	response, err := http.Post(url, "application/json", strings.NewReader(jsonBody))
	b, err := io.ReadAll(response.Body)
	if err != nil {
		return token, err
	}

	err = json.Unmarshal(b, &token)
	return token, err
}

func newOauthConfig(provider TypedProvider, redirectURL string) (*oauth2.Config, error) {
	var oauthProvider apiv3.OAuthProvider

	switch p := provider.(type) {
	case *apiv3.AzureADProvider:
		oauthProvider = p.OAuthProvider
	default:
		return nil, errors.New("provider is not a supported OAuth provider")
	}

	return &oauth2.Config{
		ClientID:    oauthProvider.ClientID,
		Scopes:      oauthProvider.Scopes,
		RedirectURL: redirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:       oauthProvider.AuthURL,
			DeviceAuthURL: oauthProvider.DeviceAuthURL,
			TokenURL:      oauthProvider.TokenURL,
		},
	}, nil
}

func initState(provider string) (string, error) {
	nonce, err := generateKey()
	if err != nil {
		return "", err
	}

	state := struct {
		Nonce    string `json:"nonce"`
		Provider string `json:"provider"`
	}{
		Nonce:    nonce,
		Provider: provider,
	}

	b, err := json.Marshal(&state)
	if err != nil {
		return "", err
	}

	encodedState := base64.RawURLEncoding.EncodeToString(b)
	return encodedState, nil
}

func initPKCE() (*authhandler.PKCEParams, error) {
	code, err := pkce.Generate()
	if err != nil {
		return nil, err
	}

	pkceParams := &authhandler.PKCEParams{
		Verifier:        code.Verifier(),
		Challenge:       code.Challenge(),
		ChallengeMethod: code.Method(),
	}

	return pkceParams, nil
}

type authorizationResult struct {
	Code             string
	State            string
	Error            string
	ErrorDescription string
}

// newAuthorizationHandler returns an AuthorizationHandler used to perform the authorization flow.
// It will wait an URL from the channel where to grab the 'code' and 'state' parameters.
// It will start a local server if prompt is false, or it will wait the URL from the console
// if prompt is true.
func newAuthorizationHandler(ch chan authorizationResult, prompt bool) (authhandler.AuthorizationHandler, string, error) {
	var redirectURL string
	var err error

	// if prompt is true wait for the user to input the URL
	if prompt {
		redirectURL = promptUser(ch)
	} else { // else start a local server that will intercept the URL
		redirectURL, err = startServer(ch)
		if err != nil {
			return nil, "", err
		}
	}

	return func(authCodeURL string) (string, string, error) {
		customPrint("\n" + authCodeURL)

		if prompt {
			customPrint("\nOpen the URL in your browser and follow the directions.")
			customPrint("\nEnter Code:")
		} else {
			customPrint("\nOpen this URL in your browser and follow the directions.")
			// if it fails to open the browser the user can still proceed manually
			_ = open.Run(authCodeURL)
		}

		// wait for the code
		authResult := <-ch

		// handle errors
		if authResult.Error != "" || authResult.ErrorDescription != "" {
			return "", "", fmt.Errorf("%s: %s", authResult.Error, authResult.ErrorDescription)
		}

		if authResult.Code == "" {
			return "", "", errors.New("code not found")
		}

		if authResult.State == "" {
			return "", "", errors.New("state not found")
		}

		return authResult.Code, authResult.State, nil
	}, redirectURL, nil
}

func promptUser(ch chan authorizationResult) string {
	go func() {
		code, err := readUserInput()
		if err != nil {
			ch <- authorizationResult{
				Error:            err.Error(),
				ErrorDescription: "error reading code",
			}
			return
		}

		customPrint("\nEnter State:")
		state, err := readUserInput()
		if err != nil {
			ch <- authorizationResult{
				Error:            err.Error(),
				ErrorDescription: "error reading state",
			}
			return
		}

		ch <- authorizationResult{
			Code:  code,
			State: state,
		}
	}()

	return "https://127.0.0.1:8005/oauth-verify-static"
}

func readUserInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	s, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(s), nil
}

func startServer(ch chan authorizationResult) (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", errors.Wrap(err, "creating listener")
	}
	localRedirectURL := fmt.Sprintf("http://127.0.0.1:%d", listener.Addr().(*net.TCPAddr).Port)

	srv := &http.Server{ReadHeaderTimeout: time.Second * 30}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Login successful! You can close this window.")

		queryParams := r.URL.Query()

		ch <- authorizationResult{
			Code:             queryParams.Get("code"),
			State:            queryParams.Get("state"),
			Error:            queryParams.Get("error"),
			ErrorDescription: queryParams.Get("error_description"),
		}
	})

	go func() { _ = srv.Serve(listener) }()

	return localRedirectURL, nil
}

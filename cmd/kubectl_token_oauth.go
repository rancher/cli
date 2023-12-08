package cmd

import (
	// "context"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/rancher/cli/pkce"
	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"golang.org/x/oauth2"
)

type OAuthLogin struct {
	TTLMillis    int64  `json:"ttl,omitempty"`
	Description  string `json:"description,omitempty"`
	ResponseType string `json:"responseType,omitempty"`
	AccessToken  string `json:"access_token" norman:"type=string"`
}

type OAuthProviders struct {
	Data []OAuthProvider `json:"data"`
}

// All of these are copied from custom v3public
type OAuthProvider struct {
	Actions struct {
		Login string `json:"login"`
	} `json:"actions"`

	Type string `json:"type"`

	Scopes    []string      `json:"scopes"`
	Endpoints OAuthEndpoint `json:"endpoints"`

	// AuthClientInfo is the info required for the Authorization Code grant. It
	// is optional because not every provider supports it.
	AuthClientInfo *OAuthAuthorizationInfo `json:"authClientInfo,omitempty"`

	// DeviceClientInfo is the info required for the Device Code grant.
	// It is optional because not every provider supports it.
	DeviceClientInfo *OAuthDeviceInfo `json:"deviceClientInfo,omitempty"`
}

type OAuthEndpoint struct {
	AuthURL       string `json:"authUrl"`
	DeviceAuthURL string `json:"deviceAuthUrl"`
	TokenURL      string `json:"tokenUrl"`
}

type OAuthAuthorizationInfo struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	RedirectURL  string `json:"redirectUrl"`
}

type OAuthDeviceInfo struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

func runAuthCodeFlow(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	codeCh := make(chan string, 1)
	go func() {
		http.ListenAndServe("127.0.0.1:53000", http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
			codeCh <- req.URL.Query().Get("code")
		}))
	}()

	pkceCode, err := pkce.Generate()
	if err != nil {
		return nil, err
	}

	url := config.AuthCodeURL(
		// TODO: Randomize this
		"mystate",
		pkceCode.Challenge(),
		pkceCode.Method(),
	)
	// TODO: Open browser
	fmt.Fprintln(os.Stderr, "My url is ", url)

	code := <-codeCh

	token, err := config.Exchange(ctx, code, pkceCode.Verifier())
	if err != nil {
		return nil, err
	}
	return token, nil
}

func runDeviceFlow(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	response, err := config.DeviceAuth(ctx)
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(os.Stderr, "Please enter code %s at %s\n", response.UserCode, response.VerificationURI)
	token, err := config.DeviceAccessToken(ctx, response)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func getAuthProviderByType(input *LoginInput) (*OAuthProvider, error) {
	url := fmt.Sprintf(authProviderURL, input.server)
	response, err := request(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var providers OAuthProviders
	if err := json.Unmarshal(response, &providers); err != nil {
		return nil, err
	}

	for _, provider := range providers.Data {
		if provider.Type != input.authProvider {
			continue
		}

		return &provider, nil
	}
	return nil, fmt.Errorf("Not found")
}

func authProviderToOAuthConfig(provider *OAuthProvider, useDeviceFlow bool) (*oauth2.Config, error) {
	if useDeviceFlow {
		if provider.DeviceClientInfo == nil {
			return nil, fmt.Errorf("device code flow not supported")
		}

		info := provider.DeviceClientInfo
		return &oauth2.Config{
			ClientID:     info.ClientID,
			ClientSecret: info.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:       provider.Endpoints.AuthURL,
				TokenURL:      provider.Endpoints.TokenURL,
				DeviceAuthURL: provider.Endpoints.DeviceAuthURL,
			},
			Scopes: provider.Scopes,
		}, nil
	} else {
		if provider.AuthClientInfo == nil {
			return nil, fmt.Errorf("authorization code flow not supported")
		}

		info := provider.AuthClientInfo
		return &oauth2.Config{
			ClientID:     info.ClientID,
			ClientSecret: info.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  provider.Endpoints.AuthURL,
				TokenURL: provider.Endpoints.TokenURL,
			},
			Scopes:      provider.Scopes,
			RedirectURL: info.RedirectURL,
		}, nil
	}
}

func oauthAuth(input *LoginInput, tlsConfig *tls.Config) (managementClient.Token, error) {
	ctx := context.Background()

	provider, err := getAuthProviderByType(input)
	if err != nil {
		return managementClient.Token{}, err
	}

	config, err := authProviderToOAuthConfig(provider, input.useDeviceFlow)
	if err != nil {
		return managementClient.Token{}, err
	}

	var token *oauth2.Token
	if input.useDeviceFlow {
		token, err = runDeviceFlow(ctx, config)
	} else {
		token, err = runAuthCodeFlow(ctx, config)
	}
	if err != nil {
		return managementClient.Token{}, err
	}

	loginURL := provider.Actions.Login
	params := OAuthLogin{
		ResponseType: "kubeconfig",
		AccessToken:  token.AccessToken,
	}
	payload, err := json.Marshal(params)
	if err != nil {
		return managementClient.Token{}, err
	}

	resp, err := http.Post(loginURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return managementClient.Token{}, err
	}

	defer resp.Body.Close()

	rancherToken := managementClient.Token{}
	payload, _ = io.ReadAll(resp.Body)
	_ = json.Unmarshal(payload, &rancherToken)
	return rancherToken, err
}

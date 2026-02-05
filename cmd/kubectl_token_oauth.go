package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	apiv3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"golang.org/x/oauth2"
)

func oauthAuth(client *http.Client, input *LoginInput, provider TypedProvider, useV1Public bool) (*managementClient.Token, error) {
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

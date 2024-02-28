package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	apiv3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"

	"golang.org/x/oauth2"
)

func oauthAuth(input *LoginInput, provider TypedProvider) (*managementClient.Token, error) {
	oauthConfig, err := newOauthConfig(provider)
	if err != nil {
		return nil, err
	}

	deviceAuthResp, err := oauthConfig.DeviceAuth(context.Background())
	if err != nil {
		return nil, err
	}

	customPrint(fmt.Sprintf(
		"\nTo sign in, use a web browser to open the page %s and enter the code %s to authenticate.\n",
		deviceAuthResp.VerificationURI,
		deviceAuthResp.UserCode,
	))

	oauthToken, err := oauthConfig.DeviceAccessToken(context.Background(), deviceAuthResp)
	if err != nil {
		return nil, err
	}

	return rancherLogin(input, provider, oauthToken)
}

func newOauthConfig(provider TypedProvider) (*oauth2.Config, error) {
	var oauthProvider apiv3.OAuthProvider

	switch p := provider.(type) {
	case *apiv3.AzureADProvider:
		oauthProvider = p.OAuthProvider
	default:
		return nil, errors.New("provider is not a supported OAuth provider")
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

func rancherLogin(input *LoginInput, provider TypedProvider, oauthToken *oauth2.Token) (*managementClient.Token, error) {
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

	response, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	b, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	token := &managementClient.Token{}
	err = json.Unmarshal(b, token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

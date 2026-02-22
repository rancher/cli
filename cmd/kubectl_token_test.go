package cmd

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rancher/cli/config"
	apiv3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

func TestGetAuthProviders(t *testing.T) {
	t.Parallel()

	client := &http.Client{Timeout: time.Second}

	expectedProviders := []TypedProvider{
		&apiv3.LocalProvider{
			AuthProvider: apiv3.AuthProvider{
				Type: "localProvider",
			},
		},
		&apiv3.AzureADProvider{
			AuthProvider: apiv3.AuthProvider{
				Type: "azureADProvider",
			},
			RedirectURL: "https://login.microsoftonline.com/258928db-3ed6-49fb-9a7e-52e492ffb066/oauth2/v2.0/authorize?client_id=56168f69-a732-48e2-aa21-8aa0909d0976&redirect_uri=https://rancher.mydomain.com/verify-auth-azure&response_type=code&scope=openid",
			TenantID:    "258928db-3ed6-49fb-9a7e-52e492ffb066",
			OAuthProvider: apiv3.OAuthProvider{
				ClientID: "56168f69-a732-48e2-aa21-8aa0909d0976",
				Scopes:   []string{"openid", "profile", "email"},
				OAuthEndpoint: apiv3.OAuthEndpoint{
					AuthURL:       "https://login.microsoftonline.com/258928db-3ed6-49fb-9a7e-52e492ffb066/oauth2/v2.0/authorize",
					DeviceAuthURL: "https://login.microsoftonline.com/258928db-3ed6-49fb-9a7e-52e492ffb066/oauth2/v2.0/devicecode",
					TokenURL:      "https://login.microsoftonline.com/258928db-3ed6-49fb-9a7e-52e492ffb066/oauth2/v2.0/token",
				},
			},
		},
	}

	t.Run("successful response with v1-public endpoints", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Contains(t, r.URL.Path, "/v1-public/authproviders")
			fmt.Fprint(w, authProvidersResponse)
		}))
		t.Cleanup(server.Close)

		providers, useV1Public, err := getAuthProviders(client, server.URL, true)

		require.NoError(t, err)
		assert.True(t, useV1Public)
		assert.Equal(t, expectedProviders, providers)
	})

	t.Run("successful response with v3-public endpoints", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Contains(t, r.URL.Path, "/v3-public/authProviders")
			fmt.Fprint(w, authProvidersResponseV3)
		}))
		t.Cleanup(server.Close)

		providers, useV1Public, err := getAuthProviders(client, server.URL, false)

		require.NoError(t, err)
		assert.False(t, useV1Public)
		assert.Equal(t, expectedProviders, providers)
	})

	t.Run("fallback from v1-public to v3-public on 404", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				// First call to v1-public should return 404.
				assert.Contains(t, r.URL.Path, "/v1-public/authproviders")
				w.WriteHeader(http.StatusNotFound)
			} else {
				// Second call to v3-public should succeed.
				assert.Contains(t, r.URL.Path, "/v3-public/authProviders")
				fmt.Fprint(w, authProvidersResponseV3)
			}
		}))
		t.Cleanup(server.Close)

		providers, useV1Public, err := getAuthProviders(client, server.URL, true)

		require.NoError(t, err)
		assert.False(t, useV1Public, "should have fallen back to v3-public")
		assert.Equal(t, expectedProviders, providers)
		assert.Equal(t, 2, callCount, "should have made exactly 2 requests")
	})

	t.Run("404 on both v1-public and v3-public endpoints", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.WriteHeader(http.StatusNotFound)
		}))
		t.Cleanup(server.Close)

		providers, useV1Public, err := getAuthProviders(client, server.URL, true)

		require.Error(t, err)
		assert.ErrorContains(t, err, "error listing auth providers")
		assert.False(t, useV1Public)
		assert.Nil(t, providers)
		assert.Equal(t, 2, callCount, "should have tried both endpoints")
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `this is not valid json`)
		}))
		t.Cleanup(server.Close)

		providers, useV1Public, err := getAuthProviders(client, server.URL, true)

		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid JSON response")
		assert.False(t, useV1Public)
		assert.Nil(t, providers)
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		t.Cleanup(server.Close)

		providers, useV1Public, err := getAuthProviders(client, server.URL, true)

		require.Error(t, err)
		assert.ErrorContains(t, err, "500")
		assert.False(t, useV1Public)
		assert.Nil(t, providers)
	})

	t.Run("filters unsupported providers", func(t *testing.T) {
		responseWithUnsupported := `{
			"data": [
				{
					"type": "localProvider",
					"id": "local"
				},
				{
					"type": "unsupportedProvider",
					"id": "unsupported"
				},
				{
					"type": "githubProvider",
					"id": "github"
				}
			]
		}`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, responseWithUnsupported)
		}))
		t.Cleanup(server.Close)

		providers, useV1Public, err := getAuthProviders(client, server.URL, true)

		require.NoError(t, err)
		assert.True(t, useV1Public)
		require.Len(t, providers, 1, "should only return supported providers")
		assert.Equal(t, "localProvider", providers[0].GetType())
	})
}

var authProvidersResponse = `{
    "data": [
        {
            "authUrl": "https://login.microsoftonline.com/258928db-3ed6-49fb-9a7e-52e492ffb066/oauth2/v2.0/authorize",
            "clientId": "56168f69-a732-48e2-aa21-8aa0909d0976",
            "deviceAuthUrl": "https://login.microsoftonline.com/258928db-3ed6-49fb-9a7e-52e492ffb066/oauth2/v2.0/devicecode",
            "id": "azuread",
            "redirectUrl": "https://login.microsoftonline.com/258928db-3ed6-49fb-9a7e-52e492ffb066/oauth2/v2.0/authorize?client_id=56168f69-a732-48e2-aa21-8aa0909d0976&redirect_uri=https://rancher.mydomain.com/verify-auth-azure&response_type=code&scope=openid",
            "scopes": [
                "openid",
                "profile",
                "email"
            ],
            "tenantId": "258928db-3ed6-49fb-9a7e-52e492ffb066",
            "tokenUrl": "https://login.microsoftonline.com/258928db-3ed6-49fb-9a7e-52e492ffb066/oauth2/v2.0/token",
            "type": "azureADProvider"
        },
        {
            "id": "local",
            "type": "localProvider"
        }
    ]
}`

var authProvidersResponseV3 = `{
    "data": [
        {
            "actions": {
                "login": "…/v3-public/azureADProviders/azuread?action=login"
            },
            "authUrl": "https://login.microsoftonline.com/258928db-3ed6-49fb-9a7e-52e492ffb066/oauth2/v2.0/authorize",
            "baseType": "authProvider",
            "clientId": "56168f69-a732-48e2-aa21-8aa0909d0976",
            "creatorId": null,
            "deviceAuthUrl": "https://login.microsoftonline.com/258928db-3ed6-49fb-9a7e-52e492ffb066/oauth2/v2.0/devicecode",
            "id": "azuread",
            "links": {
                "self": "…/v3-public/azureADProviders/azuread"
            },
            "redirectUrl": "https://login.microsoftonline.com/258928db-3ed6-49fb-9a7e-52e492ffb066/oauth2/v2.0/authorize?client_id=56168f69-a732-48e2-aa21-8aa0909d0976&redirect_uri=https://rancher.mydomain.com/verify-auth-azure&response_type=code&scope=openid",
            "scopes": [
                "openid",
                "profile",
                "email"
            ],
            "tenantId": "258928db-3ed6-49fb-9a7e-52e492ffb066",
            "tokenUrl": "https://login.microsoftonline.com/258928db-3ed6-49fb-9a7e-52e492ffb066/oauth2/v2.0/token",
            "type": "azureADProvider"
        },
        {
            "actions": {
                "login": "…/v3-public/localProviders/local?action=login"
            },
            "baseType": "authProvider",
            "creatorId": null,
            "id": "local",
            "links": {
                "self": "…/v3-public/localProviders/local"
            },
            "type": "localProvider"
        }
    ]
}`

func TestCacheCredential(t *testing.T) {
	flagSet := flag.NewFlagSet("test", 0)
	flagSet.String("server", "rancher.example.com", "doc")
	flagSet.String("config", t.TempDir(), "doc")
	flagSet.String("config-helper", "built-in", "")
	cliCtx := cli.NewContext(nil, flagSet, nil)

	serverConfig, err := lookupServerConfig(cliCtx)
	if err != nil {
		t.Fatal(err)
	}

	cred := &config.ExecCredential{Status: &config.ExecCredentialStatus{Token: "test-token"}}

	err = cacheCredential(cliCtx, serverConfig, "dev-server", cred)
	require.NoError(t, err)

	cfg, err := loadConfig(cliCtx)
	require.NoError(t, err)

	expires := &config.Time{Time: time.Now().Add(time.Hour * 2)}
	cfg.CurrentServer = "rancher.example.com"
	cfg.Servers["rancher.example.com"].KubeCredentials["dev-server"].Status.ClientKeyData = "this-is-not-real"
	cfg.Servers["rancher.example.com"].KubeCredentials["dev-server"].Status.ExpirationTimestamp = expires

	err = cfg.Write()
	require.NoError(t, err)

	serverConfig, err = cfg.FocusedServer()
	require.NoError(t, err)

	cred = &config.ExecCredential{Status: &config.ExecCredentialStatus{Token: "new-token"}}

	err = cacheCredential(cliCtx, serverConfig, "local", cred)
	require.NoError(t, err)

	cfg, err = loadConfig(cliCtx)
	require.NoError(t, err)

	clientKeyData := cfg.Servers["rancher.example.com"].KubeCredentials["dev-server"].Status.ClientKeyData
	assert.Equal(t, "this-is-not-real", clientKeyData)

	expirationTimestamp := cfg.Servers["rancher.example.com"].KubeCredentials["dev-server"].Status.ExpirationTimestamp
	require.NotNil(t, expirationTimestamp)
	assert.True(t, expirationTimestamp.Equal(expires.Time))
}

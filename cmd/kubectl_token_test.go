package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	apiv3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/stretchr/testify/assert"
)

func Test_getAuthProviders(t *testing.T) {

	setupServer := func(response string) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, response)
		}))
	}

	tt := []struct {
		name              string
		server            *httptest.Server
		expectedProviders []TypedProvider
		expectedErr       string
	}{
		{
			name:   "response ok",
			server: setupServer(responseOK),
			expectedProviders: []TypedProvider{
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
				&apiv3.LocalProvider{
					AuthProvider: apiv3.AuthProvider{
						Type: "localProvider",
					},
				},
			},
		},
		{
			name:        "json error",
			server:      setupServer(`hnjskjnksnj`),
			expectedErr: "invalid JSON input",
		},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := getAuthProviders(tc.server.URL)

			if tc.expectedErr != "" {
				assert.EqualError(t, err, tc.expectedErr)
				assert.Nil(t, got)
			} else {
				assert.Equal(t, tc.expectedProviders, got)
				assert.Nil(t, err)
			}
		})
	}
}

var responseOK = `{
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

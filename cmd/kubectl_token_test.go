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
	"github.com/urfave/cli"
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
			},
		},
		{
			name:        "json error",
			server:      setupServer(`hnjskjnksnj`),
			expectedErr: "invalid JSON response from",
		},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(tc.server.Close)

			got, err := getAuthProviders(tc.server.URL)

			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
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

func Test_cacheCredential(t *testing.T) {
	tempDir := t.TempDir()

	cred := &config.ExecCredential{Status: &config.ExecCredentialStatus{Token: "test-token"}}
	flagSet := flag.NewFlagSet("test", 0)
	flagSet.String("server", "rancher.example.com", "doc")
	flagSet.String("config", tempDir, "doc")
	cliCtx := cli.NewContext(nil, flagSet, nil)

	err := cacheCredential(cliCtx, cred, "dev-server")
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := loadConfig(cliCtx)
	if err != nil {
		t.Fatal(err)
	}
	expires := &config.Time{Time: time.Now().Add(time.Hour * 2)}
	cfg.CurrentServer = "rancher.example.com"
	cfg.Servers["rancher.example.com"].KubeCredentials["dev-server"].Status.ClientKeyData = "this-is-not-real"
	cfg.Servers["rancher.example.com"].KubeCredentials["dev-server"].Status.ExpirationTimestamp = expires
	if err := cfg.Write(); err != nil {
		t.Fatal(err)
	}

	_, err = cfg.FocusedServer()
	if err != nil {
		t.Fatal(err)
	}

	flagSet = flag.NewFlagSet("test", 0)
	flagSet.String("server", "rancher.example.com", "doc")
	flagSet.String("config", tempDir, "doc")
	cliCtx = cli.NewContext(nil, flagSet, nil)

	cred = &config.ExecCredential{Status: &config.ExecCredentialStatus{Token: "new-token"}}
	err = cacheCredential(cliCtx, cred, "local")
	if err != nil {
		t.Fatal(err)
	}

	cfg, err = loadConfig(cliCtx)
	if err != nil {
		t.Fatal(err)
	}

	if v := cfg.Servers["rancher.example.com"].KubeCredentials["dev-server"].Status.ClientKeyData; v != "this-is-not-real" {
		t.Errorf("got ClientKeyData %q, want \"this-is-not-real\"", v)
	}
	if v := cfg.Servers["rancher.example.com"].KubeCredentials["dev-server"].Status.ExpirationTimestamp; !v.Time.Equal(expires.Time) {
		t.Errorf("got ExpirationTimestamp %v, want %v", v, expires)
	}

}

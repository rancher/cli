package cmd

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rancher/cli/cliclient"
	"github.com/rancher/cli/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// fakeRancher is a Rancher server with just enough of the public auth API and
// the /v3 API for "rancher login" to sign in and pick a project.
type fakeRancher struct {
	t   *testing.T
	url string

	mu        sync.Mutex
	providers string
	// loginToken is the token the login endpoints hand out.
	loginToken string
	expiresAt  string
	logins     []loginRequest
	// samlToken is the encrypted SAML token, set once the user "signed in".
	samlToken string
	// v3Auth lists the Authorization headers of the /v3 requests.
	v3Auth []string
}

func newFakeRancher(t *testing.T) *fakeRancher {
	t.Helper()

	f := &fakeRancher{
		t:          t,
		providers:  singleProvider,
		loginToken: "token-abc:secret",
		expiresAt:  time.Now().Add(16 * time.Hour).UTC().Format(time.RFC3339),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()

		if strings.HasPrefix(r.URL.Path, "/v3") {
			f.v3Auth = append(f.v3Auth, r.Header.Get("Authorization"))
		}

		switch {
		case r.URL.Path == "/v1-public/authproviders":
			fmt.Fprint(w, f.providers)
		case r.URL.Path == "/v1-public/login":
			var req loginRequest
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			f.logins = append(f.logins, req)
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `{"type": "token", "token": %q, "expiresAt": %q}`, f.loginToken, f.expiresAt)
		case strings.HasPrefix(r.URL.Path, "/v1-public/authtokens/"):
			switch {
			case r.Method == http.MethodDelete:
				w.WriteHeader(http.StatusNoContent)
			case f.samlToken == "":
				w.WriteHeader(http.StatusNotFound)
			default:
				fmt.Fprintf(w, `{"token": %q, "expiresAt": %q}`, f.samlToken, f.expiresAt)
			}
		case r.URL.Path == "/v3":
			w.Header().Set("X-API-Schemas", f.url+"/v3/schemas")
			fmt.Fprint(w, `{}`)
		case r.URL.Path == "/v3/schemas":
			fmt.Fprintf(w, `{"data": [{"id": "project", "type": "schema",
				"collectionMethods": ["GET"], "links": {"collection": %q}}]}`, f.url+"/v3/projects")
		case r.URL.Path == "/v3/projects":
			fmt.Fprint(w, `{"data": [{"id": "local:p-abcde", "name": "Default", "clusterId": "local"}]}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	f.url = server.URL

	return f
}

// signInWithSAML plays the user signing in through the SAML login page at
// loginURL: it encrypts the login token with the CLI's public key, as Rancher
// does.
func (f *fakeRancher) signInWithSAML(loginURL string) error {
	u, err := url.Parse(loginURL)
	if err != nil {
		return err
	}
	keyJSON, err := base64.StdEncoding.DecodeString(u.Query().Get("publicKey"))
	if err != nil {
		return err
	}
	var pub rsa.PublicKey
	if err := json.Unmarshal(keyJSON, &pub); err != nil {
		return err
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &pub, []byte(f.loginToken), nil)
	if err != nil {
		return err
	}
	f.samlToken = base64.StdEncoding.EncodeToString(encrypted)
	return nil
}

// scriptPrompts answers the username and password prompts.
func scriptPrompts(t *testing.T) {
	t.Helper()
	old := promptFunc
	promptFunc = func(msg string, show bool) (string, error) {
		if show {
			return "alice", nil
		}
		return "password", nil
	}
	t.Cleanup(func() { promptFunc = old })
}

// runLogin runs "rancher login" with a config in configDir.
func runLogin(t *testing.T, configDir string, args ...string) error {
	t.Helper()
	root := &cli.Command{
		Name:     "rancher",
		Flags:    []cli.Flag{&cli.StringFlag{Name: "config", Value: configDir, Hidden: true}},
		Commands: []*cli.Command{LoginCommand()},
	}
	return root.Run(t.Context(), append([]string{"rancher", "login"}, args...))
}

func loadServerConfig(t *testing.T, configDir, name string) *config.ServerConfig {
	t.Helper()
	cf, err := config.LoadFromPath(filepath.Join(configDir, cfgFile))
	require.NoError(t, err)
	return cf.Servers[name]
}

func TestLoginAndGetToken(t *testing.T) {
	tests := []struct {
		name             string
		responseType     string
		clusterID        string
		wantResponseType string
	}{
		{name: "default is a kubeconfig token", wantResponseType: "kubeconfig"},
		{name: "default for a cluster", clusterID: "c-1234", wantResponseType: "kubeconfig_c-1234"},
		{name: "session token", responseType: sessionResponseType, wantResponseType: "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFakeRancher(t)
			scriptPrompts(t)

			input := &LoginInput{server: f.url, clusterID: tt.clusterID, responseType: tt.responseType}
			// A client without a proxy: http.ProxyFromEnvironment caches the
			// environment, which TestNewHTTPClient changes.
			token, err := loginAndGetToken(&http.Client{Transport: &http.Transport{}}, input)
			require.NoError(t, err)

			assert.Equal(t, "token-abc:secret", token.BearerToken)
			assert.Equal(t, f.expiresAt, token.ExpiresAt)
			require.Len(t, f.logins, 1)
			assert.Equal(t, tt.wantResponseType, f.logins[0].ResponseType)
			assert.Equal(t, "localProvider", f.logins[0].Type)
			assert.Equal(t, "alice", f.logins[0].Username)
		})
	}
}

func TestLoginWithAuthProvider(t *testing.T) {
	f := newFakeRancher(t)
	scriptPrompts(t)
	configDir := t.TempDir()

	err := runLogin(t, configDir, "--auth-provider", "local", f.url)
	require.NoError(t, err)

	sc := loadServerConfig(t, configDir, "rancherDefault")
	require.NotNil(t, sc)
	assert.Equal(t, "token-abc", sc.AccessKey)
	assert.Equal(t, "secret", sc.SecretKey)
	assert.Equal(t, "token-abc:secret", sc.TokenKey)
	assert.Equal(t, f.expiresAt, sc.TokenExpiresAt)
	assert.Equal(t, "local:p-abcde", sc.Project)

	require.Len(t, f.logins, 1)
	assert.Equal(t, sessionResponseType, f.logins[0].ResponseType)

	// The /v3 API was called with the new token.
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("token-abc:secret"))
	require.NotEmpty(t, f.v3Auth)
	for _, got := range f.v3Auth {
		assert.Equal(t, wantAuth, got)
	}
}

func TestLoginWithSAML(t *testing.T) {
	f := newFakeRancher(t)
	f.providers = `{"data": [{"id": "keycloak", "type": "keyCloakProvider"}]}`
	f.loginToken = "kubeconfig-u-xyz:samlsecret"
	configDir := t.TempDir()

	oldPoll := samlPollInterval
	samlPollInterval = 10 * time.Millisecond
	t.Cleanup(func() { samlPollInterval = oldPoll })

	var opened []string
	oldOpen := loginOpenBrowser
	loginOpenBrowser = func(u string) error {
		opened = append(opened, u)
		return f.signInWithSAML(u)
	}
	t.Cleanup(func() { loginOpenBrowser = oldOpen })

	err := runLogin(t, configDir, "--auth-provider", "keyCloakProvider", f.url)
	require.NoError(t, err)

	require.Len(t, opened, 1)
	assert.True(t, strings.HasPrefix(opened[0], f.url+"/dashboard/auth/login?"), opened[0])

	sc := loadServerConfig(t, configDir, "rancherDefault")
	require.NotNil(t, sc)
	assert.Equal(t, "kubeconfig-u-xyz", sc.AccessKey)
	assert.Equal(t, "samlsecret", sc.SecretKey)
	assert.Equal(t, f.expiresAt, sc.TokenExpiresAt)
}

func TestLoginSAMLBrowserFailure(t *testing.T) {
	f := newFakeRancher(t)
	f.providers = `{"data": [{"id": "keycloak", "type": "keyCloakProvider"}]}`
	configDir := t.TempDir()

	oldPoll := samlPollInterval
	samlPollInterval = 10 * time.Millisecond
	t.Cleanup(func() { samlPollInterval = oldPoll })

	// The browser does not open, but the user opens the printed URL.
	oldOpen := loginOpenBrowser
	loginOpenBrowser = func(u string) error {
		require.NoError(t, f.signInWithSAML(u))
		return fmt.Errorf("no browser")
	}
	t.Cleanup(func() { loginOpenBrowser = oldOpen })

	require.NoError(t, runLogin(t, configDir, "--auth-provider", "keycloak", f.url))
	assert.Equal(t, "token-abc", loadServerConfig(t, configDir, "rancherDefault").AccessKey)
}

func TestLoginBadToken(t *testing.T) {
	for _, token := range []string{"no-colon", "a:b:c", ""} {
		t.Run(fmt.Sprintf("%q", token), func(t *testing.T) {
			f := newFakeRancher(t)
			f.loginToken = token
			scriptPrompts(t)
			configDir := t.TempDir()

			err := runLogin(t, configDir, "--auth-provider", "local", f.url)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "login failed")
			assert.Nil(t, loadServerConfig(t, configDir, "rancherDefault"), "nothing is saved")
		})
	}
}

func TestLoginTokenFlagConflicts(t *testing.T) {
	err := runLogin(t, t.TempDir(), "--token", "a:b", "--auth-provider", "local", "https://rancher.example.com")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be combined")
}

func TestLoginWithTokenClearsExpiry(t *testing.T) {
	f := newFakeRancher(t)
	configDir := t.TempDir()

	cf, err := config.LoadFromPath(filepath.Join(configDir, cfgFile))
	require.NoError(t, err)
	cf.Servers["rancherDefault"] = &config.ServerConfig{URL: f.url, TokenExpiresAt: "2000-01-01T00:00:00Z"}
	require.NoError(t, cf.Write())

	require.NoError(t, runLogin(t, configDir, "--token", "token-abc:secret", f.url))

	sc := loadServerConfig(t, configDir, "rancherDefault")
	assert.Equal(t, "token-abc", sc.AccessKey)
	assert.Empty(t, sc.TokenExpiresAt)
}

func TestClientRejectsExpiredToken(t *testing.T) {
	sc := &config.ServerConfig{
		URL:            "https://rancher.example.com",
		AccessKey:      "token-abc",
		SecretKey:      "secret",
		Project:        "local:p-abcde",
		TokenExpiresAt: time.Now().Add(-time.Minute).UTC().Format(time.RFC3339),
	}

	for name, newClient := range map[string]func(*config.ServerConfig) (*cliclient.MasterClient, error){
		"master":     cliclient.NewMasterClient,
		"management": cliclient.NewManagementClient,
		"cluster":    cliclient.NewClusterClient,
		"project":    cliclient.NewProjectClient,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := newClient(sc)
			require.ErrorIs(t, err, config.ErrTokenExpired)
			assert.Contains(t, err.Error(), "rancher login https://rancher.example.com")
		})
	}
}

func TestLoginSAMLSkipBrowser(t *testing.T) {
	for _, flag := range []string{"--skip-browser"} {
		t.Run(flag, func(t *testing.T) {
			f := newFakeRancher(t)
			f.providers = `{"data": [{"id": "keycloak", "type": "keyCloakProvider"}]}`
			configDir := t.TempDir()

			oldPoll := samlPollInterval
			samlPollInterval = 10 * time.Millisecond
			t.Cleanup(func() { samlPollInterval = oldPoll })

			oldOpen := loginOpenBrowser
			loginOpenBrowser = func(u string) error {
				t.Errorf("browser opened with --skip-browser: %s", u)
				return nil
			}
			t.Cleanup(func() { loginOpenBrowser = oldOpen })

			// Play the user: read the printed login URL and sign in there.
			r, w, err := os.Pipe()
			require.NoError(t, err)
			oldStderr := os.Stderr
			os.Stderr = w
			t.Cleanup(func() { os.Stderr = oldStderr })

			urlRE := regexp.MustCompile(`Login to Rancher Server at (\S+)`)
			signedIn := make(chan error, 1)
			go func() {
				scanner := bufio.NewScanner(r)
				for scanner.Scan() {
					if m := urlRE.FindStringSubmatch(scanner.Text()); m != nil {
						signedIn <- f.signInWithSAML(m[1])
						break
					}
				}
				_, _ = io.Copy(io.Discard, r)
			}()

			err = runLogin(t, configDir, flag, "--auth-provider", "keycloak", f.url)
			os.Stderr = oldStderr
			require.NoError(t, w.Close())
			require.NoError(t, err)
			require.NoError(t, <-signedIn)

			assert.Equal(t, "token-abc", loadServerConfig(t, configDir, "rancherDefault").AccessKey)
		})
	}
}

func TestLoginSkipBrowserWithToken(t *testing.T) {
	err := runLogin(t, t.TempDir(), "--token", "a:b", "--skip-browser", "https://rancher.example.com")
	require.ErrorContains(t, err, "cannot be combined")
}

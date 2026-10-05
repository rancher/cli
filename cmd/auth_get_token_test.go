package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rancher/cli/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

const (
	singleProvider = `{"data": [{"id": "local", "type": "localProvider"}]}`

	// manyProviders lists providers whose name differs from their type.
	manyProviders = `{"data": [
		{"id": "local", "type": "localProvider"},
		{"id": "openldap", "type": "openLdapProvider"},
		{"id": "freeipa", "type": "freeIpaProvider"}
	]}`

	// unnamedProviders lists providers that carry no name.
	unnamedProviders = `{"data": [
		{"type": "localProvider"},
		{"type": "openLdapProvider"}
	]}`

	// collidingProviders has a provider whose name is another provider's type.
	collidingProviders = `{"data": [
		{"id": "openLdapProvider", "type": "freeIpaProvider"},
		{"id": "openldap", "type": "openLdapProvider"}
	]}`
)

type loginRequest struct {
	Type         string `json:"type"`
	ResponseType string `json:"responseType"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

// credentialHarness runs the token and auth commands against a fake Rancher
// server with a scripted user. Tests that use it must not run in parallel:
// they replace os.Stdout and promptFunc.
type credentialHarness struct {
	t         *testing.T
	configDir string
	serverURL string

	mu             sync.Mutex
	providers      string
	logins         []loginRequest
	prompts        []string
	username       string
	providerChoice string
}

// newCredentialHarness returns a harness whose server lists one provider and
// is also the proxy in the server config. The commands read the proxy from the
// server config when it is set, so no request calls http.ProxyFromEnvironment,
// which reads the environment once per process and would hide the variables
// TestNewHTTPClient sets.
func newCredentialHarness(t *testing.T) *credentialHarness {
	t.Helper()

	h := &credentialHarness{
		t:         t,
		configDir: t.TempDir(),
		providers: singleProvider,
		username:  "alice",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1-public/authproviders":
			h.mu.Lock()
			providers := h.providers
			h.mu.Unlock()
			fmt.Fprint(w, providers)
		case "/v1-public/login":
			var req loginRequest
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			h.mu.Lock()
			h.logins = append(h.logins, req)
			n := len(h.logins)
			h.mu.Unlock()
			expires := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `{"type": "token", "token": "token-%d:secret", "expiresAt": %q}`, n, expires)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	h.serverURL = server.URL

	cf, err := config.LoadFromPath(filepath.Join(h.configDir, cfgFile))
	require.NoError(t, err)
	cf.Servers[h.serverURL] = &config.ServerConfig{
		ProxyURL:        h.serverURL,
		KubeCredentials: make(map[string]*config.ExecCredential),
	}
	require.NoError(t, cf.Write())

	oldPrompt := promptFunc
	promptFunc = func(msg string, show bool) (string, error) {
		h.prompts = append(h.prompts, msg)
		switch {
		case msg == "Select auth provider: ":
			return h.providerChoice, nil
		case show:
			return h.username, nil
		default:
			return "password", nil
		}
	}
	t.Cleanup(func() { promptFunc = oldPrompt })

	return h
}

// runRoot runs the command line under a root command that has the config flag
// and returns what the command wrote to stdout.
func runRoot(t *testing.T, configDir string, args ...string) (string, error) {
	t.Helper()

	root := &cli.Command{
		Name:  "rancher",
		Flags: []cli.Flag{&cli.StringFlag{Name: "config", Value: configDir, Hidden: true}},
		Commands: []*cli.Command{
			AuthCommand(),
			CredentialCommand(),
		},
	}

	r, w, err := os.Pipe()
	require.NoError(t, err)
	oldStdout := os.Stdout
	os.Stdout = w
	runErr := root.Run(t.Context(), append([]string{"rancher"}, args...))
	os.Stdout = oldStdout
	require.NoError(t, w.Close())
	out, err := io.ReadAll(r)
	require.NoError(t, err)
	require.NoError(t, r.Close())

	return string(out), runErr
}

func (h *credentialHarness) run(args ...string) (string, error) {
	h.t.Helper()
	return runRoot(h.t, h.configDir, args...)
}

func (h *credentialHarness) getToken(args ...string) (string, error) {
	h.t.Helper()
	return h.run(append([]string{"auth", "get-token", "--server", h.serverURL}, args...)...)
}

func (h *credentialHarness) token(args ...string) (string, error) {
	h.t.Helper()
	return h.run(append([]string{"token", "--server", h.serverURL}, args...)...)
}

func (h *credentialHarness) cachedKeys() []string {
	h.t.Helper()

	cf, err := config.LoadFromPath(filepath.Join(h.configDir, cfgFile))
	require.NoError(h.t, err)
	sc := cf.Servers[h.serverURL]
	require.NotNil(h.t, sc)

	var keys []string
	for key, cred := range sc.KubeCredentials {
		if cred != nil {
			keys = append(keys, key)
		}
	}
	return keys
}

func decodeBearerToken(t *testing.T, out string) string {
	t.Helper()

	var cred config.ExecCredential
	require.NoError(t, json.Unmarshal([]byte(out), &cred))
	require.NotNil(t, cred.Status)
	return cred.Status.Token
}

type credentialCall struct {
	args       []string
	wantToken  string
	wantLogins int
}

func TestAuthGetTokenCache(t *testing.T) {
	tests := []struct {
		name  string
		calls []credentialCall
		// wantKeys are the cached credential names after all calls.
		wantKeys []string
		// wantResponseTypes are the response types of the sign-ins, in order.
		wantResponseTypes []string
	}{
		{
			name: "user id and cluster",
			calls: []credentialCall{
				{args: []string{"--cluster", "c-1", "--user-id", "u-1", "--auth-provider", "localProvider"}, wantToken: "token-1:secret", wantLogins: 1},
				{args: []string{"--cluster", "c-1", "--user-id", "u-1", "--auth-provider", "localProvider"}, wantToken: "token-1:secret", wantLogins: 1},
			},
			wantKeys:          []string{"u-1_c-1"},
			wantResponseTypes: []string{"kubeconfig_c-1"},
		},
		{
			name: "without cluster",
			calls: []credentialCall{
				{args: []string{"--user-id", "u-1"}, wantToken: "token-1:secret", wantLogins: 1},
				{args: []string{"--user-id", "u-1"}, wantToken: "token-1:secret", wantLogins: 1},
			},
			wantKeys:          []string{"u-1_"},
			wantResponseTypes: []string{"kubeconfig"},
		},
		{
			name: "per user id",
			calls: []credentialCall{
				{args: []string{"--cluster", "c-1", "--user-id", "u-1"}, wantToken: "token-1:secret", wantLogins: 1},
				{args: []string{"--cluster", "c-1", "--user-id", "u-2"}, wantToken: "token-2:secret", wantLogins: 2},
				{args: []string{"--cluster", "c-1", "--user-id", "u-1"}, wantToken: "token-1:secret", wantLogins: 2},
			},
			wantKeys:          []string{"u-1_c-1", "u-2_c-1"},
			wantResponseTypes: []string{"kubeconfig_c-1", "kubeconfig_c-1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newCredentialHarness(t)

			for _, call := range tc.calls {
				out, err := h.getToken(call.args...)
				require.NoError(t, err)
				assert.Equal(t, call.wantToken, decodeBearerToken(t, out))
				assert.Len(t, h.logins, call.wantLogins)
			}

			assert.ElementsMatch(t, tc.wantKeys, h.cachedKeys())
			var responseTypes []string
			for _, login := range h.logins {
				responseTypes = append(responseTypes, login.ResponseType)
			}
			assert.Equal(t, tc.wantResponseTypes, responseTypes)
		})
	}
}

func TestSignInPromptDefault(t *testing.T) {
	tests := []struct {
		name         string
		run          func(h *credentialHarness, args ...string) (string, error)
		args         []string
		wantPrompts  []string
		wantUsername string
		wantKeys     []string
	}{
		{
			name:         "auth get-token has no default",
			run:          (*credentialHarness).getToken,
			args:         []string{"--cluster", "c-1", "--user-id", "u-1"},
			wantPrompts:  []string{"Enter username: ", "Enter password: "},
			wantUsername: "",
			wantKeys:     []string{"u-1_c-1"},
		},
		{
			name:         "token defaults to the user",
			run:          (*credentialHarness).token,
			args:         []string{"--cluster", "c-1", "--user", "bob"},
			wantPrompts:  []string{"Enter username [bob]: ", "Enter password: "},
			wantUsername: "bob",
			wantKeys:     []string{"bob_c-1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newCredentialHarness(t)
			h.username = ""

			_, err := tc.run(h, tc.args...)
			require.NoError(t, err)

			assert.Equal(t, tc.wantPrompts, h.prompts)
			require.Len(t, h.logins, 1)
			assert.Equal(t, tc.wantUsername, h.logins[0].Username)
			assert.Equal(t, tc.wantKeys, h.cachedKeys())
		})
	}
}

func TestAuthGetTokenRequiresUserID(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "missing", args: []string{"--cluster", "c-1"}},
		{name: "empty", args: []string{"--cluster", "c-1", "--user-id", ""}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newCredentialHarness(t)

			out, err := h.getToken(tc.args...)

			require.ErrorContains(t, err, "user-id is required")
			assert.Empty(t, out)
			assert.Empty(t, h.logins)
			assert.Empty(t, h.prompts)
		})
	}
}

func TestAuthProviderSelection(t *testing.T) {
	const (
		getToken = "auth get-token"
		token    = "token"
	)

	tests := []struct {
		name      string
		command   string
		providers string
		// provider is the --auth-provider value, empty for none.
		provider string
		// choice is the answer to the provider selection prompt.
		choice     string
		wantType   string
		wantErr    string
		wantPrompt bool
	}{
		{name: "get-token name", command: getToken, providers: manyProviders, provider: "local", wantType: "localProvider"},
		{name: "get-token type", command: getToken, providers: manyProviders, provider: "localProvider", wantType: "localProvider"},
		{name: "get-token other name", command: getToken, providers: manyProviders, provider: "openldap", wantType: "openLdapProvider"},
		{name: "get-token other type", command: getToken, providers: manyProviders, provider: "openLdapProvider", wantType: "openLdapProvider"},
		{name: "get-token third name", command: getToken, providers: manyProviders, provider: "freeipa", wantType: "freeIpaProvider"},
		{name: "get-token name wins over type", command: getToken, providers: collidingProviders, provider: "openLdapProvider", wantType: "freeIpaProvider"},
		{name: "get-token unknown", command: getToken, providers: manyProviders, provider: "nosuch", wantErr: "provider nosuch not found"},
		{name: "get-token chosen from the prompt", command: getToken, providers: manyProviders, choice: "1", wantType: "openLdapProvider", wantPrompt: true},
		{name: "get-token unnamed providers are not matched by an empty value", command: getToken, providers: unnamedProviders, choice: "1", wantType: "openLdapProvider", wantPrompt: true},
		{name: "token type", command: token, providers: manyProviders, provider: "openLdapProvider", wantType: "openLdapProvider"},
		{name: "token name fails", command: token, providers: manyProviders, provider: "local", wantErr: "provider local not found"},
		{name: "token matches the type even if it is another provider's name", command: token, providers: collidingProviders, provider: "openLdapProvider", wantType: "openLdapProvider"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newCredentialHarness(t)
			h.providers = tc.providers
			h.providerChoice = tc.choice

			run, args := h.getToken, []string{"--user-id", "u-1"}
			if tc.command == token {
				run, args = h.token, []string{"--user", "bob"}
			}
			if tc.provider != "" {
				args = append(args, "--auth-provider", tc.provider)
			}
			out, err := run(args...)

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				assert.Empty(t, out)
				assert.Empty(t, h.logins)
				return
			}
			require.NoError(t, err)
			require.Len(t, h.logins, 1)
			assert.Equal(t, tc.wantType, h.logins[0].Type)
			if tc.wantPrompt {
				assert.Contains(t, h.prompts, "Select auth provider: ")
			} else {
				assert.NotContains(t, h.prompts, "Select auth provider: ")
			}
		})
	}
}

// TestListAuthProviderNames asserts that listAuthProviders maps the name of
// each supported provider, whatever its sign-in method, to its type, on
// /v1-public and on the /v3-public fallback, and leaves out unnamed providers.
func TestListAuthProviderNames(t *testing.T) {
	const providers = `{"data": [
		{"id": "local", "type": "localProvider"},
		{"id": "okta", "type": "oktaProvider"},
		{"id": "azuread", "type": "azureADProvider"},
		{"type": "openLdapProvider"}
	]}`
	want := map[string]string{
		"local":   "localProvider",
		"okta":    "oktaProvider",
		"azuread": "azureADProvider",
	}

	tests := []struct {
		name        string
		useV1Public bool
		handler     http.HandlerFunc
	}{
		{
			name:        "v1-public",
			useV1Public: true,
			handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, providers)
			},
		},
		{
			name:        "v3-public fallback",
			useV1Public: false,
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/v1-public/authproviders" {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				fmt.Fprint(w, providers)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			t.Cleanup(server.Close)

			// A transport without a proxy function: the default one reads the
			// proxy environment once per process, which TestNewHTTPClient sets.
			client := &http.Client{Timeout: time.Second, Transport: &http.Transport{}}
			list, typesByName, useV1Public, err := listAuthProviders(client, server.URL, true)

			require.NoError(t, err)
			assert.Equal(t, tc.useV1Public, useV1Public)
			assert.Len(t, list, 4)
			assert.Equal(t, want, typesByName)
		})
	}
}

// TestTokenHelp asserts that the help text of the token command, rendered with
// the default template under the test root command, lists its usage, flags
// (names, order and usage strings) and the delete subcommand as written below.
func TestTokenHelp(t *testing.T) {
	// ConfigDir reads HOME, or USERPROFILE on Windows.
	home := filepath.Join(t.TempDir(), "home")
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	out, err := runRoot(t, t.TempDir(), "token", "--help")
	require.NoError(t, err)

	want := "NAME:\n" +
		"   rancher token - Authenticate and generate new kubeconfig token\n" +
		"\n" +
		"USAGE:\n" +
		"   rancher token [command [command options]]\n" +
		"\n" +
		"COMMANDS:\n" +
		"   delete  Delete cached token used for kubectl login at [CONFIGDIR] \n" +
		"            \n" +
		"           Example:\n" +
		"             # Delete a cached credential\n" +
		"             $ rancher token delete cluster1_c-1234\n" +
		"\n" +
		"             # Delete multiple cached credentials\n" +
		"             $ rancher token delete cluster1_c-1234 cluster2_c-2345\n" +
		"\n" +
		"             # Delete all credentials\n" +
		"             $ rancher token delete all\n" +
		"\n" +
		"\n" +
		"OPTIONS:\n" +
		"   --server string         Name of rancher server\n" +
		"   --user string           user-id\n" +
		"   --cluster string        cluster-id\n" +
		"   --auth-provider string  Name of Auth Provider to use for authentication\n" +
		"   --auth-flow string      Auth flow to use for OAuth providers: 'devicecode' (default) or 'authcode'\n" +
		"   --cacerts string        Location of CaCerts to use\n" +
		"   --skip-verify           Skip verification of the CACerts presented by the Server\n" +
		"   --help, -h              show help\n"
	want = strings.ReplaceAll(want, "CONFIGDIR", filepath.Join(home, ".rancher"))
	assert.Equal(t, want, out)
}

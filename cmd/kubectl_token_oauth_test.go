package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	apiv3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestGenerateState(t *testing.T) {
	t.Parallel()

	state, err := generateState()

	require.NoError(t, err)
	assert.NotEmpty(t, state)
	assert.Greater(t, len(state), 20, "state should be sufficiently long")

	state2, err := generateState()
	require.NoError(t, err)

	assert.NotEqual(t, state, state2, "consecutive state generations should be unique")
}

func TestStartCallbackServerStateValidation(t *testing.T) {
	t.Parallel()

	expectedState := "test-state-12345"
	expectedCode := "valid-code-123"

	tests := []struct {
		name          string
		queryParams   string
		expectedError string
		shouldSucceed bool
	}{
		{
			name:          "valid state and code",
			queryParams:   fmt.Sprintf("?state=%s&code=%s", expectedState, expectedCode),
			shouldSucceed: true,
		},
		{
			name:          "invalid state",
			queryParams:   fmt.Sprintf("?state=wrong-state&code=%s", expectedCode),
			expectedError: "invalid state parameter",
		},
		{
			name:          "missing code",
			queryParams:   fmt.Sprintf("?state=%s", expectedState),
			expectedError: "missing authorization code",
		},
		{
			name:          "error from IdP",
			queryParams:   "?error=access_denied&error_description=User%20denied%20access",
			expectedError: "authentication error: access_denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultCh := make(chan callbackResult, 1)

			listener, err := net.Listen("tcp", "127.0.0.1:0")
			require.NoError(t, err)
			defer listener.Close()

			srv := startCallbackServer(listener, expectedState, resultCh)
			defer srv.Close()

			port := listener.Addr().(*net.TCPAddr).Port
			baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

			// Clear the channel.
			select {
			case <-resultCh:
			default:
			}

			resp, err := http.Get(baseURL + tt.queryParams)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Wait for the result.
			select {
			case result := <-resultCh:
				if tt.shouldSucceed {
					require.NoError(t, result.err)
					assert.Equal(t, "valid-code-123", result.code)
					assert.Equal(t, expectedState, result.state)
				} else {
					require.Error(t, result.err)
					assert.ErrorContains(t, result.err, tt.expectedError)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("timeout waiting for callback result")
			}
		})
	}
}

func TestStartCallbackServerMultipleRequests(t *testing.T) {
	t.Parallel()

	expectedState := "test-state-multi"
	resultCh := make(chan callbackResult, 1)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	srv := startCallbackServer(listener, expectedState, resultCh)
	defer srv.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	// Send first request.
	resp1, err := http.Get(fmt.Sprintf("%s?state=%s&code=code1", baseURL, expectedState))
	require.NoError(t, err)
	resp1.Body.Close()

	// Send second request immediately.
	resp2, err := http.Get(fmt.Sprintf("%s?state=%s&code=code2", baseURL, expectedState))
	require.NoError(t, err)
	resp2.Body.Close()

	// The channel has buffer size 1, so only one result should be received.
	result := <-resultCh
	require.NoError(t, result.err)

	// Verify only one result was sent.
	select {
	case <-resultCh:
		t.Fatal("unexpected second result in channel")
	case <-time.After(100 * time.Millisecond):
		// No second result.
	}
}

func TestRancherLogin(t *testing.T) {
	t.Parallel()

	expectedToken := "test-rancher-token-123"
	expiresAt := time.Now().Add(time.Hour).Format(time.RFC3339)

	tests := []struct {
		name         string
		useV1Public  bool
		statusCode   int
		responseBody string
		shouldError  bool
		errorMsg     string
	}{
		{
			name:        "successful login with v1-public",
			useV1Public: true,
			statusCode:  http.StatusCreated,
			responseBody: fmt.Sprintf(`{
				"token": "%s",
				"expiresAt": "%s",
				"type": "token"
			}`, expectedToken, expiresAt),
			shouldError: false,
		},
		{
			name:        "successful login with v3-public",
			useV1Public: false,
			statusCode:  http.StatusCreated,
			responseBody: fmt.Sprintf(`{
				"token": "%s",
				"expiresAt": "%s",
				"type": "token"
			}`, expectedToken, expiresAt),
			shouldError: false,
		},
		{
			name:         "unauthorized",
			useV1Public:  true,
			statusCode:   http.StatusUnauthorized,
			responseBody: `{"type": "error", "message": "invalid token"}`,
			shouldError:  true,
			errorMsg:     "401",
		},
		{
			name:         "invalid JSON response",
			useV1Public:  true,
			statusCode:   http.StatusCreated,
			responseBody: `this is not json`,
			shouldError:  true,
			errorMsg:     "error unmarshaling",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)

				if tt.useV1Public {
					assert.Contains(t, r.URL.Path, "/v1-public/login")
				} else {
					assert.Contains(t, r.URL.Path, "/v3-public/")
					assert.Equal(t, "login", r.URL.Query().Get("action"))
				}

				if tt.statusCode == http.StatusCreated {
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
					var body map[string]any
					err := json.NewDecoder(r.Body).Decode(&body)
					require.NoError(t, err)
					assert.Equal(t, "azureADProvider", body["type"])
					assert.Equal(t, "kubeconfig_c-12345", body["responseType"])
					assert.Equal(t, "test-id-token", body["id_token"])
				}

				w.WriteHeader(tt.statusCode)
				fmt.Fprint(w, tt.responseBody)
			}))
			defer server.Close()

			client := server.Client()
			input := &LoginInput{
				server:       server.URL,
				userID:       "test-user",
				clusterID:    "c-12345",
				authProvider: "azureADProvider",
			}

			oauthToken := &oauth2.Token{
				AccessToken: "test-access-token",
			}
			oauthToken = oauthToken.WithExtra(map[string]any{
				"id_token": "test-id-token",
			})

			token, err := rancherLogin(client, input, oauthToken, tt.useV1Public)

			if tt.shouldError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.ErrorContains(t, err, tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, token)
				assert.Equal(t, expectedToken, token.Token)
			}
		})
	}
}

func TestOauthDeviceCodeAuth(t *testing.T) {
	t.Parallel()

	// Fake OAuth server.
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/device":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{
				"device_code": "test-device-code",
				"user_code": "TEST-CODE",
				"verification_uri": "https://example.com/activate",
				"expires_in": 600,
				"interval": 5
			}`)
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{
				"access_token": "test-access-token",
				"token_type": "Bearer",
				"expires_in": 3600,
				"id_token": "test-id-token"
			}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer oauthServer.Close()

	// Fake Rancher server
	rancherServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"token": "rancher-token-123"}`)
	}))
	defer rancherServer.Close()

	client := &http.Client{Timeout: 10 * time.Second}
	provider := &apiv3.AzureADProvider{
		AuthProvider: apiv3.AuthProvider{
			Type: "azureADProvider",
		},
		OAuthProvider: apiv3.OAuthProvider{
			ClientID: "test-client-id",
			Scopes:   []string{"openid"},
			OAuthEndpoint: apiv3.OAuthEndpoint{
				DeviceAuthURL: oauthServer.URL + "/device",
				TokenURL:      oauthServer.URL + "/token",
			},
		},
	}

	input := &LoginInput{
		server:       rancherServer.URL,
		userID:       "test-user",
		clusterID:    "c-12345",
		authProvider: "azureADProvider",
	}

	token, err := oauthDeviceCodeAuth(client, input, provider, true)

	require.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "rancher-token-123", token.Token)
}

func TestOauthAuthCodeAuth(t *testing.T) {
	t.Parallel()

	// Fake OAuth server.
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth":
			redirectURL, err := url.Parse(r.URL.Query().Get("redirect_uri"))
			require.NoError(t, err)

			v := url.Values{
				"state": {r.URL.Query().Get("state")},
				"code":  {r.URL.Query().Get("code_challenge")},
			}
			redirectURL.RawQuery = v.Encode()

			http.Redirect(w, r, redirectURL.String(), http.StatusFound)
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{
				"access_token": "test-access-token",
				"token_type": "Bearer",
				"expires_in": 3600,
				"id_token": "test-id-token"
			}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer oauthServer.Close()

	// Fake Rancher server.
	rancherServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"token": "rancher-token-123"}`)
	}))
	defer rancherServer.Close()

	provider := &apiv3.AzureADProvider{
		AuthProvider: apiv3.AuthProvider{
			Type: "azureADProvider",
		},
		OAuthProvider: apiv3.OAuthProvider{
			ClientID: "test-client-id",
			Scopes:   []string{"openid"},
			OAuthEndpoint: apiv3.OAuthEndpoint{
				AuthURL:  oauthServer.URL + "/auth",
				TokenURL: oauthServer.URL + "/token",
			},
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}
	input := &LoginInput{
		server:       rancherServer.URL,
		userID:       "test-user",
		clusterID:    "c-12345",
		authProvider: "azureADProvider",
	}

	openBrowser := func(url string) error {
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code %d", resp.StatusCode)
		}

		return nil
	}

	t.Run("success", func(t *testing.T) {
		err := os.Setenv("CATTLE_OAUTH_CALLBACK_PORT", "65535")
		require.NoError(t, err)
		defer os.Unsetenv("CATTLE_OAUTH_CALLBACK_PORT")

		_, err = oauthAuthCodeAuth(client, input, provider, time.Minute, true, openBrowser)

		require.NoError(t, err)

	})

	t.Run("timeout waiting for authentication", func(t *testing.T) {
		openBrowser := func(url string) error { return nil }

		// Use a very short timeout for testing.
		timeoutAfter := 100 * time.Millisecond

		_, err := oauthAuthCodeAuth(client, input, provider, timeoutAfter, true, openBrowser)

		require.Error(t, err)
		assert.ErrorContains(t, err, "timed out waiting for browser authentication")
	})
}

func TestGetCallbackPort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		portEnv     string
		port        int
		shouldError bool
		errorMsg    string
	}{
		{
			name:    "empty port uses random",
			portEnv: "",
			port:    0,
		},
		{
			name:    "valid high port",
			portEnv: "8080",
			port:    8080,
		},
		{
			name:    "port 0 uses random",
			portEnv: "0",
			port:    0,
		},
		{
			name:        "invalid port string",
			portEnv:     "not-a-number",
			shouldError: true,
			errorMsg:    "invalid callback port value",
		},
		{
			name:        "negative port",
			portEnv:     "-1",
			shouldError: true,
			errorMsg:    "callback port value must be between 0 and 65535",
		},
		{
			name:        "port too high",
			portEnv:     "99999",
			shouldError: true,
			errorMsg:    "callback port value must be between 0 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.portEnv != "" {
				os.Setenv("CATTLE_OAUTH_CALLBACK_PORT", tt.portEnv)
				defer os.Unsetenv("CATTLE_OAUTH_CALLBACK_PORT")
			}

			port, err := getCallbackPort()
			if tt.shouldError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.ErrorContains(t, err, tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.port, port)
			}
		})
	}
}

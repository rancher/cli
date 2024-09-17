package cliclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/rancher/cli/cliclient/mocks"
	"github.com/rancher/cli/config"
	"github.com/rancher/norman/types"
)

func Test_createClientOpts(t *testing.T) {
	conf := &config.ServerConfig{
		URL:       "https://a.b",
		AccessKey: "AccessKey",
		SecretKey: "SecretKey",
		CACerts:   "CACerts",
	}

	clientOpts := createClientOpts(conf)

	assert.NotNil(t, clientOpts)
	assert.NotNil(t, clientOpts.HTTPClient)
	assert.NotNil(t, clientOpts.HTTPClient.Transport)
	assert.Equal(t, "https://a.b/v3", clientOpts.URL)
	assert.Equal(t, "AccessKey", clientOpts.AccessKey)
	assert.Equal(t, "SecretKey", clientOpts.SecretKey)
	assert.Equal(t, "CACerts", clientOpts.CACerts)
}

func TestHTTPClient_New(t *testing.T) {
	client := DefaultHTTPClient.New()

	assert.NotNil(t, client)
	assert.Equal(t, time.Minute, client.Timeout)

	assert.NotNil(t, client.Transport)
	transport, is := client.Transport.(*http.Transport)
	assert.True(t, is)
	assert.NotNil(t, transport)
	assert.NotNil(t, transport.DialContext)
	assert.NotNil(t, transport.Proxy)
	assert.Equal(t, "net/http.ProxyFromEnvironment", runtime.FuncForPC(reflect.ValueOf(transport.Proxy).Pointer()).Name())
	assert.Equal(t, 10*time.Second, transport.ResponseHeaderTimeout)
}

func TestNewManagementClient(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v3/settings/cacerts":
			crt, _ := os.ReadFile("../testdata/ca-cert.pem")
			resp := map[string]string{
				"name":  "cacerts",
				"value": string(crt),
			}
			respBody, _ := json.Marshal(resp)
			_, _ = w.Write(respBody)
		case "/v3":
			schemaURL := url.URL{
				Scheme: "https",
				Host:   r.Host,
				Path:   r.URL.Path,
			}
			w.Header().Add("X-API-Schemas", schemaURL.String())
			schemas := &types.SchemaCollection{
				Data: []types.Schema{},
			}
			respBody, _ := json.Marshal(schemas)
			_, _ = w.Write(respBody)
		default:
			fmt.Println(r.URL.Path)
		}
	}))
	defer server.Close()
	server.StartTLS()

	fakeClient := server.Client()
	mockClient := mocks.NewMockHTTPClienter(t)
	mockClient.EXPECT().New().Return(fakeClient).Once()
	defer TestingReplaceDefaultHTTPClient(mockClient)()
	defer TestingForceClientInsecure()()

	conf := &config.ServerConfig{
		URL: server.URL,
	}

	mc, err := NewManagementClient(conf)

	assert.Nil(t, err)
	assert.NotNil(t, mc)
}

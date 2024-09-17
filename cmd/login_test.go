package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/rancher/norman/types"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"

	"github.com/rancher/cli/cliclient"
	"github.com/rancher/cli/cliclient/mocks"
	"github.com/rancher/cli/config"
)

func Test_getCertFromServer(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v3/settings/cacerts":
			crt, _ := os.ReadFile("../testdata/ca-cert.pem")
			resp := &CACertResponse{
				Name:  "cacerts",
				Value: string(crt),
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
	mockClient.EXPECT().New().Return(fakeClient)
	defer cliclient.TestingReplaceDefaultHTTPClient(mockClient)()
	defer cliclient.TestingForceClientInsecure()()

	app := cli.NewApp()
	appFlags := flag.NewFlagSet("flags", flag.ContinueOnError)
	appFlags.Bool("skip-verify", true, "")
	cliCtx := cli.NewContext(app, appFlags, nil)
	conf := &config.ServerConfig{
		URL: server.URL,
	}

	masterClient, err := getCertFromServer(cliCtx, conf)

	assert.Nil(t, err)
	assert.NotNil(t, masterClient)
	mockClient.AssertNumberOfCalls(t, "New", 2)
}

package cmd

import (
	"archive/zip"
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

func Test_getSSHKey(t *testing.T) {
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
		case "/sshkeys":
			zipWriter := zip.NewWriter(w)
			idW, _ := zipWriter.Create("id_rsa")
			_, _ = idW.Write([]byte("RSA"))
			zipWriter.Close()
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
	sshkeysURL, _ := url.JoinPath(server.URL, "sshkeys")

	masterClient, err := getCertFromServer(cliCtx, conf)

	assert.Nil(t, err)
	assert.NotNil(t, masterClient)

	_, _, err = getSSHKey(masterClient, sshkeysURL, "node")

	assert.Nil(t, err)
	mockClient.AssertNumberOfCalls(t, "New", 3)
}

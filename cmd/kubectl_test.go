package cmd

import (
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rancher/cli/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServerHTTPClient(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	t.Run("trusts the CA saved at login", func(t *testing.T) {
		serverConfig := &config.ServerConfig{
			CACerts: string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw})),
		}

		client, err := newServerHTTPClient(serverConfig)
		require.NoError(t, err)

		resp, err := client.Get(srv.URL)
		require.NoError(t, err)
		resp.Body.Close()
	})

	t.Run("invalid CA certs", func(t *testing.T) {
		serverConfig := &config.ServerConfig{
			CACerts: "not a certificate",
		}

		client, err := newServerHTTPClient(serverConfig)
		assert.ErrorContains(t, err, "error creating TLS config")
		assert.Nil(t, client)
	})
}

package cmd_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/rancher/cli/cmd"
	"github.com/rancher/cli/config"
	"github.com/stretchr/testify/assert"
)

func TestServerCurrentCommand(t *testing.T) {
	tt := []struct {
		name           string
		config         *config.Config
		expectedOutput string
		expectedErr    string
	}{
		{
			name:           "existing current server set",
			config:         newTestConfig(),
			expectedOutput: "Name: server1 URL: https://myserver-1.com\n",
		},
		{
			name: "empty current server",
			config: func() *config.Config {
				cfg := newTestConfig()
				cfg.CurrentServer = ""
				return cfg
			}(),
			expectedErr: "Current server not set",
		},
		{
			name: "non existing current server set",
			config: &config.Config{
				CurrentServer: "notfound-server",
				Servers: map[string]*config.ServerConfig{
					"my-server": {URL: "https://myserver.com"},
				},
			},
			expectedErr: "Current server not set",
		},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out := &bytes.Buffer{}

			err := cmd.ServerCurrent(out, tc.config)
			if tc.expectedErr != "" {
				assert.EqualError(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedOutput, out.String())
		})
	}
}

func TestServerSwitch(t *testing.T) {
	tt := []struct {
		name                  string
		actualCurrentServer   string
		serverName            string
		expectedCurrentServer string
		expectedErr           string
	}{
		{
			name:                  "switch to different server updates the current server",
			actualCurrentServer:   "server1",
			serverName:            "server3",
			expectedCurrentServer: "server3",
		},
		{
			name:                  "switch to same server is no-op",
			actualCurrentServer:   "server1",
			serverName:            "server1",
			expectedCurrentServer: "server1",
		},
		{
			name:                  "switch to non existing server",
			actualCurrentServer:   "server1",
			serverName:            "server-nope",
			expectedCurrentServer: "server1",
			expectedErr:           "Server not found",
		},
		{
			name:                  "switch to empty server fails",
			actualCurrentServer:   "server1",
			serverName:            "",
			expectedCurrentServer: "server1",
			expectedErr:           "Server not found",
		},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpConfig, err := os.CreateTemp("", "*-rancher-config.json")
			assert.NoError(t, err)
			defer os.Remove(tmpConfig.Name())

			// setup test config
			cfg := newTestConfig()
			cfg.Path = tmpConfig.Name()
			cfg.CurrentServer = tc.actualCurrentServer

			// do test and check resulting config
			err = cmd.ServerSwitch(cfg, tc.serverName)
			if err != nil {
				assert.EqualError(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedCurrentServer, cfg.CurrentServer)
		})
	}
}

func newTestConfig() *config.Config {
	return &config.Config{
		CurrentServer: "server1",
		Servers: map[string]*config.ServerConfig{
			"server1": {URL: "https://myserver-1.com"},
			"server2": {URL: "https://myserver-2.com"},
			"server3": {URL: "https://myserver-3.com"},
		},
	}
}

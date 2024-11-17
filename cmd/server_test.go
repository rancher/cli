package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/rancher/cli/config"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestServerCommand(t *testing.T) {
	// Checking if two functions are equal is not possible
	// Hence, we set them to nil here to skip checking
	got := ServerCommand()
	got.Before = nil
	for idx := range got.Subcommands {
		got.Subcommands[idx].Action = nil
	}

	expected := cli.Command{
		Name:  "server",
		Usage: "Operations for the server",
		Description: `Switch or view the server currently in focus.
`,
		Subcommands: []cli.Command{
			{
				Name:  "current",
				Usage: "Display the current server",
			},
			{
				Name:      "delete",
				Usage:     "Delete a server from the local config",
				ArgsUsage: "[SERVER_NAME]",
				Description: `
The server arg is optional, if not passed in a list of available servers will
be displayed and one can be selected.
`,
			},
			{
				Name:      "ls",
				Usage:     "List all servers",
				ArgsUsage: "None",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "format",
						Usage: "'json', 'yaml' or Custom format: '{{.Name}} {{.URL}}'",
					},
				},
			},
			{
				Name:      "switch",
				Usage:     "Switch to a new server",
				ArgsUsage: "[SERVER_NAME]",
				Description: `
		The server arg is optional, if not passed in a list of available servers will
		be displayed and one can be selected.
		`,
			},
		},
	}

	assert.Equal(t, expected, got)
}

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

			err := serverCurrent(out, tc.config)
			if tc.expectedErr != "" {
				assert.EqualError(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedOutput, out.String())
		})
	}
}

func TestServerDelete(t *testing.T) {
	tt := []struct {
		name                  string
		actualCurrentServer   string
		serverToDelete        string
		expectedCurrentServer string
		expectedErr           string
	}{
		{
			name:                  "delete a different server will delete it",
			actualCurrentServer:   "server1",
			serverToDelete:        "server3",
			expectedCurrentServer: "server1",
		},
		{
			name:                  "delete the same server will blank the current",
			actualCurrentServer:   "server1",
			serverToDelete:        "server1",
			expectedCurrentServer: "",
		},
		{
			name:                  "delete a non existing server",
			actualCurrentServer:   "server1",
			serverToDelete:        "server-nope",
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
			err = serverDelete(cfg, tc.serverToDelete)
			if err != nil {
				assert.EqualError(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedCurrentServer, cfg.CurrentServer)
			assert.Empty(t, cfg.Servers[tc.serverToDelete])
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
			err = serverSwitch(cfg, tc.serverName)
			if err != nil {
				assert.EqualError(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedCurrentServer, cfg.CurrentServer)
		})
	}
}

func TestServerLs(t *testing.T) {
	tt := []struct {
		name           string
		config         *config.Config
		format         string
		expectedOutput string
		expectedErr    bool
	}{
		{
			name: "list servers",
			expectedOutput: `CURRENT   NAME      URL
*         server1   https://myserver-1.com
          server2   https://myserver-2.com
          server3   https://myserver-3.com
`,
		},
		{
			name:           "list empty config",
			config:         &config.Config{},
			format:         "",
			expectedOutput: "CURRENT   NAME      URL\n",
		},
		{
			name:   "list servers with json format",
			format: "json",
			expectedOutput: `{"Index":1,"Current":"*","Name":"server1","URL":"https://myserver-1.com"}
{"Index":2,"Current":"","Name":"server2","URL":"https://myserver-2.com"}
{"Index":3,"Current":"","Name":"server3","URL":"https://myserver-3.com"}
`,
		},
		{
			name:   "list servers with yaml format",
			format: "yaml",
			expectedOutput: `Current: '*'
Index: 1
Name: server1
URL: https://myserver-1.com

Current: ""
Index: 2
Name: server2
URL: https://myserver-2.com

Current: ""
Index: 3
Name: server3
URL: https://myserver-3.com

`,
		},
		{
			name:   "list servers with custom format",
			format: "{{.URL}}",
			expectedOutput: `https://myserver-1.com
https://myserver-2.com
https://myserver-3.com
`,
		},
		{
			name:        "list servers with custom format",
			format:      "{{.err}}",
			expectedErr: true,
		},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out := &bytes.Buffer{}

			if tc.config == nil {
				tc.config = newTestConfig()
			}

			// do test and check resulting config
			err := serverLs(out, tc.config, tc.format)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedOutput, out.String())
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

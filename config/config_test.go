package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	validFile = `
{
  "Servers": {
    "rancherDefault": {
      "accessKey": "the-access-key",
      "secretKey": "the-secret-key",
      "tokenKey": "the-token-key",
      "url": "https://example.com",
      "project": "cluster-id:project-id",
      "cacert": "",
      "kubeCredentials": null,
      "kubeConfigs": null
    }
  },
  "CurrentServer": "rancherDefault"
}`
	invalidFile = `invalid config file`
)

func TestGetFilePermissionWarnings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		mode             os.FileMode
		expectedWarnings int
	}{
		{
			name:             "neither group-readable nor world-readable",
			mode:             os.FileMode(0600),
			expectedWarnings: 0,
		},
		{
			name:             "group-readable and world-readable",
			mode:             os.FileMode(0644),
			expectedWarnings: 2,
		},
		{
			name:             "group-readable",
			mode:             os.FileMode(0640),
			expectedWarnings: 1,
		},
		{
			name:             "world-readable",
			mode:             os.FileMode(0604),
			expectedWarnings: 1,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert := assert.New(t)

			dir, err := os.MkdirTemp("", "rancher-cli-test-*")
			assert.NoError(err)
			defer os.RemoveAll(dir)

			path := filepath.Join(dir, "cli2.json")
			err = os.WriteFile(path, []byte(validFile), tt.mode)
			assert.NoError(err)

			warnings, err := GetFilePermissionWarnings(path)
			assert.NoError(err)
			assert.Len(warnings, tt.expectedWarnings)
		})
	}
}

func TestPermission(t *testing.T) {
	t.Parallel()

	// New config files should have 0600 permissions
	t.Run("new config file", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)

		dir, err := os.MkdirTemp("", "rancher-cli-test-*")
		assert.NoError(err)
		defer os.RemoveAll(dir)

		path := filepath.Join(dir, "cli2.json")
		conf, err := LoadFromPath(path)
		assert.NoError(err)

		err = conf.Write()
		assert.NoError(err)

		info, err := os.Stat(path)
		assert.NoError(err)
		assert.Equal(os.FileMode(0600), info.Mode())

		// make sure new file doesn't create permission warnings
		warnings, err := GetFilePermissionWarnings(path)
		assert.NoError(err)
		assert.Len(warnings, 0)
	})
	// Already existing config files should keep their current permissions
	t.Run("existing config file", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)

		dir, err := os.MkdirTemp("", "rancher-cli-test-*")
		assert.NoError(err)
		defer os.RemoveAll(dir)

		path := filepath.Join(dir, "cli2.json")
		err = os.WriteFile(path, []byte(validFile), 0700)
		assert.NoError(err)

		conf, err := LoadFromPath(path)
		assert.NoError(err)

		err = conf.Write()
		assert.NoError(err)

		info, err := os.Stat(path)
		assert.NoError(err)
		assert.Equal(os.FileMode(0700), info.Mode())
	})
}

func TestLoadFromPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		content      string
		expectedConf Config
		expectedErr  bool
	}{
		{
			name:    "valid config",
			content: validFile,
			expectedConf: Config{
				Servers: map[string]*ServerConfig{
					"rancherDefault": {
						AccessKey: "the-access-key",
						SecretKey: "the-secret-key",
						TokenKey:  "the-token-key",
						URL:       "https://example.com",
						Project:   "cluster-id:project-id",
						CACerts:   "",
					},
				},
				CurrentServer: "rancherDefault",
				Helper:        "built-in",
			},
		},
		{
			name:    "invalid config",
			content: invalidFile,
			expectedConf: Config{
				Servers: map[string]*ServerConfig{},
				Helper:  "built-in",
			},
			expectedErr: true,
		},
		{
			name:    "non existing file",
			content: "",
			expectedConf: Config{
				Servers:       map[string]*ServerConfig{},
				CurrentServer: "",
				Helper:        "built-in",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert := assert.New(t)

			dir, err := os.MkdirTemp("", "rancher-cli-test-*")
			assert.NoError(err)
			defer os.RemoveAll(dir)

			path := filepath.Join(dir, "cli2.json")
			// make sure the path points to the temp dir created in the test
			tt.expectedConf.Path = path

			if tt.content != "" {
				err = os.WriteFile(path, []byte(tt.content), 0600)
				assert.NoError(err)
			}

			conf, err := LoadFromPath(path)
			if tt.expectedErr {
				assert.Error(err)
				// We kept the old behavior of returning a valid config even in
				// case of an error so we assert it here. If you change this
				// behavior, make sure there aren't any regressions.
				assert.Equal(tt.expectedConf, conf)
				return
			}

			assert.NoError(err)
			assert.Equal(tt.expectedConf, conf)
		})
	}
}

func TestLoadWithHelper(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		helperScript string
		expectedConf Config
		expectedErr  bool
	}{
		{
			name: "valid helper response",
			helperScript: `#!/bin/bash
echo '{"Servers":{"test":{"accessKey":"key","url":"https://test.com"}},"CurrentServer":"test"}'`,
			expectedConf: Config{
				Servers: map[string]*ServerConfig{
					"test": {
						AccessKey: "key",
						URL:       "https://test.com",
					},
				},
				CurrentServer: "test",
			},
		},
		{
			name:         "helper not found",
			helperScript: "",
			expectedErr:  true,
		},
		{
			name: "helper returns invalid json",
			helperScript: `#!/bin/bash
echo 'invalid json'`,
			expectedErr: true,
		},
		{
			name: "helper exits with error",
			helperScript: `#!/bin/bash
exit 1`,
			expectedErr: true,
		},
		{
			name: "helper returns empty output",
			helperScript: `#!/bin/bash
echo ''`,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert := assert.New(t)

			var helperPath string
			if tt.helperScript != "" {
				// Create a temporary helper script
				dir, err := os.MkdirTemp("", "rancher-cli-test-*")
				assert.NoError(err)
				defer os.RemoveAll(dir)

				helperPath = filepath.Join(dir, "test-helper")
				err = os.WriteFile(helperPath, []byte(tt.helperScript), 0755)
				assert.NoError(err)
			} else {
				// Use non-existent helper
				helperPath = "non-existent-helper"
			}

			conf, err := LoadWithHelper(helperPath)
			if tt.expectedErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			assert.Equal(helperPath, conf.Helper)
			assert.Equal(tt.expectedConf.CurrentServer, conf.CurrentServer)
			assert.Equal(len(tt.expectedConf.Servers), len(conf.Servers))
			assert.Empty(conf.Path) // Path should be empty for helper-loaded configs
		})
	}
}

func TestConfigWrite(t *testing.T) {
	t.Parallel()

	t.Run("writes using native method for built-in helper", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)

		dir, err := os.MkdirTemp("", "rancher-cli-test-*")
		assert.NoError(err)
		defer os.RemoveAll(dir)

		path := filepath.Join(dir, "cli2.json")
		conf := Config{
			Path:          path,
			Helper:        "built-in",
			Servers:       make(map[string]*ServerConfig),
			CurrentServer: "test",
		}

		err = conf.Write()
		assert.NoError(err)

		// Verify file was created
		_, err = os.Stat(path)
		assert.NoError(err)
	})

	t.Run("writes using helper method for external helper", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)

		// Create a mock helper that accepts store commands
		dir, err := os.MkdirTemp("", "rancher-cli-test-*")
		assert.NoError(err)
		defer os.RemoveAll(dir)

		helperScript := `#!/bin/bash
if [ "$1" = "store" ]; then
    # Just exit successfully for the test
    exit 0
fi
exit 1`

		helperPath := filepath.Join(dir, "test-helper")
		err = os.WriteFile(helperPath, []byte(helperScript), 0755)
		assert.NoError(err)

		conf := Config{
			Helper:        helperPath,
			Servers:       make(map[string]*ServerConfig),
			CurrentServer: "test",
		}

		err = conf.Write()
		assert.NoError(err)
	})

	t.Run("helper write fails when helper not found", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)

		conf := Config{
			Helper:        "non-existent-helper",
			Servers:       make(map[string]*ServerConfig),
			CurrentServer: "test",
		}

		err := conf.Write()
		assert.Error(err)
	})

	t.Run("helper write fails when helper exits with error", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)

		// Create a mock helper that fails
		dir, err := os.MkdirTemp("", "rancher-cli-test-*")
		assert.NoError(err)
		defer os.RemoveAll(dir)

		helperScript := `#!/bin/bash
exit 1`

		helperPath := filepath.Join(dir, "test-helper")
		err = os.WriteFile(helperPath, []byte(helperScript), 0755)
		assert.NoError(err)

		conf := Config{
			Helper:        helperPath,
			Servers:       make(map[string]*ServerConfig),
			CurrentServer: "test",
		}

		err = conf.Write()
		assert.Error(err)
	})
}

func TestHelperProtocol(t *testing.T) {
	t.Parallel()

	t.Run("helper receives correct get command", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)
		require := require.New(t)

		// Create a helper that logs its arguments
		dir, err := os.MkdirTemp("", "rancher-cli-test-*")
		require.NoError(err)
		defer os.RemoveAll(dir)

		logFile := filepath.Join(dir, "args.log")
		helperScript := fmt.Sprintf(`#!/bin/bash
echo "$@" > %s
echo '{"Servers":{},"CurrentServer":""}'`, logFile)

		helperPath := filepath.Join(dir, "test-helper")
		err = os.WriteFile(helperPath, []byte(helperScript), 0755)
		require.NoError(err)

		_, err = LoadWithHelper(helperPath)
		require.NoError(err)

		// Check that the helper was called with "get"
		logContent, err := os.ReadFile(logFile)
		require.NoError(err)
		assert.Contains(string(logContent), "get")
	})

	t.Run("helper receives correct store command and data", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)
		require := require.New(t)

		// Skip on Windows since bash scripts don't work there
		if runtime.GOOS == "windows" {
			t.Skip("Skipping bash script test on Windows")
		}

		// Create a helper that logs its arguments
		dir, err := os.MkdirTemp("", "rancher-cli-test-*")
		require.NoError(err)
		defer os.RemoveAll(dir)

		logFile := filepath.Join(dir, "store.log")
		helperScript := fmt.Sprintf(`#!/bin/bash
echo "$1" > %s
echo "$2" >> %s`, logFile, logFile)

		helperPath := filepath.Join(dir, "test-helper")
		err = os.WriteFile(helperPath, []byte(helperScript), 0755)
		require.NoError(err)

		conf := Config{
			Helper:        helperPath,
			Servers:       make(map[string]*ServerConfig),
			CurrentServer: "test",
		}

		err = conf.Write()
		require.NoError(err)

		// Check that the helper was called with "store" and JSON data
		logContent, err := os.ReadFile(logFile)
		require.NoError(err)
		lines := string(logContent)
		assert.Contains(lines, "store")
		assert.Contains(lines, "test") // Should contain the CurrentServer value
	})
}

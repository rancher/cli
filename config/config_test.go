package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	validConfigContent = `
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
	invalidConfigContent = `invalid config file`
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

			path := filepath.Join(t.TempDir(), "cli2.json")
			err := os.WriteFile(path, []byte(validConfigContent), tt.mode)
			assert.NoError(t, err)

			warnings, err := GetFilePermissionWarnings(path)
			assert.NoError(t, err)
			assert.Len(t, warnings, tt.expectedWarnings)
		})
	}
}

func TestPermission(t *testing.T) {
	t.Parallel()

	// New config files should have 0600 permissions
	t.Run("new config file", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "cli2.json")
		conf, err := LoadFromPath(path)
		assert.NoError(t, err)

		err = conf.Write()
		assert.NoError(t, err)

		info, err := os.Stat(path)
		assert.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode())

		// make sure new file doesn't create permission warnings
		warnings, err := GetFilePermissionWarnings(path)
		assert.NoError(t, err)
		assert.Len(t, warnings, 0)
	})
	// Already existing config files should keep their current permissions
	t.Run("existing config file", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "cli2.json")
		err := os.WriteFile(path, []byte(validConfigContent), 0700)
		assert.NoError(t, err)

		conf, err := LoadFromPath(path)
		assert.NoError(t, err)

		err = conf.Write()
		assert.NoError(t, err)

		info, err := os.Stat(path)
		assert.NoError(t, err)
		assert.Equal(t, os.FileMode(0700), info.Mode())
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
			content: validConfigContent,
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
			},
		},
		{
			name:    "invalid config",
			content: invalidConfigContent,
			expectedConf: Config{
				Servers: map[string]*ServerConfig{},
			},
			expectedErr: true,
		},
		{
			name:    "non existing file",
			content: "",
			expectedConf: Config{
				Servers:       map[string]*ServerConfig{},
				CurrentServer: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "cli2.json")
			// make sure the path points to the temp dir created in the test
			tt.expectedConf.Path = path

			if tt.content != "" {
				err := os.WriteFile(path, []byte(tt.content), 0600)
				assert.NoError(t, err)
			}

			conf, err := LoadFromPath(path)
			if tt.expectedErr {
				assert.Error(t, err)
				// We kept the old behavior of returning a valid config even in
				// case of an error so we assert it here. If you change this
				// behavior, make sure there aren't any regressions.
				assert.Equal(t, tt.expectedConf, conf)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedConf, conf)
		})
	}
}

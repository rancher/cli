package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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

func Test_GetFilePermissionWarnings(t *testing.T) {
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

func Test_Permission(t *testing.T) {
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

func Test_LoadFromPath(t *testing.T) {
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
			},
		},
		{
			name:    "invalid config",
			content: invalidFile,
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

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

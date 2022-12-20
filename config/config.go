package config

import (
	"encoding/json"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Config holds the main config for the user
type Config struct {
	Servers map[string]*ServerConfig
	//Path to the config file
	Path string `json:"path,omitempty"`
	// CurrentServer the user has in focus
	CurrentServer string
	// Helper executable to store config
	Helper string
}

//ServerConfig holds the config for each server the user has setup
type ServerConfig struct {
	AccessKey       string                     `json:"accessKey"`
	SecretKey       string                     `json:"secretKey"`
	TokenKey        string                     `json:"tokenKey"`
	URL             string                     `json:"url"`
	Project         string                     `json:"project"`
	CACerts         string                     `json:"cacert"`
	KubeCredentials map[string]*ExecCredential `json:"kubeCredentials"`
	KubeConfigs     map[string]*api.Config     `json:"kubeConfigs"`
}

func (c Config) Write() error {
	switch c.Helper {
	case "build-in":
		return c.writeNative()
	default:
		// if rancher config was loaded by external helper
		// use the same helper to persist the config
		return c.writeWithHelper()
	}
}

func (c Config) writeNative() error {
	err := os.MkdirAll(path.Dir(c.Path), 0700)
	if err != nil {
		return err
	}

	logrus.Infof("Saving config to %s", c.Path)
	p := c.Path
	c.Path = ""
	output, err := os.Create(p)
	if err != nil {
		return err
	}
	defer output.Close()

	return json.NewEncoder(output).Encode(c)
}

func (c Config) writeWithHelper() error {
	logrus.Infof("Saving config with helper %s", c.Helper)
	jsonConfig, err := json.Marshal(c)
	if err != nil {
		return err
	}
	cmd := exec.Command(c.Helper, "store", string(jsonConfig))
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	return err
}

func (c Config) FocusedServer() *ServerConfig {
	return c.Servers[c.CurrentServer]
}

func (c ServerConfig) FocusedCluster() string {
	return strings.Split(c.Project, ":")[0]
}

func (c ServerConfig) KubeToken(key string) *ExecCredential {
	return c.KubeCredentials[key]
}

func (c ServerConfig) EnvironmentURL() (string, error) {
	url, err := baseURL(c.URL)
	if err != nil {
		return "", err
	}
	return url, nil
}

func baseURL(fullURL string) (string, error) {
	idx := strings.LastIndex(fullURL, "/v3")
	if idx == -1 {
		u, err := url.Parse(fullURL)
		if err != nil {
			return "", err
		}
		newURL := url.URL{
			Scheme: u.Scheme,
			Host:   u.Host,
		}
		return newURL.String(), nil
	}
	return fullURL[:idx], nil
}

package config

import (
	"encoding/json"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
)

// Config holds the main config for the user
type Config struct {
	Servers map[string]*ServerConfig
	//Path to the config file
	Path string `json:"path,omitempty"`
	// CurrentServer the user has in focus
	CurrentServer string
}

//ServerConfig holds the config for each server the user has setup
type ServerConfig struct {
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	TokenKey  string `json:"tokenKey"`
	URL       string `json:"url"`
	Project   string `json:"project"`
	CACerts   string `json:"cacert"`
}

func (c Config) Write() error {
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

func (c Config) FocusedServer() *ServerConfig {
	return c.Servers[c.CurrentServer]
}

func (c ServerConfig) FocusedCluster() string {
	return strings.Split(c.Project, ":")[0]
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

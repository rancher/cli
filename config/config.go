package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd/api"
)

var ErrNoConfigurationFound = errors.New("no configuration found, run `login`")

// Config holds the main config for the user
type Config struct {
	Servers map[string]*ServerConfig
	//Path to the config file
	Path string `json:"path,omitempty"`
	// CurrentServer the user has in focus
	CurrentServer string
	// Helper executable to store config
	Helper string `json:"helper,omitempty"`
}

// ServerConfig holds the config for each server the user has setup
type ServerConfig struct {
	AccessKey          string                     `json:"accessKey"`
	SecretKey          string                     `json:"secretKey"`
	TokenKey           string                     `json:"tokenKey"`
	URL                string                     `json:"url"`
	Project            string                     `json:"project"`
	CACerts            string                     `json:"cacert"`
	KubeCredentials    map[string]*ExecCredential `json:"kubeCredentials"`
	KubeConfigs        map[string]*api.Config     `json:"kubeConfigs"`
	ProxyURL           string                     `json:"proxyUrl"`
	HTTPTimeoutSeconds int                        `json:"httpTimeoutSeconds"`
}

func (c *ServerConfig) GetHTTPTimeout() time.Duration {
	return time.Duration(c.HTTPTimeoutSeconds) * time.Second
}

// LoadFromPath attempts to load a config from the given file path. If the file
// doesn't exist, an empty config is returned.
func LoadFromPath(path string) (Config, error) {
	cf := Config{
		Path:    path,
		Servers: make(map[string]*ServerConfig),
		Helper:  "built-in",
	}

	content, err := os.ReadFile(path)
	if err != nil {
		// it's okay if the file is empty, we still return a valid config
		if os.IsNotExist(err) {
			return cf, nil
		}

		return cf, err
	}

	if err := json.Unmarshal(content, &cf); err != nil {
		return cf, fmt.Errorf("unmarshaling %s: %w", path, err)
	}
	cf.Path = path

	return cf, nil
}

// fetch rancher cli config by executing an external helper
func LoadWithHelper(helper string) (Config, error) {
	cf := Config{
		Path:    "",
		Servers: make(map[string]*ServerConfig),
		Helper:  helper,
	}

	cmd := exec.Command(helper, "get")
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	content, err := cmd.Output()
	if err != nil {
		return cf, err
	}

	err = json.Unmarshal(content, &cf)
	return cf, err
}

// GetFilePermissionWarnings returns the following warnings based on the file permission:
// - one warning if the file is group-readable
// - one warning if the file is world-readable
// We want this because configuration may have sensitive information (eg: creds).
// A nil error is returned if the file doesn't exist.
func GetFilePermissionWarnings(path string) ([]string, error) {
	// Permission bits on Windows are not representative
	// of the actual ACLs set for a given file. As the owner
	// bit is always used for all three bits on Windows, this check
	// will always fail, even if the file has restricted access to admins only.
	if runtime.GOOS == "windows" {
		return nil, nil
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return []string{}, fmt.Errorf("get file info: %w", err)
	}

	var warnings []string
	if info.Mode()&0040 > 0 {
		warnings = append(warnings, fmt.Sprintf("Rancher configuration file %s is group-readable. This is insecure.", path))
	}
	if info.Mode()&0004 > 0 {
		warnings = append(warnings, fmt.Sprintf("Rancher configuration file %s is world-readable. This is insecure.", path))
	}
	return warnings, nil
}

func (c Config) Write() error {
	switch c.Helper {
	case "built-in":
		return c.writeNative()
	default:
		// if rancher config was loaded by external helper
		// use the same helper to persist the config
		return c.writeWithHelper()
	}
}

func (c Config) writeNative() error {
	err := os.MkdirAll(filepath.Dir(c.Path), 0700)
	if err != nil {
		return err
	}
	logrus.Infof("Saving config to %s", c.Path)
	p := c.Path
	c.Path = ""
	output, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
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

func (c Config) FocusedServer() (*ServerConfig, error) {
	currentServer, found := c.Servers[c.CurrentServer]
	if !found || currentServer == nil {
		return nil, ErrNoConfigurationFound
	}
	return currentServer, nil
}

func (c ServerConfig) FocusedCluster() string {
	cluster, _, ok := strings.Cut(c.Project, ":")
	if !ok {
		return ""
	}
	return cluster
}

func (c ServerConfig) FocusedProject() string {
	return c.Project
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

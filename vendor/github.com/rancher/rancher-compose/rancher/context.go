package rancher

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/utils"
	rancherClient "github.com/rancher/go-rancher/client"
	rUtils "github.com/rancher/rancher-compose/utils"
	rVersion "github.com/rancher/rancher-compose/version"

	"github.com/hashicorp/go-version"
)

var projectRegexp = regexp.MustCompile("[^a-zA-Z0-9-]")

type Context struct {
	project.Context

	RancherConfig       map[string]RancherConfig
	RancherComposeFile  string
	RancherComposeBytes []byte
	Url                 string
	AccessKey           string
	SecretKey           string
	Client              *rancherClient.RancherClient
	Environment         *rancherClient.Environment
	isOpen              bool
	SidekickInfo        *SidekickInfo
	Uploader            Uploader
	PullCached          bool
	Pull                bool
	Args                []string

	Upgrade        bool
	ForceUpgrade   bool
	Rollback       bool
	Interval       int64
	BatchSize      int64
	ConfirmUpgrade bool
}

type RancherConfig struct {
	// VirtualMachine fields
	Vcpu     int64                              `yaml:"vcpu,omitempty"`
	Userdata string                             `yaml:"userdata,omitempty"`
	Memory   int64                              `yaml:"memory,omitempty"`
	Disks    []rancherClient.VirtualMachineDisk `yaml:"disks,omitempty"`

	Type               string                                 `yaml:"type,omitempty"`
	Scale              int                                    `yaml:"scale,omitempty"`
	RetainIp           bool                                   `yaml:"retain_ip,omitempty"`
	LoadBalancerConfig *rancherClient.LoadBalancerConfig      `yaml:"load_balancer_config,omitempty"`
	ExternalIps        []string                               `yaml:"external_ips,omitempty"`
	Hostname           string                                 `yaml:"hostname,omitempty"`
	HealthCheck        *rancherClient.InstanceHealthCheck     `yaml:"health_check,omitempty"`
	DefaultCert        string                                 `yaml:"default_cert,omitempty"`
	Certs              []string                               `yaml:"certs,omitempty"`
	Metadata           map[string]interface{}                 `yaml:"metadata,omitempty"`
	ScalePolicy        *rancherClient.ScalePolicy             `yaml:"scale_policy,omitempty"`
	ServiceSchemas     map[string]rancherClient.Schema        `yaml:"service_schemas,omitempty"`
	UpgradeStrategy    rancherClient.InServiceUpgradeStrategy `yaml:"upgrade_strategy,omitempty"`
}

func ResolveRancherCompose(composeFile, rancherComposeFile string) (string, error) {
	if rancherComposeFile == "" && composeFile != "" {
		f, err := filepath.Abs(composeFile)
		if err != nil {
			return "", err
		}

		return path.Join(path.Dir(f), "rancher-compose.yml"), nil
	}

	return rancherComposeFile, nil
}

func (c *Context) readRancherConfig() error {
	if c.RancherComposeBytes == nil {
		var err error
		c.RancherComposeFile, err = ResolveRancherCompose(c.ComposeFiles[0], c.RancherComposeFile)
		if err != nil {
			return err
		}
	}

	if c.RancherComposeBytes == nil {
		logrus.Debugf("Opening rancher-compose file: %s", c.RancherComposeFile)
		if composeBytes, err := ioutil.ReadFile(c.RancherComposeFile); os.IsNotExist(err) {
			logrus.Debugf("Not found: %s", c.RancherComposeFile)
		} else if err != nil {
			logrus.Errorf("Failed to open %s", c.RancherComposeFile)
			return err
		} else {
			c.RancherComposeBytes = composeBytes
		}
	}

	return c.unmarshalBytes(c.ComposeBytes[0], c.RancherComposeBytes)
}

func (c *Context) unmarshalBytes(composeBytes, bytes []byte) error {
	rawServiceMap := config.RawServiceMap{}
	if composeBytes != nil {
		if err := yaml.Unmarshal(composeBytes, &rawServiceMap); err != nil {
			return err
		}

		for key := range rawServiceMap {
			delete(rawServiceMap[key], "hostname")
		}
	}
	if bytes != nil {
		if err := yaml.Unmarshal(bytes, &rawServiceMap); err != nil {
			return err
		}
	}
	if err := config.Interpolate(c.EnvironmentLookup, &rawServiceMap); err != nil {
		return err
	}
	if err := utils.Convert(rawServiceMap, &c.RancherConfig); err != nil {
		return err
	}
	for _, v := range c.RancherConfig {
		rUtils.RemoveInterfaceKeys(v.Metadata)
	}
	return nil
}

func (c *Context) sanitizedProjectName() string {
	projectName := projectRegexp.ReplaceAllString(strings.ToLower(c.ProjectName), "-")

	if len(projectName) > 0 && strings.ContainsAny(projectName[0:1], "_.-") {
		projectName = "x" + projectName
	}

	return projectName
}

func (c *Context) loadClient() (*rancherClient.RancherClient, error) {
	if c.Client == nil {
		if c.Url == "" {
			return nil, fmt.Errorf("RANCHER_URL is not set")
		}

		if client, err := rancherClient.NewRancherClient(&rancherClient.ClientOpts{
			Url:       c.Url,
			AccessKey: c.AccessKey,
			SecretKey: c.SecretKey,
		}); err != nil {
			return nil, err
		} else {
			c.Client = client
		}
	}

	return c.Client, nil
}

func (c *Context) open() error {
	if c.isOpen {
		return nil
	}

	c.ProjectName = c.sanitizedProjectName()

	if err := c.readRancherConfig(); err != nil {
		return err
	}

	if _, err := c.loadClient(); err != nil {
		return err
	}

	if envSchema, ok := c.Client.Types["environment"]; !ok || !rUtils.Contains(envSchema.CollectionMethods, "POST") {
		return fmt.Errorf("Can not create a stack, check API key [%s] for [%s]", c.AccessKey, c.Url)
	}

	c.checkVersion()

	if _, err := c.LoadEnv(); err != nil {
		return err
	}

	c.isOpen = true
	return nil
}

func (c *Context) checkVersion() {
	// We don't care about errors from this code
	newVersion := c.getSetting("rancher.compose.version")
	if len(newVersion) <= 1 && len(rVersion.VERSION) <= 1 {
		return
	}

	current, err := version.NewVersion(strings.TrimLeft(rVersion.VERSION, "v"))
	if err != nil {
		return
	}

	// strip out beta/ from string
	parts := strings.SplitN(newVersion, "v", 2)
	if len(parts) == 2 {
		newVersion = parts[1]
	}

	toCheck, err := version.NewVersion(newVersion)
	if err != nil {
		return
	}

	if toCheck.GreaterThan(current) {
		logrus.Warnf("A newer version of rancher-compose is available: %s", newVersion)
	}
}

func (c *Context) getSetting(key string) string {
	s, err := c.Client.Setting.ById(key)
	if err != nil || s == nil {
		return ""
	}
	return s.Value
}

func (c *Context) LoadEnv() (*rancherClient.Environment, error) {
	if c.Environment != nil {
		return c.Environment, nil
	}

	projectName := c.sanitizedProjectName()
	if _, err := c.loadClient(); err != nil {
		return nil, err
	}

	logrus.Debugf("Looking for stack %s", projectName)
	// First try by name
	envs, err := c.Client.Environment.List(&rancherClient.ListOpts{
		Filters: map[string]interface{}{
			"name":         projectName,
			"removed_null": nil,
		},
	})
	if err != nil {
		return nil, err
	}

	for _, env := range envs.Data {
		if strings.EqualFold(projectName, env.Name) {
			logrus.Debugf("Found stack: %s(%s)", env.Name, env.Id)
			c.Environment = &env
			return c.Environment, nil
		}
	}

	// Now try not by name for case sensitive databases
	envs, err = c.Client.Environment.List(&rancherClient.ListOpts{
		Filters: map[string]interface{}{
			"removed_null": nil,
		},
	})
	if err != nil {
		return nil, err
	}

	for _, env := range envs.Data {
		if strings.EqualFold(projectName, env.Name) {
			logrus.Debugf("Found stack: %s(%s)", env.Name, env.Id)
			c.Environment = &env
			return c.Environment, nil
		}
	}

	logrus.Infof("Creating stack %s", projectName)
	env, err := c.Client.Environment.Create(&rancherClient.Environment{
		Name: projectName,
	})
	if err != nil {
		return nil, err
	}

	c.Environment = env

	return c.Environment, nil
}

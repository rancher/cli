package rancher

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/utils"
	composeYaml "github.com/docker/libcompose/yaml"
	"github.com/fatih/structs"
	legacyClient "github.com/rancher/go-rancher/client"
	"github.com/rancher/go-rancher/v2"
	"github.com/rancher/rancher-compose/preprocess"
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
	BindingsFile        string
	Binding             *client.Binding
	BindingsBytes       []byte
	Url                 string
	AccessKey           string
	SecretKey           string
	Client              *client.RancherClient
	Stack               *client.Stack
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

type PortRule struct {
	SourcePort  int    `json:"source_port" yaml:"source_port"`
	Protocol    string `json:"protocol" yaml:"protocol"`
	Path        string `json:"path" yaml:"path"`
	Hostname    string `json:"hostname" yaml:"hostname"`
	Service     string `json:"service" yaml:"service"`
	TargetPort  int    `json:"target_port" yaml:"target_port"`
	Priority    int    `json:"priority" yaml:"priority"`
	BackendName string `json:"backend_name" yaml:"backend_name"`
	Selector    string `json:"selector" yaml:"selector"`
}

type LBConfig struct {
	Certs            []string            `json:"certs" yaml:"certs"`
	DefaultCert      string              `json:"default_cert" yaml:"default_cert"`
	PortRules        []PortRule          `json:"port_rules" yaml:"port_rules"`
	Config           string              `json:"config" yaml:"config"`
	StickinessPolicy *LBStickinessPolicy `json:"stickiness_policy" yaml:"stickiness_policy"`
}

type LBStickinessPolicy struct {
	Name     string `json:"name" yaml:"name"`
	Cookie   string `json:"cookie" yaml:"cookie"`
	Domain   string `json:"domain" yaml:"domain"`
	Indirect bool   `json:"indirect" yaml:"indirect"`
	Nocache  bool   `json:"nocache" yaml:"nocache"`
	Postonly bool   `json:"postonly" yaml:"postonly"`
	Mode     string `json:"mode" yaml:"mode"`
}

type RancherConfig struct {
	// VirtualMachine fields
	Vcpu     composeYaml.StringorInt     `yaml:"vcpu,omitempty"`
	Userdata string                      `yaml:"userdata,omitempty"`
	Memory   composeYaml.StringorInt     `yaml:"memory,omitempty"`
	Disks    []client.VirtualMachineDisk `yaml:"disks,omitempty"`

	Type        string                      `yaml:"type,omitempty"`
	Scale       composeYaml.StringorInt     `yaml:"scale,omitempty"`
	RetainIp    bool                        `yaml:"retain_ip,omitempty"`
	LbConfig    *LBConfig                   `yaml:"lb_config,omitempty"`
	ExternalIps []string                    `yaml:"external_ips,omitempty"`
	Hostname    string                      `yaml:"hostname,omitempty"`
	HealthCheck *client.InstanceHealthCheck `yaml:"health_check,omitempty"`

	// Present only for compatibility with legacy load balancers
	// New load balancers will have these fields under 'lb_config'
	LegacyLoadBalancerConfig *legacyClient.LoadBalancerConfig `yaml:"load_balancer_config,omitempty"`
	DefaultCert              string                           `yaml:"default_cert,omitempty"`
	Certs                    []string                         `yaml:"certs,omitempty"`

	Metadata        map[string]interface{}          `yaml:"metadata,omitempty"`
	ScalePolicy     *client.ScalePolicy             `yaml:"scale_policy,omitempty"`
	ServiceSchemas  map[string]client.Schema        `yaml:"service_schemas,omitempty"`
	UpgradeStrategy client.InServiceUpgradeStrategy `yaml:"upgrade_strategy,omitempty"`
	StorageDriver   *client.StorageDriver           `yaml:"storage_driver,omitempty"`
	NetworkDriver   *client.NetworkDriver           `yaml:"network_driver,omitempty"`
}

func getRancherConfigObjects() map[string]bool {
	rancherConfig := structs.New(RancherConfig{})
	fields := map[string]bool{}
	for _, field := range rancherConfig.Fields() {
		kind := field.Kind().String()
		if kind == "struct" || kind == "ptr" || kind == "slice" {
			split := strings.Split(field.Tag("yaml"), ",")
			fields[split[0]] = true
		}
	}
	return fields
}

type BindingProperty struct {
	Services map[string]Service `json:"services"`
}

type Service struct {
	Labels map[string]interface{} `json:"labels"`
	Ports  []interface{}          `json:"ports"`
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
		config, err := config.CreateConfig(composeBytes)
		if err != nil {
			return err
		}
		rawServiceMap = config.Services
		for key := range rawServiceMap {
			delete(rawServiceMap[key], "hostname")
		}
	}
	if bytes != nil && len(bytes) > 0 {
		config, err := config.CreateConfig(bytes)
		if err != nil {
			return err
		}
		rawServiceMap = config.Services
	}
	return c.fillInRancherConfig(rawServiceMap)
}

func (c *Context) fillInRancherConfig(rawServiceMap config.RawServiceMap) error {
	if err := config.InterpolateRawServiceMap(&rawServiceMap, c.EnvironmentLookup); err != nil {
		return err
	}

	rawServiceMap, err := preprocess.TryConvertStringsToInts(rawServiceMap, getRancherConfigObjects())
	if err != nil {
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

func (c *Context) loadClient() (*client.RancherClient, error) {
	if c.Client == nil {
		if c.Url == "" {
			return nil, fmt.Errorf("RANCHER_URL is not set")
		}

		if client, err := client.NewRancherClient(&client.ClientOpts{
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

	if stackSchema, ok := c.Client.GetTypes()["stack"]; !ok || !rUtils.Contains(stackSchema.CollectionMethods, "POST") {
		return fmt.Errorf("Can not create a stack, check API key [%s] for [%s]", c.AccessKey, c.Url)
	}

	c.checkVersion()

	if _, err := c.LoadStack(); err != nil {
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

func (c *Context) LoadStack() (*client.Stack, error) {
	if c.Stack != nil {
		return c.Stack, nil
	}

	projectName := c.sanitizedProjectName()
	if _, err := c.loadClient(); err != nil {
		return nil, err
	}

	logrus.Debugf("Looking for stack %s", projectName)
	// First try by name
	stacks, err := c.Client.Stack.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"name":         projectName,
			"removed_null": nil,
		},
	})
	if err != nil {
		return nil, err
	}

	for _, stack := range stacks.Data {
		if strings.EqualFold(projectName, stack.Name) {
			logrus.Debugf("Found stack: %s(%s)", stack.Name, stack.Id)
			c.Stack = &stack
			return c.Stack, nil
		}
	}

	// Now try not by name for case sensitive databases
	stacks, err = c.Client.Stack.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"removed_null": nil,
		},
	})
	if err != nil {
		return nil, err
	}

	for _, stack := range stacks.Data {
		if strings.EqualFold(projectName, stack.Name) {
			logrus.Debugf("Found stack: %s(%s)", stack.Name, stack.Id)
			c.Stack = &stack
			return c.Stack, nil
		}
	}

	logrus.Infof("Creating stack %s", projectName)
	stack, err := c.Client.Stack.Create(&client.Stack{
		Name: projectName,
	})
	if err != nil {
		return nil, err
	}

	c.Stack = stack

	return c.Stack, nil
}

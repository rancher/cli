package config

import (
	"sync"

	legacyClient "github.com/rancher/go-rancher/client"
	"github.com/rancher/go-rancher/v2"
	"github.com/rancher/rancher-compose-executor/yaml"
)

// EnvironmentLookup defines methods to provides environment variable loading.
type EnvironmentLookup interface {
	Lookup(key string, config *ServiceConfig) []string
	Variables() map[string]string
}

// ResourceLookup defines methods to provides file loading.
type ResourceLookup interface {
	Lookup(file, relativeTo string) ([]byte, string, error)
	ResolvePath(path, inFile string) string
}

// ServiceConfigV1 holds version 1 of libcompose service configuration
type ServiceConfigV1 struct {
	BlkioWeight       yaml.StringorInt     `yaml:"blkio_weight,omitempty"`
	BlkioWeightDevice []string             `yaml:"blkio_weight_device,omitempty"`
	Build             string               `yaml:"build,omitempty"`
	CapAdd            []string             `yaml:"cap_add,omitempty"`
	CapDrop           []string             `yaml:"cap_drop,omitempty"`
	CgroupParent      string               `yaml:"cgroup_parent,omitempty"`
	CPUPeriod         yaml.StringorInt     `yaml:"cpu_period,omitempty"`
	CPUQuota          yaml.StringorInt     `yaml:"cpu_quota,omitempty"`
	CPUSet            string               `yaml:"cpuset,omitempty"`
	CPUShares         yaml.StringorInt     `yaml:"cpu_shares,omitempty"`
	Command           yaml.Command         `yaml:"command,flow,omitempty"`
	ContainerName     string               `yaml:"container_name,omitempty"`
	DeviceReadBps     yaml.MaporColonSlice `yaml:"device_read_bps,omitempty"`
	DeviceReadIOps    yaml.MaporColonSlice `yaml:"device_read_iops,omitempty"`
	Devices           []string             `yaml:"devices,omitempty"`
	DeviceWriteBps    yaml.MaporColonSlice `yaml:"device_write_bps,omitempty"`
	DeviceWriteIOps   yaml.MaporColonSlice `yaml:"device_write_iops,omitempty"`
	DNS               yaml.Stringorslice   `yaml:"dns,omitempty"`
	DNSOpt            []string             `yaml:"dns_opt,omitempty"`
	DNSSearch         yaml.Stringorslice   `yaml:"dns_search,omitempty"`
	Dockerfile        string               `yaml:"dockerfile,omitempty"`
	DomainName        string               `yaml:"domainname,omitempty"`
	Entrypoint        yaml.Command         `yaml:"entrypoint,flow,omitempty"`
	EnvFile           yaml.Stringorslice   `yaml:"env_file,omitempty"`
	Environment       yaml.MaporEqualSlice `yaml:"environment,omitempty"`
	GroupAdd          []string             `yaml:"group_add,omitempty"`
	Hostname          string               `yaml:"hostname,omitempty"`
	Image             string               `yaml:"image,omitempty"`
	Isolation         string               `yaml:"isolation,omitempty"`
	Labels            yaml.SliceorMap      `yaml:"labels,omitempty"`
	Links             yaml.MaporColonSlice `yaml:"links,omitempty"`
	LogDriver         string               `yaml:"log_driver,omitempty"`
	MacAddress        string               `yaml:"mac_address,omitempty"`
	MemLimit          yaml.MemStringorInt  `yaml:"mem_limit,omitempty"`
	MemSwapLimit      yaml.MemStringorInt  `yaml:"memswap_limit,omitempty"`
	MemSwappiness     yaml.StringorInt     `yaml:"mem_swappiness,omitempty"`
	Name              string               `yaml:"name,omitempty"`
	Net               string               `yaml:"net,omitempty"`
	OomKillDisable    bool                 `yaml:"oom_kill_disable,omitempty"`
	OomScoreAdj       yaml.StringorInt     `yaml:"oom_score_adj,omitempty"`
	Pid               string               `yaml:"pid,omitempty"`
	Uts               string               `yaml:"uts,omitempty"`
	Ipc               string               `yaml:"ipc,omitempty"`
	Ports             []string             `yaml:"ports,omitempty"`
	Privileged        bool                 `yaml:"privileged,omitempty"`
	Restart           string               `yaml:"restart,omitempty"`
	ReadOnly          bool                 `yaml:"read_only,omitempty"`
	Secrets           SecretReferences     `yaml:"secrets,omitempty"`
	ShmSize           yaml.MemStringorInt  `yaml:"shm_size,omitempty"`
	StdinOpen         bool                 `yaml:"stdin_open,omitempty"`
	SecurityOpt       []string             `yaml:"security_opt,omitempty"`
	StopSignal        string               `yaml:"stop_signal,omitempty"`
	Sysctls           yaml.SliceorMap      `yaml:"sysctls,omitempty"`
	Init              bool                 `yaml:"init,omitempty"`
	Tmpfs             yaml.Stringorslice   `yaml:"tmpfs,omitempty"`
	Tty               bool                 `yaml:"tty,omitempty"`
	User              string               `yaml:"user,omitempty"`
	VolumeDriver      string               `yaml:"volume_driver,omitempty"`
	Volumes           []string             `yaml:"volumes,omitempty"`
	VolumesFrom       []string             `yaml:"volumes_from,omitempty"`
	WorkingDir        string               `yaml:"working_dir,omitempty"`
	Expose            []string             `yaml:"expose,omitempty"`
	ExternalLinks     []string             `yaml:"external_links,omitempty"`
	LogOpt            map[string]string    `yaml:"log_opt,omitempty"`
	ExtraHosts        []string             `yaml:"extra_hosts,omitempty"`
	Ulimits           yaml.Ulimits         `yaml:"ulimits,omitempty"`

	LbConfig                 *LBConfig                        `yaml:"lb_config"`
	LegacyLoadBalancerConfig *legacyClient.LoadBalancerConfig `yaml:"load_balancer_config,omitempty"`
	DefaultCert              string                           `yaml:"default_cert,omitempty"`
	Certs                    []string                         `yaml:"certs,omitempty"`

	Vcpu     yaml.StringorInt            `yaml:"vcpu,omitempty"`
	Userdata string                      `yaml:"userdata,omitempty"`
	Memory   yaml.MemStringorInt         `yaml:"memory,omitempty"`
	Disks    []client.VirtualMachineDisk `yaml:"disks,omitempty"`

	Type        string           `yaml:"type,omitempty"`
	Scale       yaml.StringorInt `yaml:"scale,omitempty"`
	RetainIp    bool             `yaml:"retain_ip,omitempty"`
	ExternalIps []string         `yaml:"external_ips,omitempty"`
	// TODO: hostname is in docker-compose.yml and rancher-compose.yml
	//Hostname    string                      `yaml:"hostname,omitempty"`
	HealthCheck *client.InstanceHealthCheck `yaml:"health_check,omitempty"`

	Metadata        map[string]interface{}          `yaml:"metadata,omitempty"`
	ScalePolicy     *client.ScalePolicy             `yaml:"scale_policy,omitempty"`
	ServiceSchemas  map[string]client.Schema        `yaml:"service_schemas,omitempty"`
	UpgradeStrategy client.InServiceUpgradeStrategy `yaml:"upgrade_strategy,omitempty"`
	StorageDriver   *client.StorageDriver           `yaml:"storage_driver,omitempty"`
	NetworkDriver   *client.NetworkDriver           `yaml:"network_driver,omitempty"`
}

// Log holds v2 logging information
type Log struct {
	Driver  string            `yaml:"driver,omitempty"`
	Options map[string]string `yaml:"options,omitempty"`
}

// ServiceConfig holds version 2 of libcompose service configuration
type ServiceConfig struct {
	BlkioWeight       yaml.StringorInt     `yaml:"blkio_weight,omitempty"`
	BlkioWeightDevice []string             `yaml:"blkio_weight_device,omitempty"`
	Build             yaml.Build           `yaml:"build,omitempty"`
	CapAdd            []string             `yaml:"cap_add,omitempty"`
	CapDrop           []string             `yaml:"cap_drop,omitempty"`
	CPUPeriod         yaml.StringorInt     `yaml:"cpu_period,omitempty"`
	CPUSet            string               `yaml:"cpuset,omitempty"`
	CPUShares         yaml.StringorInt     `yaml:"cpu_shares,omitempty"`
	CPUQuota          yaml.StringorInt     `yaml:"cpu_quota,omitempty"`
	Command           yaml.Command         `yaml:"command,flow,omitempty"`
	CgroupParent      string               `yaml:"cgroup_parent,omitempty"`
	ContainerName     string               `yaml:"container_name,omitempty"`
	DeviceReadBps     yaml.MaporColonSlice `yaml:"device_read_bps,omitempty"`
	DeviceReadIOps    yaml.MaporColonSlice `yaml:"device_read_iops,omitempty"`
	Devices           []string             `yaml:"devices,omitempty"`
	DeviceWriteBps    yaml.MaporColonSlice `yaml:"device_write_bps,omitempty"`
	DeviceWriteIOps   yaml.MaporColonSlice `yaml:"device_write_iops,omitempty"`
	DependsOn         []string             `yaml:"depends_on,omitempty"`
	DNS               yaml.Stringorslice   `yaml:"dns,omitempty"`
	DNSOpt            []string             `yaml:"dns_opt,omitempty"`
	DNSSearch         yaml.Stringorslice   `yaml:"dns_search,omitempty"`
	DomainName        string               `yaml:"domainname,omitempty"`
	Entrypoint        yaml.Command         `yaml:"entrypoint,flow,omitempty"`
	EnvFile           yaml.Stringorslice   `yaml:"env_file,omitempty"`
	Environment       yaml.MaporEqualSlice `yaml:"environment,omitempty"`
	Expose            []string             `yaml:"expose,omitempty"`
	Extends           yaml.MaporEqualSlice `yaml:"extends,omitempty"`
	ExternalLinks     []string             `yaml:"external_links,omitempty"`
	ExtraHosts        []string             `yaml:"extra_hosts,omitempty"`
	GroupAdd          []string             `yaml:"group_add,omitempty"`
	Image             string               `yaml:"image,omitempty"`
	Init              bool                 `yaml:"init,omitempty"`
	Isolation         string               `yaml:"isolation,omitempty"`
	Hostname          string               `yaml:"hostname,omitempty"`
	Ipc               string               `yaml:"ipc,omitempty"`
	Labels            yaml.SliceorMap      `yaml:"labels,omitempty"`
	Links             yaml.MaporColonSlice `yaml:"links,omitempty"`
	Logging           Log                  `yaml:"logging,omitempty"`
	MacAddress        string               `yaml:"mac_address,omitempty"`
	MemLimit          yaml.MemStringorInt  `yaml:"mem_limit,omitempty"`
	MemReservation    yaml.MemStringorInt  `yaml:"mem_reservation,omitempty"`
	MemSwapLimit      yaml.MemStringorInt  `yaml:"memswap_limit,omitempty"`
	MemSwappiness     yaml.StringorInt     `yaml:"mem_swappiness,omitempty"`
	NetworkMode       string               `yaml:"network_mode,omitempty"`
	Networks          *yaml.Networks       `yaml:"networks,omitempty"`
	OomKillDisable    bool                 `yaml:"oom_kill_disable,omitempty"`
	OomScoreAdj       yaml.StringorInt     `yaml:"oom_score_adj,omitempty"`
	Pid               string               `yaml:"pid,omitempty"`
	Ports             []string             `yaml:"ports,omitempty"`
	Privileged        bool                 `yaml:"privileged,omitempty"`
	Secrets           SecretReferences     `yaml:"secrets,omitempty"`
	SecurityOpt       []string             `yaml:"security_opt,omitempty"`
	ShmSize           yaml.MemStringorInt  `yaml:"shm_size,omitempty"`
	StopSignal        string               `yaml:"stop_signal,omitempty"`
	Sysctls           yaml.SliceorMap      `yaml:"sysctls,omitempty"`
	Tmpfs             yaml.Stringorslice   `yaml:"tmpfs,omitempty"`
	VolumeDriver      string               `yaml:"volume_driver,omitempty"`
	Volumes           *yaml.Volumes        `yaml:"volumes,omitempty"`
	VolumesFrom       []string             `yaml:"volumes_from,omitempty"`
	Uts               string               `yaml:"uts,omitempty"`
	Restart           string               `yaml:"restart,omitempty"`
	ReadOnly          bool                 `yaml:"read_only,omitempty"`
	StdinOpen         bool                 `yaml:"stdin_open,omitempty"`
	Tty               bool                 `yaml:"tty,omitempty"`
	User              string               `yaml:"user,omitempty"`
	WorkingDir        string               `yaml:"working_dir,omitempty"`
	Ulimits           yaml.Ulimits         `yaml:"ulimits,omitempty"`

	RancherConfig `yaml:",inline"`
}

type RancherConfig struct {
	LbConfig                 *LBConfig                        `yaml:"lb_config"`
	LegacyLoadBalancerConfig *legacyClient.LoadBalancerConfig `yaml:"load_balancer_config,omitempty"`
	DefaultCert              string                           `yaml:"default_cert,omitempty"`
	Certs                    []string                         `yaml:"certs,omitempty"`

	Vcpu     yaml.StringorInt            `yaml:"vcpu,omitempty"`
	Userdata string                      `yaml:"userdata,omitempty"`
	Memory   yaml.MemStringorInt         `yaml:"memory,omitempty"`
	Disks    []client.VirtualMachineDisk `yaml:"disks,omitempty"`

	Type        string                      `yaml:"type,omitempty"`
	Scale       yaml.StringorInt            `yaml:"scale,omitempty"`
	RetainIp    bool                        `yaml:"retain_ip,omitempty"`
	ExternalIps []string                    `yaml:"external_ips,omitempty"`
	HealthCheck *client.InstanceHealthCheck `yaml:"health_check,omitempty"`

	Metadata        map[string]interface{}          `yaml:"metadata,omitempty"`
	ScalePolicy     *client.ScalePolicy             `yaml:"scale_policy,omitempty"`
	ServiceSchemas  map[string]client.Schema        `yaml:"service_schemas,omitempty"`
	UpgradeStrategy client.InServiceUpgradeStrategy `yaml:"upgrade_strategy,omitempty"`
	StorageDriver   *client.StorageDriver           `yaml:"storage_driver,omitempty"`
	NetworkDriver   *client.NetworkDriver           `yaml:"network_driver,omitempty"`
}

// TODO: json tags needed?
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

type LBStickinessPolicy struct {
	Name     string `json:"name" yaml:"name"`
	Cookie   string `json:"cookie" yaml:"cookie"`
	Domain   string `json:"domain" yaml:"domain"`
	Indirect bool   `json:"indirect" yaml:"indirect"`
	Nocache  bool   `json:"nocache" yaml:"nocache"`
	Postonly bool   `json:"postonly" yaml:"postonly"`
	Mode     string `json:"mode" yaml:"mode"`
}

type LBConfig struct {
	Certs            []string            `json:"certs" yaml:"certs"`
	DefaultCert      string              `json:"default_cert" yaml:"default_cert"`
	PortRules        []PortRule          `json:"port_rules" yaml:"port_rules"`
	Config           string              `json:"config" yaml:"config"`
	StickinessPolicy *LBStickinessPolicy `json:"stickiness_policy" yaml:"stickiness_policy"`
}

// VolumeConfig holds v2 volume configuration
type VolumeConfig struct {
	Driver       string            `yaml:"driver,omitempty"`
	DriverOpts   map[string]string `yaml:"driver_opts,omitempty"`
	External     yaml.External     `yaml:"external,omitempty"`
	PerContainer bool              `yaml:"per_container,omitempty"`
}

// Ipam holds v2 network IPAM information
type Ipam struct {
	Driver string       `yaml:"driver,omitempty"`
	Config []IpamConfig `yaml:"config,omitempty"`
}

// IpamConfig holds v2 network IPAM configuration information
type IpamConfig struct {
	Subnet     string            `yaml:"subnet,omitempty"`
	IPRange    string            `yaml:"ip_range,omitempty"`
	Gateway    string            `yaml:"gateway,omitempty"`
	AuxAddress map[string]string `yaml:"aux_addresses,omitempty"`
}

// NetworkConfig holds v2 network configuration
type NetworkConfig struct {
	Driver     string            `yaml:"driver,omitempty"`
	DriverOpts map[string]string `yaml:"driver_opts,omitempty"`
	External   yaml.External     `yaml:"external,omitempty"`
	Ipam       Ipam              `yaml:"ipam,omitempty"`
}

type SecretConfig struct {
	File     string `yaml:"file,omitempty"`
	External string `yaml:"external,omitempty"`
}

type HostConfig struct {
	Count    int    `yaml:"count,omitempty"`
	Template string `yaml:"template,omitempty"`
	// Fancy trick to catch any other fields
	Dynamic map[string]interface{} `yaml:",inline"`
}

type DependencyConfig struct {
	Name     string `yaml:"name,omitempty"`
	Template string `yaml:"template,omitempty"`
	Version  string `yaml:"version,omitempty"`
	// TODO: additional answers/environment?
}

type RawConfig struct {
	Version string `yaml:"version,omitempty"`

	Services         RawServiceMap `yaml:"services,omitempty"`
	Containers       RawServiceMap `yaml:"containers,omitempty"`
	LoadBalancers    RawServiceMap `yaml:"load_balancers,omitempty"`
	StorageDrivers   RawServiceMap `yaml:"storage_drivers,omitempty"`
	NetworkDrivers   RawServiceMap `yaml:"network_drivers,omitempty"`
	VirtualMachines  RawServiceMap `yaml:"virtual_machines,omitempty"`
	ExternalServices RawServiceMap `yaml:"external_services,omitempty"`
	Aliases          RawServiceMap `yaml:"aliases,omitempty"`

	Dependencies map[string]interface{} `yaml:"dependencies,omitempty"`
	Volumes      map[string]interface{} `yaml:"volumes,omitempty"`
	Networks     map[string]interface{} `yaml:"networks,omitempty"`
	Secrets      map[string]interface{} `yaml:"secrets,omitempty"`
	Hosts        map[string]interface{} `yaml:"hosts,omitempty"`
}

type Config struct {
	Services     map[string]*ServiceConfig    `yaml:"services,omitempty"`
	Containers   map[string]*ServiceConfig    `yaml:"containers,omitempty"`
	Dependencies map[string]*DependencyConfig `yaml:"dependencies,omitempty"`
	Volumes      map[string]*VolumeConfig     `yaml:"volumes,omitempty"`
	Networks     map[string]*NetworkConfig    `yaml:"networks,omitempty"`
	Secrets      map[string]*SecretConfig     `yaml:"secrets,omitempty"`
	Hosts        map[string]*HostConfig       `yaml:"hosts,omitempty"`
}

// NewServiceConfigs initializes a new Configs struct
func NewServiceConfigs() *ServiceConfigs {
	return &ServiceConfigs{
		m: make(map[string]*ServiceConfig),
	}
}

// ServiceConfigs holds a concurrent safe map of ServiceConfig
type ServiceConfigs struct {
	m  map[string]*ServiceConfig
	mu sync.RWMutex
}

// Has checks if the config map has the specified name
func (c *ServiceConfigs) Has(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.m[name]
	return ok
}

// Get returns the config and the presence of the specified name
func (c *ServiceConfigs) Get(name string) (*ServiceConfig, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	service, ok := c.m[name]
	return service, ok
}

// Add add the specifed config with the specified name
func (c *ServiceConfigs) Add(name string, service *ServiceConfig) {
	c.mu.Lock()
	c.m[name] = service
	c.mu.Unlock()
}

// Remove removes the config with the specified name
func (c *ServiceConfigs) Remove(name string) {
	c.mu.Lock()
	delete(c.m, name)
	c.mu.Unlock()
}

// Len returns the len of the configs
func (c *ServiceConfigs) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.m)
}

// Keys returns the names of the config
func (c *ServiceConfigs) Keys() []string {
	keys := []string{}
	c.mu.RLock()
	defer c.mu.RUnlock()
	for name := range c.m {
		keys = append(keys, name)
	}
	return keys
}

// All returns all the config at once
func (c *ServiceConfigs) All() map[string]*ServiceConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m
}

// RawService is represent a Service in map form unparsed
type RawService map[string]interface{}

// RawServiceMap is a collection of RawServices
type RawServiceMap map[string]RawService

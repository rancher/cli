package client

const (
	LAUNCH_CONFIG_TYPE = "launchConfig"
)

type LaunchConfig struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	AgentId string `json:"agentId,omitempty" yaml:"agent_id,omitempty"`

	BlkioDeviceOptions map[string]interface{} `json:"blkioDeviceOptions,omitempty" yaml:"blkio_device_options,omitempty"`

	BlkioWeight int64 `json:"blkioWeight,omitempty" yaml:"blkio_weight,omitempty"`

	CapAdd []string `json:"capAdd,omitempty" yaml:"cap_add,omitempty"`

	CapDrop []string `json:"capDrop,omitempty" yaml:"cap_drop,omitempty"`

	CgroupParent string `json:"cgroupParent,omitempty" yaml:"cgroup_parent,omitempty"`

	ClusterId string `json:"clusterId,omitempty" yaml:"cluster_id,omitempty"`

	Command []string `json:"command,omitempty" yaml:"command,omitempty"`

	CompleteUpdate bool `json:"completeUpdate,omitempty" yaml:"complete_update,omitempty"`

	Count int64 `json:"count,omitempty" yaml:"count,omitempty"`

	CpuCount int64 `json:"cpuCount,omitempty" yaml:"cpu_count,omitempty"`

	CpuPercent int64 `json:"cpuPercent,omitempty" yaml:"cpu_percent,omitempty"`

	CpuPeriod int64 `json:"cpuPeriod,omitempty" yaml:"cpu_period,omitempty"`

	CpuQuota int64 `json:"cpuQuota,omitempty" yaml:"cpu_quota,omitempty"`

	CpuSetCpu string `json:"cpuSetCpu,omitempty" yaml:"cpu_set_cpu,omitempty"`

	CpuSetMems string `json:"cpuSetMems,omitempty" yaml:"cpu_set_mems,omitempty"`

	CpuShares int64 `json:"cpuShares,omitempty" yaml:"cpu_shares,omitempty"`

	CreateIndex int64 `json:"createIndex,omitempty" yaml:"create_index,omitempty"`

	CreateOnly bool `json:"createOnly,omitempty" yaml:"create_only,omitempty"`

	Created string `json:"created,omitempty" yaml:"created,omitempty"`

	DataVolumeMounts map[string]interface{} `json:"dataVolumeMounts,omitempty" yaml:"data_volume_mounts,omitempty"`

	DataVolumes []string `json:"dataVolumes,omitempty" yaml:"data_volumes,omitempty"`

	DataVolumesFrom []string `json:"dataVolumesFrom,omitempty" yaml:"data_volumes_from,omitempty"`

	DependsOn []DependsOn `json:"dependsOn,omitempty" yaml:"depends_on,omitempty"`

	DeploymentUnitId string `json:"deploymentUnitId,omitempty" yaml:"deployment_unit_id,omitempty"`

	DeploymentUnitUuid string `json:"deploymentUnitUuid,omitempty" yaml:"deployment_unit_uuid,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	Desired bool `json:"desired,omitempty" yaml:"desired,omitempty"`

	Devices []string `json:"devices,omitempty" yaml:"devices,omitempty"`

	DiskQuota int64 `json:"diskQuota,omitempty" yaml:"disk_quota,omitempty"`

	Dns []string `json:"dns,omitempty" yaml:"dns,omitempty"`

	DnsOpt []string `json:"dnsOpt,omitempty" yaml:"dns_opt,omitempty"`

	DnsSearch []string `json:"dnsSearch,omitempty" yaml:"dns_search,omitempty"`

	DomainName string `json:"domainName,omitempty" yaml:"domain_name,omitempty"`

	EntryPoint []string `json:"entryPoint,omitempty" yaml:"entry_point,omitempty"`

	Environment map[string]string `json:"environment,omitempty" yaml:"environment,omitempty"`

	ExitCode int64 `json:"exitCode,omitempty" yaml:"exit_code,omitempty"`

	Expose []string `json:"expose,omitempty" yaml:"expose,omitempty"`

	ExternalId string `json:"externalId,omitempty" yaml:"external_id,omitempty"`

	ExtraHosts []string `json:"extraHosts,omitempty" yaml:"extra_hosts,omitempty"`

	FirstRunning string `json:"firstRunning,omitempty" yaml:"first_running,omitempty"`

	ForceUpgrade bool `json:"forceUpgrade,omitempty" yaml:"force_upgrade,omitempty"`

	GroupAdd []string `json:"groupAdd,omitempty" yaml:"group_add,omitempty"`

	HealthCheck *InstanceHealthCheck `json:"healthCheck,omitempty" yaml:"health_check,omitempty"`

	HealthCmd []string `json:"healthCmd,omitempty" yaml:"health_cmd,omitempty"`

	HealthInterval int64 `json:"healthInterval,omitempty" yaml:"health_interval,omitempty"`

	HealthRetries int64 `json:"healthRetries,omitempty" yaml:"health_retries,omitempty"`

	HealthState string `json:"healthState,omitempty" yaml:"health_state,omitempty"`

	HealthTimeout int64 `json:"healthTimeout,omitempty" yaml:"health_timeout,omitempty"`

	HealthcheckStates []HealthcheckState `json:"healthcheckStates,omitempty" yaml:"healthcheck_states,omitempty"`

	HostId string `json:"hostId,omitempty" yaml:"host_id,omitempty"`

	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty"`

	Image string `json:"image,omitempty" yaml:"image,omitempty"`

	ImageUuid string `json:"imageUuid,omitempty" yaml:"image_uuid,omitempty"`

	InstanceTriggeredStop string `json:"instanceTriggeredStop,omitempty" yaml:"instance_triggered_stop,omitempty"`

	IoMaximumBandwidth int64 `json:"ioMaximumBandwidth,omitempty" yaml:"io_maximum_bandwidth,omitempty"`

	IoMaximumIOps int64 `json:"ioMaximumIOps,omitempty" yaml:"io_maximum_iops,omitempty"`

	Ip string `json:"ip,omitempty" yaml:"ip,omitempty"`

	Ip6 string `json:"ip6,omitempty" yaml:"ip6,omitempty"`

	IpcContainerId string `json:"ipcContainerId,omitempty" yaml:"ipc_container_id,omitempty"`

	IpcMode string `json:"ipcMode,omitempty" yaml:"ipc_mode,omitempty"`

	Isolation string `json:"isolation,omitempty" yaml:"isolation,omitempty"`

	KernelMemory int64 `json:"kernelMemory,omitempty" yaml:"kernel_memory,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`

	LogConfig *LogConfig `json:"logConfig,omitempty" yaml:"log_config,omitempty"`

	LxcConf map[string]string `json:"lxcConf,omitempty" yaml:"lxc_conf,omitempty"`

	Memory int64 `json:"memory,omitempty" yaml:"memory,omitempty"`

	MemoryMb int64 `json:"memoryMb,omitempty" yaml:"memory_mb,omitempty"`

	MemoryReservation int64 `json:"memoryReservation,omitempty" yaml:"memory_reservation,omitempty"`

	MemorySwap int64 `json:"memorySwap,omitempty" yaml:"memory_swap,omitempty"`

	MemorySwappiness int64 `json:"memorySwappiness,omitempty" yaml:"memory_swappiness,omitempty"`

	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	MilliCpuReservation int64 `json:"milliCpuReservation,omitempty" yaml:"milli_cpu_reservation,omitempty"`

	Mounts []MountEntry `json:"mounts,omitempty" yaml:"mounts,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	NativeContainer bool `json:"nativeContainer,omitempty" yaml:"native_container,omitempty"`

	NetAlias []string `json:"netAlias,omitempty" yaml:"net_alias,omitempty"`

	NetworkContainerId string `json:"networkContainerId,omitempty" yaml:"network_container_id,omitempty"`

	NetworkIds []string `json:"networkIds,omitempty" yaml:"network_ids,omitempty"`

	NetworkMode string `json:"networkMode,omitempty" yaml:"network_mode,omitempty"`

	OomKillDisable bool `json:"oomKillDisable,omitempty" yaml:"oom_kill_disable,omitempty"`

	OomScoreAdj int64 `json:"oomScoreAdj,omitempty" yaml:"oom_score_adj,omitempty"`

	PidContainerId string `json:"pidContainerId,omitempty" yaml:"pid_container_id,omitempty"`

	PidMode string `json:"pidMode,omitempty" yaml:"pid_mode,omitempty"`

	PidsLimit int64 `json:"pidsLimit,omitempty" yaml:"pids_limit,omitempty"`

	Ports []string `json:"ports,omitempty" yaml:"ports,omitempty"`

	PrePullOnUpgrade string `json:"prePullOnUpgrade,omitempty" yaml:"pre_pull_on_upgrade,omitempty"`

	PrimaryIpAddress string `json:"primaryIpAddress,omitempty" yaml:"primary_ip_address,omitempty"`

	PrimaryNetworkId string `json:"primaryNetworkId,omitempty" yaml:"primary_network_id,omitempty"`

	Privileged bool `json:"privileged,omitempty" yaml:"privileged,omitempty"`

	PublicEndpoints []PublicEndpoint `json:"publicEndpoints,omitempty" yaml:"public_endpoints,omitempty"`

	PublishAllPorts bool `json:"publishAllPorts,omitempty" yaml:"publish_all_ports,omitempty"`

	ReadOnly bool `json:"readOnly,omitempty" yaml:"read_only,omitempty"`

	RegistryCredentialId string `json:"registryCredentialId,omitempty" yaml:"registry_credential_id,omitempty"`

	RemoveTime string `json:"removeTime,omitempty" yaml:"remove_time,omitempty"`

	Removed string `json:"removed,omitempty" yaml:"removed,omitempty"`

	RequestedHostId string `json:"requestedHostId,omitempty" yaml:"requested_host_id,omitempty"`

	RequestedIpAddress string `json:"requestedIpAddress,omitempty" yaml:"requested_ip_address,omitempty"`

	RestartPolicy *RestartPolicy `json:"restartPolicy,omitempty" yaml:"restart_policy,omitempty"`

	RetainIp bool `json:"retainIp,omitempty" yaml:"retain_ip,omitempty"`

	RevisionId string `json:"revisionId,omitempty" yaml:"revision_id,omitempty"`

	Secrets []SecretReference `json:"secrets,omitempty" yaml:"secrets,omitempty"`

	SecurityOpt []string `json:"securityOpt,omitempty" yaml:"security_opt,omitempty"`

	ServiceId string `json:"serviceId,omitempty" yaml:"service_id,omitempty"`

	ServiceIds []string `json:"serviceIds,omitempty" yaml:"service_ids,omitempty"`

	ShmSize int64 `json:"shmSize,omitempty" yaml:"shm_size,omitempty"`

	ShouldRestart bool `json:"shouldRestart,omitempty" yaml:"should_restart,omitempty"`

	SidekickTo string `json:"sidekickTo,omitempty" yaml:"sidekick_to,omitempty"`

	StackId string `json:"stackId,omitempty" yaml:"stack_id,omitempty"`

	StartCount int64 `json:"startCount,omitempty" yaml:"start_count,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	StdinOpen bool `json:"stdinOpen,omitempty" yaml:"stdin_open,omitempty"`

	StopSignal string `json:"stopSignal,omitempty" yaml:"stop_signal,omitempty"`

	StopTimeout int64 `json:"stopTimeout,omitempty" yaml:"stop_timeout,omitempty"`

	StorageOpt map[string]string `json:"storageOpt,omitempty" yaml:"storage_opt,omitempty"`

	Sysctls map[string]string `json:"sysctls,omitempty" yaml:"sysctls,omitempty"`

	Tmpfs map[string]string `json:"tmpfs,omitempty" yaml:"tmpfs,omitempty"`

	Token string `json:"token,omitempty" yaml:"token,omitempty"`

	Transitioning string `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`

	TransitioningMessage string `json:"transitioningMessage,omitempty" yaml:"transitioning_message,omitempty"`

	Tty bool `json:"tty,omitempty" yaml:"tty,omitempty"`

	Ulimits []Ulimit `json:"ulimits,omitempty" yaml:"ulimits,omitempty"`

	User string `json:"user,omitempty" yaml:"user,omitempty"`

	UserPorts []string `json:"userPorts,omitempty" yaml:"user_ports,omitempty"`

	UsernsMode string `json:"usernsMode,omitempty" yaml:"userns_mode,omitempty"`

	Uts string `json:"uts,omitempty" yaml:"uts,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`

	Version string `json:"version,omitempty" yaml:"version,omitempty"`

	VolumeDriver string `json:"volumeDriver,omitempty" yaml:"volume_driver,omitempty"`

	WorkingDir string `json:"workingDir,omitempty" yaml:"working_dir,omitempty"`
}

type LaunchConfigCollection struct {
	Collection
	Data   []LaunchConfig `json:"data,omitempty"`
	client *LaunchConfigClient
}

type LaunchConfigClient struct {
	rancherClient *RancherClient
}

type LaunchConfigOperations interface {
	List(opts *ListOpts) (*LaunchConfigCollection, error)
	Create(opts *LaunchConfig) (*LaunchConfig, error)
	Update(existing *LaunchConfig, updates interface{}) (*LaunchConfig, error)
	ById(id string) (*LaunchConfig, error)
	Delete(container *LaunchConfig) error

	ActionConsole(*LaunchConfig, *InstanceConsoleInput) (*InstanceConsole, error)

	ActionConverttoservice(*LaunchConfig) (*Service, error)

	ActionCreate(*LaunchConfig) (*Instance, error)

	ActionError(*LaunchConfig) (*Instance, error)

	ActionExecute(*LaunchConfig, *ContainerExec) (*HostAccess, error)

	ActionProxy(*LaunchConfig, *ContainerProxy) (*HostAccess, error)

	ActionRemove(*LaunchConfig, *InstanceRemove) (*Instance, error)

	ActionRestart(*LaunchConfig) (*Instance, error)

	ActionStart(*LaunchConfig) (*Instance, error)

	ActionStop(*LaunchConfig, *InstanceStop) (*Instance, error)

	ActionUpgrade(*LaunchConfig, *ContainerUpgrade) (*Revision, error)
}

func newLaunchConfigClient(rancherClient *RancherClient) *LaunchConfigClient {
	return &LaunchConfigClient{
		rancherClient: rancherClient,
	}
}

func (c *LaunchConfigClient) Create(container *LaunchConfig) (*LaunchConfig, error) {
	resp := &LaunchConfig{}
	err := c.rancherClient.doCreate(LAUNCH_CONFIG_TYPE, container, resp)
	return resp, err
}

func (c *LaunchConfigClient) Update(existing *LaunchConfig, updates interface{}) (*LaunchConfig, error) {
	resp := &LaunchConfig{}
	err := c.rancherClient.doUpdate(LAUNCH_CONFIG_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *LaunchConfigClient) List(opts *ListOpts) (*LaunchConfigCollection, error) {
	resp := &LaunchConfigCollection{}
	err := c.rancherClient.doList(LAUNCH_CONFIG_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *LaunchConfigCollection) Next() (*LaunchConfigCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &LaunchConfigCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *LaunchConfigClient) ById(id string) (*LaunchConfig, error) {
	resp := &LaunchConfig{}
	err := c.rancherClient.doById(LAUNCH_CONFIG_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *LaunchConfigClient) Delete(container *LaunchConfig) error {
	return c.rancherClient.doResourceDelete(LAUNCH_CONFIG_TYPE, &container.Resource)
}

func (c *LaunchConfigClient) ActionConsole(resource *LaunchConfig, input *InstanceConsoleInput) (*InstanceConsole, error) {

	resp := &InstanceConsole{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "console", &resource.Resource, input, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionConverttoservice(resource *LaunchConfig) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "converttoservice", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionCreate(resource *LaunchConfig) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionError(resource *LaunchConfig) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "error", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionExecute(resource *LaunchConfig, input *ContainerExec) (*HostAccess, error) {

	resp := &HostAccess{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "execute", &resource.Resource, input, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionProxy(resource *LaunchConfig, input *ContainerProxy) (*HostAccess, error) {

	resp := &HostAccess{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "proxy", &resource.Resource, input, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionRemove(resource *LaunchConfig, input *InstanceRemove) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "remove", &resource.Resource, input, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionRestart(resource *LaunchConfig) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "restart", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionStart(resource *LaunchConfig) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "start", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionStop(resource *LaunchConfig, input *InstanceStop) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "stop", &resource.Resource, input, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionUpgrade(resource *LaunchConfig, input *ContainerUpgrade) (*Revision, error) {

	resp := &Revision{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "upgrade", &resource.Resource, input, resp)

	return resp, err
}

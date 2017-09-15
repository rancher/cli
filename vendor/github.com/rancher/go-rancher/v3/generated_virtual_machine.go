package client

const (
	VIRTUAL_MACHINE_TYPE = "virtualMachine"
)

type VirtualMachine struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	AgentId string `json:"agentId,omitempty" yaml:"agent_id,omitempty"`

	BlkioDeviceOptions map[string]interface{} `json:"blkioDeviceOptions,omitempty" yaml:"blkio_device_options,omitempty"`

	BlkioWeight int64 `json:"blkioWeight,omitempty" yaml:"blkio_weight,omitempty"`

	CgroupParent string `json:"cgroupParent,omitempty" yaml:"cgroup_parent,omitempty"`

	ClusterId string `json:"clusterId,omitempty" yaml:"cluster_id,omitempty"`

	Command []string `json:"command,omitempty" yaml:"command,omitempty"`

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

	DependsOn []DependsOn `json:"dependsOn,omitempty" yaml:"depends_on,omitempty"`

	DeploymentUnitId string `json:"deploymentUnitId,omitempty" yaml:"deployment_unit_id,omitempty"`

	DeploymentUnitUuid string `json:"deploymentUnitUuid,omitempty" yaml:"deployment_unit_uuid,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	Desired bool `json:"desired,omitempty" yaml:"desired,omitempty"`

	DiskQuota int64 `json:"diskQuota,omitempty" yaml:"disk_quota,omitempty"`

	Disks []VirtualMachineDisk `json:"disks,omitempty" yaml:"disks,omitempty"`

	Dns []string `json:"dns,omitempty" yaml:"dns,omitempty"`

	DnsOpt []string `json:"dnsOpt,omitempty" yaml:"dns_opt,omitempty"`

	DnsSearch []string `json:"dnsSearch,omitempty" yaml:"dns_search,omitempty"`

	DomainName string `json:"domainName,omitempty" yaml:"domain_name,omitempty"`

	ExitCode int64 `json:"exitCode,omitempty" yaml:"exit_code,omitempty"`

	Expose []string `json:"expose,omitempty" yaml:"expose,omitempty"`

	ExternalId string `json:"externalId,omitempty" yaml:"external_id,omitempty"`

	ExtraHosts []string `json:"extraHosts,omitempty" yaml:"extra_hosts,omitempty"`

	FirstRunning string `json:"firstRunning,omitempty" yaml:"first_running,omitempty"`

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

	InstanceLinks []Link `json:"instanceLinks,omitempty" yaml:"instance_links,omitempty"`

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

	NetworkIds []string `json:"networkIds,omitempty" yaml:"network_ids,omitempty"`

	NetworkMode string `json:"networkMode,omitempty" yaml:"network_mode,omitempty"`

	OomKillDisable bool `json:"oomKillDisable,omitempty" yaml:"oom_kill_disable,omitempty"`

	OomScoreAdj int64 `json:"oomScoreAdj,omitempty" yaml:"oom_score_adj,omitempty"`

	PidContainerId string `json:"pidContainerId,omitempty" yaml:"pid_container_id,omitempty"`

	PidsLimit int64 `json:"pidsLimit,omitempty" yaml:"pids_limit,omitempty"`

	Ports []string `json:"ports,omitempty" yaml:"ports,omitempty"`

	PrimaryIpAddress string `json:"primaryIpAddress,omitempty" yaml:"primary_ip_address,omitempty"`

	PrimaryNetworkId string `json:"primaryNetworkId,omitempty" yaml:"primary_network_id,omitempty"`

	PublicEndpoints []PublicEndpoint `json:"publicEndpoints,omitempty" yaml:"public_endpoints,omitempty"`

	RegistryCredentialId string `json:"registryCredentialId,omitempty" yaml:"registry_credential_id,omitempty"`

	RemoveTime string `json:"removeTime,omitempty" yaml:"remove_time,omitempty"`

	Removed string `json:"removed,omitempty" yaml:"removed,omitempty"`

	RequestedHostId string `json:"requestedHostId,omitempty" yaml:"requested_host_id,omitempty"`

	RequestedIpAddress string `json:"requestedIpAddress,omitempty" yaml:"requested_ip_address,omitempty"`

	RestartPolicy *RestartPolicy `json:"restartPolicy,omitempty" yaml:"restart_policy,omitempty"`

	RetainIp bool `json:"retainIp,omitempty" yaml:"retain_ip,omitempty"`

	SecurityOpt []string `json:"securityOpt,omitempty" yaml:"security_opt,omitempty"`

	ServiceId string `json:"serviceId,omitempty" yaml:"service_id,omitempty"`

	ServiceIds []string `json:"serviceIds,omitempty" yaml:"service_ids,omitempty"`

	ShmSize int64 `json:"shmSize,omitempty" yaml:"shm_size,omitempty"`

	ShouldRestart bool `json:"shouldRestart,omitempty" yaml:"should_restart,omitempty"`

	SidekickTo string `json:"sidekickTo,omitempty" yaml:"sidekick_to,omitempty"`

	StackId string `json:"stackId,omitempty" yaml:"stack_id,omitempty"`

	StartCount int64 `json:"startCount,omitempty" yaml:"start_count,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	StopSignal string `json:"stopSignal,omitempty" yaml:"stop_signal,omitempty"`

	StopTimeout int64 `json:"stopTimeout,omitempty" yaml:"stop_timeout,omitempty"`

	StorageOpt map[string]string `json:"storageOpt,omitempty" yaml:"storage_opt,omitempty"`

	Sysctls map[string]string `json:"sysctls,omitempty" yaml:"sysctls,omitempty"`

	Tmpfs map[string]string `json:"tmpfs,omitempty" yaml:"tmpfs,omitempty"`

	Token string `json:"token,omitempty" yaml:"token,omitempty"`

	Transitioning string `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`

	TransitioningMessage string `json:"transitioningMessage,omitempty" yaml:"transitioning_message,omitempty"`

	Ulimits []Ulimit `json:"ulimits,omitempty" yaml:"ulimits,omitempty"`

	UserPorts []string `json:"userPorts,omitempty" yaml:"user_ports,omitempty"`

	Userdata string `json:"userdata,omitempty" yaml:"userdata,omitempty"`

	UsernsMode string `json:"usernsMode,omitempty" yaml:"userns_mode,omitempty"`

	Uts string `json:"uts,omitempty" yaml:"uts,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`

	Vcpu int64 `json:"vcpu,omitempty" yaml:"vcpu,omitempty"`

	Version string `json:"version,omitempty" yaml:"version,omitempty"`

	VolumeDriver string `json:"volumeDriver,omitempty" yaml:"volume_driver,omitempty"`
}

type VirtualMachineCollection struct {
	Collection
	Data   []VirtualMachine `json:"data,omitempty"`
	client *VirtualMachineClient
}

type VirtualMachineClient struct {
	rancherClient *RancherClient
}

type VirtualMachineOperations interface {
	List(opts *ListOpts) (*VirtualMachineCollection, error)
	Create(opts *VirtualMachine) (*VirtualMachine, error)
	Update(existing *VirtualMachine, updates interface{}) (*VirtualMachine, error)
	ById(id string) (*VirtualMachine, error)
	Delete(container *VirtualMachine) error

	ActionConsole(*VirtualMachine, *InstanceConsoleInput) (*InstanceConsole, error)

	ActionConverttoservice(*VirtualMachine) (*Service, error)

	ActionCreate(*VirtualMachine) (*Instance, error)

	ActionError(*VirtualMachine) (*Instance, error)

	ActionExecute(*VirtualMachine, *ContainerExec) (*HostAccess, error)

	ActionLogs(*VirtualMachine, *ContainerLogs) (*HostAccess, error)

	ActionProxy(*VirtualMachine, *ContainerProxy) (*HostAccess, error)

	ActionRemove(*VirtualMachine, *InstanceRemove) (*Instance, error)

	ActionRestart(*VirtualMachine) (*Instance, error)

	ActionStart(*VirtualMachine) (*Instance, error)

	ActionStop(*VirtualMachine, *InstanceStop) (*Instance, error)

	ActionUpgrade(*VirtualMachine, *ContainerUpgrade) (*Revision, error)
}

func newVirtualMachineClient(rancherClient *RancherClient) *VirtualMachineClient {
	return &VirtualMachineClient{
		rancherClient: rancherClient,
	}
}

func (c *VirtualMachineClient) Create(container *VirtualMachine) (*VirtualMachine, error) {
	resp := &VirtualMachine{}
	err := c.rancherClient.doCreate(VIRTUAL_MACHINE_TYPE, container, resp)
	return resp, err
}

func (c *VirtualMachineClient) Update(existing *VirtualMachine, updates interface{}) (*VirtualMachine, error) {
	resp := &VirtualMachine{}
	err := c.rancherClient.doUpdate(VIRTUAL_MACHINE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *VirtualMachineClient) List(opts *ListOpts) (*VirtualMachineCollection, error) {
	resp := &VirtualMachineCollection{}
	err := c.rancherClient.doList(VIRTUAL_MACHINE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *VirtualMachineCollection) Next() (*VirtualMachineCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &VirtualMachineCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *VirtualMachineClient) ById(id string) (*VirtualMachine, error) {
	resp := &VirtualMachine{}
	err := c.rancherClient.doById(VIRTUAL_MACHINE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *VirtualMachineClient) Delete(container *VirtualMachine) error {
	return c.rancherClient.doResourceDelete(VIRTUAL_MACHINE_TYPE, &container.Resource)
}

func (c *VirtualMachineClient) ActionConsole(resource *VirtualMachine, input *InstanceConsoleInput) (*InstanceConsole, error) {

	resp := &InstanceConsole{}

	err := c.rancherClient.doAction(VIRTUAL_MACHINE_TYPE, "console", &resource.Resource, input, resp)

	return resp, err
}

func (c *VirtualMachineClient) ActionConverttoservice(resource *VirtualMachine) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(VIRTUAL_MACHINE_TYPE, "converttoservice", &resource.Resource, nil, resp)

	return resp, err
}

func (c *VirtualMachineClient) ActionCreate(resource *VirtualMachine) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(VIRTUAL_MACHINE_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *VirtualMachineClient) ActionError(resource *VirtualMachine) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(VIRTUAL_MACHINE_TYPE, "error", &resource.Resource, nil, resp)

	return resp, err
}

func (c *VirtualMachineClient) ActionExecute(resource *VirtualMachine, input *ContainerExec) (*HostAccess, error) {

	resp := &HostAccess{}

	err := c.rancherClient.doAction(VIRTUAL_MACHINE_TYPE, "execute", &resource.Resource, input, resp)

	return resp, err
}

func (c *VirtualMachineClient) ActionLogs(resource *VirtualMachine, input *ContainerLogs) (*HostAccess, error) {

	resp := &HostAccess{}

	err := c.rancherClient.doAction(VIRTUAL_MACHINE_TYPE, "logs", &resource.Resource, input, resp)

	return resp, err
}

func (c *VirtualMachineClient) ActionProxy(resource *VirtualMachine, input *ContainerProxy) (*HostAccess, error) {

	resp := &HostAccess{}

	err := c.rancherClient.doAction(VIRTUAL_MACHINE_TYPE, "proxy", &resource.Resource, input, resp)

	return resp, err
}

func (c *VirtualMachineClient) ActionRemove(resource *VirtualMachine, input *InstanceRemove) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(VIRTUAL_MACHINE_TYPE, "remove", &resource.Resource, input, resp)

	return resp, err
}

func (c *VirtualMachineClient) ActionRestart(resource *VirtualMachine) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(VIRTUAL_MACHINE_TYPE, "restart", &resource.Resource, nil, resp)

	return resp, err
}

func (c *VirtualMachineClient) ActionStart(resource *VirtualMachine) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(VIRTUAL_MACHINE_TYPE, "start", &resource.Resource, nil, resp)

	return resp, err
}

func (c *VirtualMachineClient) ActionStop(resource *VirtualMachine, input *InstanceStop) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(VIRTUAL_MACHINE_TYPE, "stop", &resource.Resource, input, resp)

	return resp, err
}

func (c *VirtualMachineClient) ActionUpgrade(resource *VirtualMachine, input *ContainerUpgrade) (*Revision, error) {

	resp := &Revision{}

	err := c.rancherClient.doAction(VIRTUAL_MACHINE_TYPE, "upgrade", &resource.Resource, input, resp)

	return resp, err
}

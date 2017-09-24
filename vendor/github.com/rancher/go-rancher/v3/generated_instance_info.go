package client

const (
	INSTANCE_INFO_TYPE = "instanceInfo"
)

type InstanceInfo struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	AgentId string `json:"agentId,omitempty" yaml:"agent_id,omitempty"`

	CreateIndex int64 `json:"createIndex,omitempty" yaml:"create_index,omitempty"`

	DeploymentUnitId string `json:"deploymentUnitId,omitempty" yaml:"deployment_unit_id,omitempty"`

	Desired bool `json:"desired,omitempty" yaml:"desired,omitempty"`

	Dns []string `json:"dns,omitempty" yaml:"dns,omitempty"`

	DnsSearch []string `json:"dnsSearch,omitempty" yaml:"dns_search,omitempty"`

	EnvironmentUuid string `json:"environmentUuid,omitempty" yaml:"environment_uuid,omitempty"`

	ExitCode int64 `json:"exitCode,omitempty" yaml:"exit_code,omitempty"`

	ExternalId string `json:"externalId,omitempty" yaml:"external_id,omitempty"`

	HealthCheck HealthcheckInfo `json:"healthCheck,omitempty" yaml:"health_check,omitempty"`

	HealthCheckHosts []HealthcheckState `json:"healthCheckHosts,omitempty" yaml:"health_check_hosts,omitempty"`

	HealthState string `json:"healthState,omitempty" yaml:"health_state,omitempty"`

	HostId string `json:"hostId,omitempty" yaml:"host_id,omitempty"`

	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty"`

	InfoType string `json:"infoType,omitempty" yaml:"info_type,omitempty"`

	InfoTypeId string `json:"infoTypeId,omitempty" yaml:"info_type_id,omitempty"`

	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`

	Links []Link `json:"links,omitempty" yaml:"links,omitempty"`

	MemoryReservation int64 `json:"memoryReservation,omitempty" yaml:"memory_reservation,omitempty"`

	MilliCpuReservation int64 `json:"milliCpuReservation,omitempty" yaml:"milli_cpu_reservation,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	NativeContainer bool `json:"nativeContainer,omitempty" yaml:"native_container,omitempty"`

	NetworkFromContainerId string `json:"networkFromContainerId,omitempty" yaml:"network_from_container_id,omitempty"`

	NetworkId string `json:"networkId,omitempty" yaml:"network_id,omitempty"`

	Ports []PublicEndpoint `json:"ports,omitempty" yaml:"ports,omitempty"`

	PrimaryIp string `json:"primaryIp,omitempty" yaml:"primary_ip,omitempty"`

	PrimaryMacAddress string `json:"primaryMacAddress,omitempty" yaml:"primary_mac_address,omitempty"`

	ServiceId string `json:"serviceId,omitempty" yaml:"service_id,omitempty"`

	ServiceIds []string `json:"serviceIds,omitempty" yaml:"service_ids,omitempty"`

	ServiceIndex int64 `json:"serviceIndex,omitempty" yaml:"service_index,omitempty"`

	ShouldRestart bool `json:"shouldRestart,omitempty" yaml:"should_restart,omitempty"`

	StackId string `json:"stackId,omitempty" yaml:"stack_id,omitempty"`

	StartCount int64 `json:"startCount,omitempty" yaml:"start_count,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type InstanceInfoCollection struct {
	Collection
	Data   []InstanceInfo `json:"data,omitempty"`
	client *InstanceInfoClient
}

type InstanceInfoClient struct {
	rancherClient *RancherClient
}

type InstanceInfoOperations interface {
	List(opts *ListOpts) (*InstanceInfoCollection, error)
	Create(opts *InstanceInfo) (*InstanceInfo, error)
	Update(existing *InstanceInfo, updates interface{}) (*InstanceInfo, error)
	ById(id string) (*InstanceInfo, error)
	Delete(container *InstanceInfo) error
}

func newInstanceInfoClient(rancherClient *RancherClient) *InstanceInfoClient {
	return &InstanceInfoClient{
		rancherClient: rancherClient,
	}
}

func (c *InstanceInfoClient) Create(container *InstanceInfo) (*InstanceInfo, error) {
	resp := &InstanceInfo{}
	err := c.rancherClient.doCreate(INSTANCE_INFO_TYPE, container, resp)
	return resp, err
}

func (c *InstanceInfoClient) Update(existing *InstanceInfo, updates interface{}) (*InstanceInfo, error) {
	resp := &InstanceInfo{}
	err := c.rancherClient.doUpdate(INSTANCE_INFO_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *InstanceInfoClient) List(opts *ListOpts) (*InstanceInfoCollection, error) {
	resp := &InstanceInfoCollection{}
	err := c.rancherClient.doList(INSTANCE_INFO_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *InstanceInfoCollection) Next() (*InstanceInfoCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &InstanceInfoCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *InstanceInfoClient) ById(id string) (*InstanceInfo, error) {
	resp := &InstanceInfo{}
	err := c.rancherClient.doById(INSTANCE_INFO_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *InstanceInfoClient) Delete(container *InstanceInfo) error {
	return c.rancherClient.doResourceDelete(INSTANCE_INFO_TYPE, &container.Resource)
}

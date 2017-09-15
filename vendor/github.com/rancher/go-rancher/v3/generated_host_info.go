package client

const (
	HOST_INFO_TYPE = "hostInfo"
)

type HostInfo struct {
	Resource

	AgentId string `json:"agentId,omitempty" yaml:"agent_id,omitempty"`

	AgentIp string `json:"agentIp,omitempty" yaml:"agent_ip,omitempty"`

	AgentState string `json:"agentState,omitempty" yaml:"agent_state,omitempty"`

	ClusterId string `json:"clusterId,omitempty" yaml:"cluster_id,omitempty"`

	EnvironmentUuid string `json:"environmentUuid,omitempty" yaml:"environment_uuid,omitempty"`

	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty"`

	InfoType string `json:"infoType,omitempty" yaml:"info_type,omitempty"`

	InfoTypeId string `json:"infoTypeId,omitempty" yaml:"info_type_id,omitempty"`

	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`

	Memory int64 `json:"memory,omitempty" yaml:"memory,omitempty"`

	MilliCpu int64 `json:"milliCpu,omitempty" yaml:"milli_cpu,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	NodeName string `json:"nodeName,omitempty" yaml:"node_name,omitempty"`

	Ports []PublicEndpoint `json:"ports,omitempty" yaml:"ports,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type HostInfoCollection struct {
	Collection
	Data   []HostInfo `json:"data,omitempty"`
	client *HostInfoClient
}

type HostInfoClient struct {
	rancherClient *RancherClient
}

type HostInfoOperations interface {
	List(opts *ListOpts) (*HostInfoCollection, error)
	Create(opts *HostInfo) (*HostInfo, error)
	Update(existing *HostInfo, updates interface{}) (*HostInfo, error)
	ById(id string) (*HostInfo, error)
	Delete(container *HostInfo) error
}

func newHostInfoClient(rancherClient *RancherClient) *HostInfoClient {
	return &HostInfoClient{
		rancherClient: rancherClient,
	}
}

func (c *HostInfoClient) Create(container *HostInfo) (*HostInfo, error) {
	resp := &HostInfo{}
	err := c.rancherClient.doCreate(HOST_INFO_TYPE, container, resp)
	return resp, err
}

func (c *HostInfoClient) Update(existing *HostInfo, updates interface{}) (*HostInfo, error) {
	resp := &HostInfo{}
	err := c.rancherClient.doUpdate(HOST_INFO_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *HostInfoClient) List(opts *ListOpts) (*HostInfoCollection, error) {
	resp := &HostInfoCollection{}
	err := c.rancherClient.doList(HOST_INFO_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *HostInfoCollection) Next() (*HostInfoCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &HostInfoCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *HostInfoClient) ById(id string) (*HostInfo, error) {
	resp := &HostInfo{}
	err := c.rancherClient.doById(HOST_INFO_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *HostInfoClient) Delete(container *HostInfo) error {
	return c.rancherClient.doResourceDelete(HOST_INFO_TYPE, &container.Resource)
}

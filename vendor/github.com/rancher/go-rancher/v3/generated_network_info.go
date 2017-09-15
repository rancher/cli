package client

const (
	NETWORK_INFO_TYPE = "networkInfo"
)

type NetworkInfo struct {
	Resource

	DefaultPolicyAction string `json:"defaultPolicyAction,omitempty" yaml:"default_policy_action,omitempty"`

	EnvironmentUuid string `json:"environmentUuid,omitempty" yaml:"environment_uuid,omitempty"`

	HostPorts bool `json:"hostPorts,omitempty" yaml:"host_ports,omitempty"`

	InfoType string `json:"infoType,omitempty" yaml:"info_type,omitempty"`

	InfoTypeId string `json:"infoTypeId,omitempty" yaml:"info_type_id,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	Policy interface{} `json:"policy,omitempty" yaml:"policy,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type NetworkInfoCollection struct {
	Collection
	Data   []NetworkInfo `json:"data,omitempty"`
	client *NetworkInfoClient
}

type NetworkInfoClient struct {
	rancherClient *RancherClient
}

type NetworkInfoOperations interface {
	List(opts *ListOpts) (*NetworkInfoCollection, error)
	Create(opts *NetworkInfo) (*NetworkInfo, error)
	Update(existing *NetworkInfo, updates interface{}) (*NetworkInfo, error)
	ById(id string) (*NetworkInfo, error)
	Delete(container *NetworkInfo) error
}

func newNetworkInfoClient(rancherClient *RancherClient) *NetworkInfoClient {
	return &NetworkInfoClient{
		rancherClient: rancherClient,
	}
}

func (c *NetworkInfoClient) Create(container *NetworkInfo) (*NetworkInfo, error) {
	resp := &NetworkInfo{}
	err := c.rancherClient.doCreate(NETWORK_INFO_TYPE, container, resp)
	return resp, err
}

func (c *NetworkInfoClient) Update(existing *NetworkInfo, updates interface{}) (*NetworkInfo, error) {
	resp := &NetworkInfo{}
	err := c.rancherClient.doUpdate(NETWORK_INFO_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *NetworkInfoClient) List(opts *ListOpts) (*NetworkInfoCollection, error) {
	resp := &NetworkInfoCollection{}
	err := c.rancherClient.doList(NETWORK_INFO_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *NetworkInfoCollection) Next() (*NetworkInfoCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &NetworkInfoCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *NetworkInfoClient) ById(id string) (*NetworkInfo, error) {
	resp := &NetworkInfo{}
	err := c.rancherClient.doById(NETWORK_INFO_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *NetworkInfoClient) Delete(container *NetworkInfo) error {
	return c.rancherClient.doResourceDelete(NETWORK_INFO_TYPE, &container.Resource)
}

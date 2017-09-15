package client

const (
	SERVICE_INFO_TYPE = "serviceInfo"
)

type ServiceInfo struct {
	Resource

	EnvironmentUuid string `json:"environmentUuid,omitempty" yaml:"environment_uuid,omitempty"`

	ExternalIps []string `json:"externalIps,omitempty" yaml:"external_ips,omitempty"`

	Fqdn string `json:"fqdn,omitempty" yaml:"fqdn,omitempty"`

	Global bool `json:"global,omitempty" yaml:"global,omitempty"`

	HealthCheck HealthcheckInfo `json:"healthCheck,omitempty" yaml:"health_check,omitempty"`

	HealthState string `json:"healthState,omitempty" yaml:"health_state,omitempty"`

	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty"`

	InfoType string `json:"infoType,omitempty" yaml:"info_type,omitempty"`

	InfoTypeId string `json:"infoTypeId,omitempty" yaml:"info_type_id,omitempty"`

	InstanceIds []string `json:"instanceIds,omitempty" yaml:"instance_ids,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`

	LbConfig LbConfig `json:"lbConfig,omitempty" yaml:"lb_config,omitempty"`

	Links []Link `json:"links,omitempty" yaml:"links,omitempty"`

	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	Ports []PublicEndpoint `json:"ports,omitempty" yaml:"ports,omitempty"`

	Scale int64 `json:"scale,omitempty" yaml:"scale,omitempty"`

	Selector string `json:"selector,omitempty" yaml:"selector,omitempty"`

	Sidekicks []string `json:"sidekicks,omitempty" yaml:"sidekicks,omitempty"`

	StackId string `json:"stackId,omitempty" yaml:"stack_id,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	Token string `json:"token,omitempty" yaml:"token,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`

	Vip string `json:"vip,omitempty" yaml:"vip,omitempty"`
}

type ServiceInfoCollection struct {
	Collection
	Data   []ServiceInfo `json:"data,omitempty"`
	client *ServiceInfoClient
}

type ServiceInfoClient struct {
	rancherClient *RancherClient
}

type ServiceInfoOperations interface {
	List(opts *ListOpts) (*ServiceInfoCollection, error)
	Create(opts *ServiceInfo) (*ServiceInfo, error)
	Update(existing *ServiceInfo, updates interface{}) (*ServiceInfo, error)
	ById(id string) (*ServiceInfo, error)
	Delete(container *ServiceInfo) error
}

func newServiceInfoClient(rancherClient *RancherClient) *ServiceInfoClient {
	return &ServiceInfoClient{
		rancherClient: rancherClient,
	}
}

func (c *ServiceInfoClient) Create(container *ServiceInfo) (*ServiceInfo, error) {
	resp := &ServiceInfo{}
	err := c.rancherClient.doCreate(SERVICE_INFO_TYPE, container, resp)
	return resp, err
}

func (c *ServiceInfoClient) Update(existing *ServiceInfo, updates interface{}) (*ServiceInfo, error) {
	resp := &ServiceInfo{}
	err := c.rancherClient.doUpdate(SERVICE_INFO_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ServiceInfoClient) List(opts *ListOpts) (*ServiceInfoCollection, error) {
	resp := &ServiceInfoCollection{}
	err := c.rancherClient.doList(SERVICE_INFO_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ServiceInfoCollection) Next() (*ServiceInfoCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ServiceInfoCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ServiceInfoClient) ById(id string) (*ServiceInfo, error) {
	resp := &ServiceInfo{}
	err := c.rancherClient.doById(SERVICE_INFO_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *ServiceInfoClient) Delete(container *ServiceInfo) error {
	return c.rancherClient.doResourceDelete(SERVICE_INFO_TYPE, &container.Resource)
}

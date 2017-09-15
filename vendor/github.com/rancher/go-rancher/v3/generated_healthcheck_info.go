package client

const (
	HEALTHCHECK_INFO_TYPE = "healthcheckInfo"
)

type HealthcheckInfo struct {
	Resource

	HealthyThreshold int64 `json:"healthyThreshold,omitempty" yaml:"healthy_threshold,omitempty"`

	InitializingTimeout int64 `json:"initializingTimeout,omitempty" yaml:"initializing_timeout,omitempty"`

	Interval int64 `json:"interval,omitempty" yaml:"interval,omitempty"`

	Port int64 `json:"port,omitempty" yaml:"port,omitempty"`

	RequestLine string `json:"requestLine,omitempty" yaml:"request_line,omitempty"`

	ResponseTimeout int64 `json:"responseTimeout,omitempty" yaml:"response_timeout,omitempty"`

	UnhealthyThreshold int64 `json:"unhealthyThreshold,omitempty" yaml:"unhealthy_threshold,omitempty"`
}

type HealthcheckInfoCollection struct {
	Collection
	Data   []HealthcheckInfo `json:"data,omitempty"`
	client *HealthcheckInfoClient
}

type HealthcheckInfoClient struct {
	rancherClient *RancherClient
}

type HealthcheckInfoOperations interface {
	List(opts *ListOpts) (*HealthcheckInfoCollection, error)
	Create(opts *HealthcheckInfo) (*HealthcheckInfo, error)
	Update(existing *HealthcheckInfo, updates interface{}) (*HealthcheckInfo, error)
	ById(id string) (*HealthcheckInfo, error)
	Delete(container *HealthcheckInfo) error
}

func newHealthcheckInfoClient(rancherClient *RancherClient) *HealthcheckInfoClient {
	return &HealthcheckInfoClient{
		rancherClient: rancherClient,
	}
}

func (c *HealthcheckInfoClient) Create(container *HealthcheckInfo) (*HealthcheckInfo, error) {
	resp := &HealthcheckInfo{}
	err := c.rancherClient.doCreate(HEALTHCHECK_INFO_TYPE, container, resp)
	return resp, err
}

func (c *HealthcheckInfoClient) Update(existing *HealthcheckInfo, updates interface{}) (*HealthcheckInfo, error) {
	resp := &HealthcheckInfo{}
	err := c.rancherClient.doUpdate(HEALTHCHECK_INFO_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *HealthcheckInfoClient) List(opts *ListOpts) (*HealthcheckInfoCollection, error) {
	resp := &HealthcheckInfoCollection{}
	err := c.rancherClient.doList(HEALTHCHECK_INFO_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *HealthcheckInfoCollection) Next() (*HealthcheckInfoCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &HealthcheckInfoCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *HealthcheckInfoClient) ById(id string) (*HealthcheckInfo, error) {
	resp := &HealthcheckInfo{}
	err := c.rancherClient.doById(HEALTHCHECK_INFO_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *HealthcheckInfoClient) Delete(container *HealthcheckInfo) error {
	return c.rancherClient.doResourceDelete(HEALTHCHECK_INFO_TYPE, &container.Resource)
}

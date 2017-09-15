package client

const (
	HEALTHCHECK_STATE_TYPE = "healthcheckState"
)

type HealthcheckState struct {
	Resource

	HealthState string `json:"healthState,omitempty" yaml:"health_state,omitempty"`

	HostId string `json:"hostId,omitempty" yaml:"host_id,omitempty"`
}

type HealthcheckStateCollection struct {
	Collection
	Data   []HealthcheckState `json:"data,omitempty"`
	client *HealthcheckStateClient
}

type HealthcheckStateClient struct {
	rancherClient *RancherClient
}

type HealthcheckStateOperations interface {
	List(opts *ListOpts) (*HealthcheckStateCollection, error)
	Create(opts *HealthcheckState) (*HealthcheckState, error)
	Update(existing *HealthcheckState, updates interface{}) (*HealthcheckState, error)
	ById(id string) (*HealthcheckState, error)
	Delete(container *HealthcheckState) error
}

func newHealthcheckStateClient(rancherClient *RancherClient) *HealthcheckStateClient {
	return &HealthcheckStateClient{
		rancherClient: rancherClient,
	}
}

func (c *HealthcheckStateClient) Create(container *HealthcheckState) (*HealthcheckState, error) {
	resp := &HealthcheckState{}
	err := c.rancherClient.doCreate(HEALTHCHECK_STATE_TYPE, container, resp)
	return resp, err
}

func (c *HealthcheckStateClient) Update(existing *HealthcheckState, updates interface{}) (*HealthcheckState, error) {
	resp := &HealthcheckState{}
	err := c.rancherClient.doUpdate(HEALTHCHECK_STATE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *HealthcheckStateClient) List(opts *ListOpts) (*HealthcheckStateCollection, error) {
	resp := &HealthcheckStateCollection{}
	err := c.rancherClient.doList(HEALTHCHECK_STATE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *HealthcheckStateCollection) Next() (*HealthcheckStateCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &HealthcheckStateCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *HealthcheckStateClient) ById(id string) (*HealthcheckState, error) {
	resp := &HealthcheckState{}
	err := c.rancherClient.doById(HEALTHCHECK_STATE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *HealthcheckStateClient) Delete(container *HealthcheckState) error {
	return c.rancherClient.doResourceDelete(HEALTHCHECK_STATE_TYPE, &container.Resource)
}

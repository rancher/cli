package client

const (
	VOLUME_TEMPLATE_TYPE = "volumeTemplate"
)

type VolumeTemplate struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	ClusterId string `json:"clusterId,omitempty" yaml:"cluster_id,omitempty"`

	Created string `json:"created,omitempty" yaml:"created,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	Driver string `json:"driver,omitempty" yaml:"driver,omitempty"`

	DriverOpts map[string]string `json:"driverOpts,omitempty" yaml:"driver_opts,omitempty"`

	External bool `json:"external,omitempty" yaml:"external,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	PerContainer bool `json:"perContainer,omitempty" yaml:"per_container,omitempty"`

	RemoveTime string `json:"removeTime,omitempty" yaml:"remove_time,omitempty"`

	Removed string `json:"removed,omitempty" yaml:"removed,omitempty"`

	StackId string `json:"stackId,omitempty" yaml:"stack_id,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	Transitioning string `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`

	TransitioningMessage string `json:"transitioningMessage,omitempty" yaml:"transitioning_message,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type VolumeTemplateCollection struct {
	Collection
	Data   []VolumeTemplate `json:"data,omitempty"`
	client *VolumeTemplateClient
}

type VolumeTemplateClient struct {
	rancherClient *RancherClient
}

type VolumeTemplateOperations interface {
	List(opts *ListOpts) (*VolumeTemplateCollection, error)
	Create(opts *VolumeTemplate) (*VolumeTemplate, error)
	Update(existing *VolumeTemplate, updates interface{}) (*VolumeTemplate, error)
	ById(id string) (*VolumeTemplate, error)
	Delete(container *VolumeTemplate) error

	ActionCreate(*VolumeTemplate) (*VolumeTemplate, error)

	ActionRemove(*VolumeTemplate) (*VolumeTemplate, error)
}

func newVolumeTemplateClient(rancherClient *RancherClient) *VolumeTemplateClient {
	return &VolumeTemplateClient{
		rancherClient: rancherClient,
	}
}

func (c *VolumeTemplateClient) Create(container *VolumeTemplate) (*VolumeTemplate, error) {
	resp := &VolumeTemplate{}
	err := c.rancherClient.doCreate(VOLUME_TEMPLATE_TYPE, container, resp)
	return resp, err
}

func (c *VolumeTemplateClient) Update(existing *VolumeTemplate, updates interface{}) (*VolumeTemplate, error) {
	resp := &VolumeTemplate{}
	err := c.rancherClient.doUpdate(VOLUME_TEMPLATE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *VolumeTemplateClient) List(opts *ListOpts) (*VolumeTemplateCollection, error) {
	resp := &VolumeTemplateCollection{}
	err := c.rancherClient.doList(VOLUME_TEMPLATE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *VolumeTemplateCollection) Next() (*VolumeTemplateCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &VolumeTemplateCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *VolumeTemplateClient) ById(id string) (*VolumeTemplate, error) {
	resp := &VolumeTemplate{}
	err := c.rancherClient.doById(VOLUME_TEMPLATE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *VolumeTemplateClient) Delete(container *VolumeTemplate) error {
	return c.rancherClient.doResourceDelete(VOLUME_TEMPLATE_TYPE, &container.Resource)
}

func (c *VolumeTemplateClient) ActionCreate(resource *VolumeTemplate) (*VolumeTemplate, error) {

	resp := &VolumeTemplate{}

	err := c.rancherClient.doAction(VOLUME_TEMPLATE_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *VolumeTemplateClient) ActionRemove(resource *VolumeTemplate) (*VolumeTemplate, error) {

	resp := &VolumeTemplate{}

	err := c.rancherClient.doAction(VOLUME_TEMPLATE_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}

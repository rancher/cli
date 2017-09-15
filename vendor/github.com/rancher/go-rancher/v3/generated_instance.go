package client

const (
	INSTANCE_TYPE = "instance"
)

type Instance struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	ClusterId string `json:"clusterId,omitempty" yaml:"cluster_id,omitempty"`

	Created string `json:"created,omitempty" yaml:"created,omitempty"`

	DependsOn []DependsOn `json:"dependsOn,omitempty" yaml:"depends_on,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	Desired bool `json:"desired,omitempty" yaml:"desired,omitempty"`

	ExternalId string `json:"externalId,omitempty" yaml:"external_id,omitempty"`

	HealthcheckStates []HealthcheckState `json:"healthcheckStates,omitempty" yaml:"healthcheck_states,omitempty"`

	HostId string `json:"hostId,omitempty" yaml:"host_id,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	RemoveTime string `json:"removeTime,omitempty" yaml:"remove_time,omitempty"`

	Removed string `json:"removed,omitempty" yaml:"removed,omitempty"`

	RevisionId string `json:"revisionId,omitempty" yaml:"revision_id,omitempty"`

	ServiceId string `json:"serviceId,omitempty" yaml:"service_id,omitempty"`

	ServiceIds []string `json:"serviceIds,omitempty" yaml:"service_ids,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	Transitioning string `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`

	TransitioningMessage string `json:"transitioningMessage,omitempty" yaml:"transitioning_message,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type InstanceCollection struct {
	Collection
	Data   []Instance `json:"data,omitempty"`
	client *InstanceClient
}

type InstanceClient struct {
	rancherClient *RancherClient
}

type InstanceOperations interface {
	List(opts *ListOpts) (*InstanceCollection, error)
	Create(opts *Instance) (*Instance, error)
	Update(existing *Instance, updates interface{}) (*Instance, error)
	ById(id string) (*Instance, error)
	Delete(container *Instance) error

	ActionConsole(*Instance, *InstanceConsoleInput) (*InstanceConsole, error)

	ActionCreate(*Instance) (*Instance, error)

	ActionError(*Instance) (*Instance, error)

	ActionRemove(*Instance, *InstanceRemove) (*Instance, error)

	ActionRestart(*Instance) (*Instance, error)

	ActionStart(*Instance) (*Instance, error)

	ActionStop(*Instance, *InstanceStop) (*Instance, error)
}

func newInstanceClient(rancherClient *RancherClient) *InstanceClient {
	return &InstanceClient{
		rancherClient: rancherClient,
	}
}

func (c *InstanceClient) Create(container *Instance) (*Instance, error) {
	resp := &Instance{}
	err := c.rancherClient.doCreate(INSTANCE_TYPE, container, resp)
	return resp, err
}

func (c *InstanceClient) Update(existing *Instance, updates interface{}) (*Instance, error) {
	resp := &Instance{}
	err := c.rancherClient.doUpdate(INSTANCE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *InstanceClient) List(opts *ListOpts) (*InstanceCollection, error) {
	resp := &InstanceCollection{}
	err := c.rancherClient.doList(INSTANCE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *InstanceCollection) Next() (*InstanceCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &InstanceCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *InstanceClient) ById(id string) (*Instance, error) {
	resp := &Instance{}
	err := c.rancherClient.doById(INSTANCE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *InstanceClient) Delete(container *Instance) error {
	return c.rancherClient.doResourceDelete(INSTANCE_TYPE, &container.Resource)
}

func (c *InstanceClient) ActionConsole(resource *Instance, input *InstanceConsoleInput) (*InstanceConsole, error) {

	resp := &InstanceConsole{}

	err := c.rancherClient.doAction(INSTANCE_TYPE, "console", &resource.Resource, input, resp)

	return resp, err
}

func (c *InstanceClient) ActionCreate(resource *Instance) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(INSTANCE_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *InstanceClient) ActionError(resource *Instance) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(INSTANCE_TYPE, "error", &resource.Resource, nil, resp)

	return resp, err
}

func (c *InstanceClient) ActionRemove(resource *Instance, input *InstanceRemove) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(INSTANCE_TYPE, "remove", &resource.Resource, input, resp)

	return resp, err
}

func (c *InstanceClient) ActionRestart(resource *Instance) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(INSTANCE_TYPE, "restart", &resource.Resource, nil, resp)

	return resp, err
}

func (c *InstanceClient) ActionStart(resource *Instance) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(INSTANCE_TYPE, "start", &resource.Resource, nil, resp)

	return resp, err
}

func (c *InstanceClient) ActionStop(resource *Instance, input *InstanceStop) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(INSTANCE_TYPE, "stop", &resource.Resource, input, resp)

	return resp, err
}

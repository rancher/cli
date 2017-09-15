package client

const (
	DEPLOYMENT_UNIT_TYPE = "deploymentUnit"
)

type DeploymentUnit struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	ClusterId string `json:"clusterId,omitempty" yaml:"cluster_id,omitempty"`

	Created string `json:"created,omitempty" yaml:"created,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	HostId string `json:"hostId,omitempty" yaml:"host_id,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	RemoveTime string `json:"removeTime,omitempty" yaml:"remove_time,omitempty"`

	Removed string `json:"removed,omitempty" yaml:"removed,omitempty"`

	RequestedRevisionId string `json:"requestedRevisionId,omitempty" yaml:"requested_revision_id,omitempty"`

	RevisionId string `json:"revisionId,omitempty" yaml:"revision_id,omitempty"`

	ServiceId string `json:"serviceId,omitempty" yaml:"service_id,omitempty"`

	ServiceIndex string `json:"serviceIndex,omitempty" yaml:"service_index,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	Transitioning string `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`

	TransitioningMessage string `json:"transitioningMessage,omitempty" yaml:"transitioning_message,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type DeploymentUnitCollection struct {
	Collection
	Data   []DeploymentUnit `json:"data,omitempty"`
	client *DeploymentUnitClient
}

type DeploymentUnitClient struct {
	rancherClient *RancherClient
}

type DeploymentUnitOperations interface {
	List(opts *ListOpts) (*DeploymentUnitCollection, error)
	Create(opts *DeploymentUnit) (*DeploymentUnit, error)
	Update(existing *DeploymentUnit, updates interface{}) (*DeploymentUnit, error)
	ById(id string) (*DeploymentUnit, error)
	Delete(container *DeploymentUnit) error

	ActionActivate(*DeploymentUnit) (*DeploymentUnit, error)

	ActionCreate(*DeploymentUnit) (*DeploymentUnit, error)

	ActionDeactivate(*DeploymentUnit) (*DeploymentUnit, error)

	ActionError(*DeploymentUnit) (*DeploymentUnit, error)

	ActionPause(*DeploymentUnit) (*DeploymentUnit, error)

	ActionRemove(*DeploymentUnit) (*DeploymentUnit, error)

	ActionUpdate(*DeploymentUnit) (*DeploymentUnit, error)
}

func newDeploymentUnitClient(rancherClient *RancherClient) *DeploymentUnitClient {
	return &DeploymentUnitClient{
		rancherClient: rancherClient,
	}
}

func (c *DeploymentUnitClient) Create(container *DeploymentUnit) (*DeploymentUnit, error) {
	resp := &DeploymentUnit{}
	err := c.rancherClient.doCreate(DEPLOYMENT_UNIT_TYPE, container, resp)
	return resp, err
}

func (c *DeploymentUnitClient) Update(existing *DeploymentUnit, updates interface{}) (*DeploymentUnit, error) {
	resp := &DeploymentUnit{}
	err := c.rancherClient.doUpdate(DEPLOYMENT_UNIT_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *DeploymentUnitClient) List(opts *ListOpts) (*DeploymentUnitCollection, error) {
	resp := &DeploymentUnitCollection{}
	err := c.rancherClient.doList(DEPLOYMENT_UNIT_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *DeploymentUnitCollection) Next() (*DeploymentUnitCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &DeploymentUnitCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *DeploymentUnitClient) ById(id string) (*DeploymentUnit, error) {
	resp := &DeploymentUnit{}
	err := c.rancherClient.doById(DEPLOYMENT_UNIT_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *DeploymentUnitClient) Delete(container *DeploymentUnit) error {
	return c.rancherClient.doResourceDelete(DEPLOYMENT_UNIT_TYPE, &container.Resource)
}

func (c *DeploymentUnitClient) ActionActivate(resource *DeploymentUnit) (*DeploymentUnit, error) {

	resp := &DeploymentUnit{}

	err := c.rancherClient.doAction(DEPLOYMENT_UNIT_TYPE, "activate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *DeploymentUnitClient) ActionCreate(resource *DeploymentUnit) (*DeploymentUnit, error) {

	resp := &DeploymentUnit{}

	err := c.rancherClient.doAction(DEPLOYMENT_UNIT_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *DeploymentUnitClient) ActionDeactivate(resource *DeploymentUnit) (*DeploymentUnit, error) {

	resp := &DeploymentUnit{}

	err := c.rancherClient.doAction(DEPLOYMENT_UNIT_TYPE, "deactivate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *DeploymentUnitClient) ActionError(resource *DeploymentUnit) (*DeploymentUnit, error) {

	resp := &DeploymentUnit{}

	err := c.rancherClient.doAction(DEPLOYMENT_UNIT_TYPE, "error", &resource.Resource, nil, resp)

	return resp, err
}

func (c *DeploymentUnitClient) ActionPause(resource *DeploymentUnit) (*DeploymentUnit, error) {

	resp := &DeploymentUnit{}

	err := c.rancherClient.doAction(DEPLOYMENT_UNIT_TYPE, "pause", &resource.Resource, nil, resp)

	return resp, err
}

func (c *DeploymentUnitClient) ActionRemove(resource *DeploymentUnit) (*DeploymentUnit, error) {

	resp := &DeploymentUnit{}

	err := c.rancherClient.doAction(DEPLOYMENT_UNIT_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}

func (c *DeploymentUnitClient) ActionUpdate(resource *DeploymentUnit) (*DeploymentUnit, error) {

	resp := &DeploymentUnit{}

	err := c.rancherClient.doAction(DEPLOYMENT_UNIT_TYPE, "update", &resource.Resource, nil, resp)

	return resp, err
}

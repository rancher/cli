package client

const (
	STACK_TYPE = "stack"
)

type Stack struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	Answers map[string]interface{} `json:"answers,omitempty" yaml:"answers,omitempty"`

	ClusterId string `json:"clusterId,omitempty" yaml:"cluster_id,omitempty"`

	Created string `json:"created,omitempty" yaml:"created,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	ExternalId string `json:"externalId,omitempty" yaml:"external_id,omitempty"`

	Group string `json:"group,omitempty" yaml:"group,omitempty"`

	HealthState string `json:"healthState,omitempty" yaml:"health_state,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	Outputs map[string]string `json:"outputs,omitempty" yaml:"outputs,omitempty"`

	ParentStackId string `json:"parentStackId,omitempty" yaml:"parent_stack_id,omitempty"`

	Prune bool `json:"prune,omitempty" yaml:"prune,omitempty"`

	RemoveTime string `json:"removeTime,omitempty" yaml:"remove_time,omitempty"`

	Removed string `json:"removed,omitempty" yaml:"removed,omitempty"`

	ServiceIds []string `json:"serviceIds,omitempty" yaml:"service_ids,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	Templates map[string]string `json:"templates,omitempty" yaml:"templates,omitempty"`

	Transitioning string `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`

	TransitioningMessage string `json:"transitioningMessage,omitempty" yaml:"transitioning_message,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`

	WorkingConfiguration StackConfiguration `json:"workingConfiguration,omitempty" yaml:"working_configuration,omitempty"`
}

type StackCollection struct {
	Collection
	Data   []Stack `json:"data,omitempty"`
	client *StackClient
}

type StackClient struct {
	rancherClient *RancherClient
}

type StackOperations interface {
	List(opts *ListOpts) (*StackCollection, error)
	Create(opts *Stack) (*Stack, error)
	Update(existing *Stack, updates interface{}) (*Stack, error)
	ById(id string) (*Stack, error)
	Delete(container *Stack) error

	ActionActivateservices(*Stack) (*Stack, error)

	ActionAddoutputs(*Stack, *AddOutputsInput) (*Stack, error)

	ActionCreate(*Stack) (*Stack, error)

	ActionDeactivateservices(*Stack) (*Stack, error)

	ActionError(*Stack) (*Stack, error)

	ActionExportconfig(*Stack, *ComposeConfigInput) (*ComposeConfig, error)

	ActionPause(*Stack) (*Stack, error)

	ActionRemove(*Stack) (*Stack, error)

	ActionRollback(*Stack) (*Stack, error)

	ActionUpdate(*Stack) (*Stack, error)
}

func newStackClient(rancherClient *RancherClient) *StackClient {
	return &StackClient{
		rancherClient: rancherClient,
	}
}

func (c *StackClient) Create(container *Stack) (*Stack, error) {
	resp := &Stack{}
	err := c.rancherClient.doCreate(STACK_TYPE, container, resp)
	return resp, err
}

func (c *StackClient) Update(existing *Stack, updates interface{}) (*Stack, error) {
	resp := &Stack{}
	err := c.rancherClient.doUpdate(STACK_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *StackClient) List(opts *ListOpts) (*StackCollection, error) {
	resp := &StackCollection{}
	err := c.rancherClient.doList(STACK_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *StackCollection) Next() (*StackCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &StackCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *StackClient) ById(id string) (*Stack, error) {
	resp := &Stack{}
	err := c.rancherClient.doById(STACK_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *StackClient) Delete(container *Stack) error {
	return c.rancherClient.doResourceDelete(STACK_TYPE, &container.Resource)
}

func (c *StackClient) ActionActivateservices(resource *Stack) (*Stack, error) {

	resp := &Stack{}

	err := c.rancherClient.doAction(STACK_TYPE, "activateservices", &resource.Resource, nil, resp)

	return resp, err
}

func (c *StackClient) ActionAddoutputs(resource *Stack, input *AddOutputsInput) (*Stack, error) {

	resp := &Stack{}

	err := c.rancherClient.doAction(STACK_TYPE, "addoutputs", &resource.Resource, input, resp)

	return resp, err
}

func (c *StackClient) ActionCreate(resource *Stack) (*Stack, error) {

	resp := &Stack{}

	err := c.rancherClient.doAction(STACK_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *StackClient) ActionDeactivateservices(resource *Stack) (*Stack, error) {

	resp := &Stack{}

	err := c.rancherClient.doAction(STACK_TYPE, "deactivateservices", &resource.Resource, nil, resp)

	return resp, err
}

func (c *StackClient) ActionError(resource *Stack) (*Stack, error) {

	resp := &Stack{}

	err := c.rancherClient.doAction(STACK_TYPE, "error", &resource.Resource, nil, resp)

	return resp, err
}

func (c *StackClient) ActionExportconfig(resource *Stack, input *ComposeConfigInput) (*ComposeConfig, error) {

	resp := &ComposeConfig{}

	err := c.rancherClient.doAction(STACK_TYPE, "exportconfig", &resource.Resource, input, resp)

	return resp, err
}

func (c *StackClient) ActionPause(resource *Stack) (*Stack, error) {

	resp := &Stack{}

	err := c.rancherClient.doAction(STACK_TYPE, "pause", &resource.Resource, nil, resp)

	return resp, err
}

func (c *StackClient) ActionRemove(resource *Stack) (*Stack, error) {

	resp := &Stack{}

	err := c.rancherClient.doAction(STACK_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}

func (c *StackClient) ActionRollback(resource *Stack) (*Stack, error) {

	resp := &Stack{}

	err := c.rancherClient.doAction(STACK_TYPE, "rollback", &resource.Resource, nil, resp)

	return resp, err
}

func (c *StackClient) ActionUpdate(resource *Stack) (*Stack, error) {

	resp := &Stack{}

	err := c.rancherClient.doAction(STACK_TYPE, "update", &resource.Resource, nil, resp)

	return resp, err
}

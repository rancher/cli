package client

const (
	EXTERNAL_SERVICE_TYPE = "externalService"
)

type ExternalService struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	ClusterId string `json:"clusterId,omitempty" yaml:"cluster_id,omitempty"`

	CompleteUpdate bool `json:"completeUpdate,omitempty" yaml:"complete_update,omitempty"`

	CreateOnly bool `json:"createOnly,omitempty" yaml:"create_only,omitempty"`

	Created string `json:"created,omitempty" yaml:"created,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	ExternalId string `json:"externalId,omitempty" yaml:"external_id,omitempty"`

	ExternalIpAddresses []string `json:"externalIpAddresses,omitempty" yaml:"external_ip_addresses,omitempty"`

	Fqdn string `json:"fqdn,omitempty" yaml:"fqdn,omitempty"`

	HealthCheck *InstanceHealthCheck `json:"healthCheck,omitempty" yaml:"health_check,omitempty"`

	HealthState string `json:"healthState,omitempty" yaml:"health_state,omitempty"`

	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty"`

	InstanceIds []string `json:"instanceIds,omitempty" yaml:"instance_ids,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	LaunchConfig *LaunchConfig `json:"launchConfig,omitempty" yaml:"launch_config,omitempty"`

	LbTargetConfig *LbTargetConfig `json:"lbTargetConfig,omitempty" yaml:"lb_target_config,omitempty"`

	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	PreviousRevisionId string `json:"previousRevisionId,omitempty" yaml:"previous_revision_id,omitempty"`

	RemoveTime string `json:"removeTime,omitempty" yaml:"remove_time,omitempty"`

	Removed string `json:"removed,omitempty" yaml:"removed,omitempty"`

	RevisionId string `json:"revisionId,omitempty" yaml:"revision_id,omitempty"`

	ServiceLinks []Link `json:"serviceLinks,omitempty" yaml:"service_links,omitempty"`

	StackId string `json:"stackId,omitempty" yaml:"stack_id,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	Transitioning string `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`

	TransitioningMessage string `json:"transitioningMessage,omitempty" yaml:"transitioning_message,omitempty"`

	Upgrade *ServiceUpgrade `json:"upgrade,omitempty" yaml:"upgrade,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type ExternalServiceCollection struct {
	Collection
	Data   []ExternalService `json:"data,omitempty"`
	client *ExternalServiceClient
}

type ExternalServiceClient struct {
	rancherClient *RancherClient
}

type ExternalServiceOperations interface {
	List(opts *ListOpts) (*ExternalServiceCollection, error)
	Create(opts *ExternalService) (*ExternalService, error)
	Update(existing *ExternalService, updates interface{}) (*ExternalService, error)
	ById(id string) (*ExternalService, error)
	Delete(container *ExternalService) error

	ActionActivate(*ExternalService) (*Service, error)

	ActionCancelupgrade(*ExternalService) (*Service, error)

	ActionCreate(*ExternalService) (*Service, error)

	ActionDeactivate(*ExternalService) (*Service, error)

	ActionError(*ExternalService) (*Service, error)

	ActionFinishupgrade(*ExternalService) (*Service, error)

	ActionGarbagecollect(*ExternalService) (*Service, error)

	ActionPause(*ExternalService) (*Service, error)

	ActionRemove(*ExternalService) (*Service, error)

	ActionRestart(*ExternalService) (*Service, error)

	ActionRollback(*ExternalService, *ServiceRollback) (*Service, error)

	ActionUpdate(*ExternalService) (*Service, error)

	ActionUpgrade(*ExternalService, *ServiceUpgrade) (*Service, error)
}

func newExternalServiceClient(rancherClient *RancherClient) *ExternalServiceClient {
	return &ExternalServiceClient{
		rancherClient: rancherClient,
	}
}

func (c *ExternalServiceClient) Create(container *ExternalService) (*ExternalService, error) {
	resp := &ExternalService{}
	err := c.rancherClient.doCreate(EXTERNAL_SERVICE_TYPE, container, resp)
	return resp, err
}

func (c *ExternalServiceClient) Update(existing *ExternalService, updates interface{}) (*ExternalService, error) {
	resp := &ExternalService{}
	err := c.rancherClient.doUpdate(EXTERNAL_SERVICE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ExternalServiceClient) List(opts *ListOpts) (*ExternalServiceCollection, error) {
	resp := &ExternalServiceCollection{}
	err := c.rancherClient.doList(EXTERNAL_SERVICE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ExternalServiceCollection) Next() (*ExternalServiceCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ExternalServiceCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ExternalServiceClient) ById(id string) (*ExternalService, error) {
	resp := &ExternalService{}
	err := c.rancherClient.doById(EXTERNAL_SERVICE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *ExternalServiceClient) Delete(container *ExternalService) error {
	return c.rancherClient.doResourceDelete(EXTERNAL_SERVICE_TYPE, &container.Resource)
}

func (c *ExternalServiceClient) ActionActivate(resource *ExternalService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(EXTERNAL_SERVICE_TYPE, "activate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ExternalServiceClient) ActionCancelupgrade(resource *ExternalService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(EXTERNAL_SERVICE_TYPE, "cancelupgrade", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ExternalServiceClient) ActionCreate(resource *ExternalService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(EXTERNAL_SERVICE_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ExternalServiceClient) ActionDeactivate(resource *ExternalService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(EXTERNAL_SERVICE_TYPE, "deactivate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ExternalServiceClient) ActionError(resource *ExternalService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(EXTERNAL_SERVICE_TYPE, "error", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ExternalServiceClient) ActionFinishupgrade(resource *ExternalService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(EXTERNAL_SERVICE_TYPE, "finishupgrade", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ExternalServiceClient) ActionGarbagecollect(resource *ExternalService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(EXTERNAL_SERVICE_TYPE, "garbagecollect", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ExternalServiceClient) ActionPause(resource *ExternalService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(EXTERNAL_SERVICE_TYPE, "pause", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ExternalServiceClient) ActionRemove(resource *ExternalService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(EXTERNAL_SERVICE_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ExternalServiceClient) ActionRestart(resource *ExternalService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(EXTERNAL_SERVICE_TYPE, "restart", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ExternalServiceClient) ActionRollback(resource *ExternalService, input *ServiceRollback) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(EXTERNAL_SERVICE_TYPE, "rollback", &resource.Resource, input, resp)

	return resp, err
}

func (c *ExternalServiceClient) ActionUpdate(resource *ExternalService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(EXTERNAL_SERVICE_TYPE, "update", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ExternalServiceClient) ActionUpgrade(resource *ExternalService, input *ServiceUpgrade) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(EXTERNAL_SERVICE_TYPE, "upgrade", &resource.Resource, input, resp)

	return resp, err
}

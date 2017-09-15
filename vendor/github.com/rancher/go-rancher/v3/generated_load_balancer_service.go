package client

const (
	LOAD_BALANCER_SERVICE_TYPE = "loadBalancerService"
)

type LoadBalancerService struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	AssignServiceIpAddress bool `json:"assignServiceIpAddress,omitempty" yaml:"assign_service_ip_address,omitempty"`

	BatchSize int64 `json:"batchSize,omitempty" yaml:"batch_size,omitempty"`

	ClusterId string `json:"clusterId,omitempty" yaml:"cluster_id,omitempty"`

	CompleteUpdate bool `json:"completeUpdate,omitempty" yaml:"complete_update,omitempty"`

	CreateOnly bool `json:"createOnly,omitempty" yaml:"create_only,omitempty"`

	Created string `json:"created,omitempty" yaml:"created,omitempty"`

	CurrentScale int64 `json:"currentScale,omitempty" yaml:"current_scale,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	ExternalId string `json:"externalId,omitempty" yaml:"external_id,omitempty"`

	Fqdn string `json:"fqdn,omitempty" yaml:"fqdn,omitempty"`

	HealthState string `json:"healthState,omitempty" yaml:"health_state,omitempty"`

	InstanceIds []string `json:"instanceIds,omitempty" yaml:"instance_ids,omitempty"`

	IntervalMillis int64 `json:"intervalMillis,omitempty" yaml:"interval_millis,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	LaunchConfig *LaunchConfig `json:"launchConfig,omitempty" yaml:"launch_config,omitempty"`

	LbConfig *LbConfig `json:"lbConfig,omitempty" yaml:"lb_config,omitempty"`

	LbTargetConfig *LbTargetConfig `json:"lbTargetConfig,omitempty" yaml:"lb_target_config,omitempty"`

	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	PreviousRevisionId string `json:"previousRevisionId,omitempty" yaml:"previous_revision_id,omitempty"`

	PublicEndpoints []PublicEndpoint `json:"publicEndpoints,omitempty" yaml:"public_endpoints,omitempty"`

	RemoveTime string `json:"removeTime,omitempty" yaml:"remove_time,omitempty"`

	Removed string `json:"removed,omitempty" yaml:"removed,omitempty"`

	RevisionId string `json:"revisionId,omitempty" yaml:"revision_id,omitempty"`

	Scale int64 `json:"scale,omitempty" yaml:"scale,omitempty"`

	ScaleIncrement int64 `json:"scaleIncrement,omitempty" yaml:"scale_increment,omitempty"`

	ScaleMax int64 `json:"scaleMax,omitempty" yaml:"scale_max,omitempty"`

	ScaleMin int64 `json:"scaleMin,omitempty" yaml:"scale_min,omitempty"`

	ServiceLinks []Link `json:"serviceLinks,omitempty" yaml:"service_links,omitempty"`

	StackId string `json:"stackId,omitempty" yaml:"stack_id,omitempty"`

	StartFirst bool `json:"startFirst,omitempty" yaml:"start_first,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	Transitioning string `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`

	TransitioningMessage string `json:"transitioningMessage,omitempty" yaml:"transitioning_message,omitempty"`

	Upgrade *ServiceUpgrade `json:"upgrade,omitempty" yaml:"upgrade,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`

	Vip string `json:"vip,omitempty" yaml:"vip,omitempty"`
}

type LoadBalancerServiceCollection struct {
	Collection
	Data   []LoadBalancerService `json:"data,omitempty"`
	client *LoadBalancerServiceClient
}

type LoadBalancerServiceClient struct {
	rancherClient *RancherClient
}

type LoadBalancerServiceOperations interface {
	List(opts *ListOpts) (*LoadBalancerServiceCollection, error)
	Create(opts *LoadBalancerService) (*LoadBalancerService, error)
	Update(existing *LoadBalancerService, updates interface{}) (*LoadBalancerService, error)
	ById(id string) (*LoadBalancerService, error)
	Delete(container *LoadBalancerService) error

	ActionActivate(*LoadBalancerService) (*Service, error)

	ActionCancelupgrade(*LoadBalancerService) (*Service, error)

	ActionCreate(*LoadBalancerService) (*Service, error)

	ActionDeactivate(*LoadBalancerService) (*Service, error)

	ActionError(*LoadBalancerService) (*Service, error)

	ActionFinishupgrade(*LoadBalancerService) (*Service, error)

	ActionGarbagecollect(*LoadBalancerService) (*Service, error)

	ActionPause(*LoadBalancerService) (*Service, error)

	ActionRemove(*LoadBalancerService) (*Service, error)

	ActionRestart(*LoadBalancerService) (*Service, error)

	ActionRollback(*LoadBalancerService, *ServiceRollback) (*Service, error)

	ActionUpdate(*LoadBalancerService) (*Service, error)

	ActionUpgrade(*LoadBalancerService, *ServiceUpgrade) (*Service, error)
}

func newLoadBalancerServiceClient(rancherClient *RancherClient) *LoadBalancerServiceClient {
	return &LoadBalancerServiceClient{
		rancherClient: rancherClient,
	}
}

func (c *LoadBalancerServiceClient) Create(container *LoadBalancerService) (*LoadBalancerService, error) {
	resp := &LoadBalancerService{}
	err := c.rancherClient.doCreate(LOAD_BALANCER_SERVICE_TYPE, container, resp)
	return resp, err
}

func (c *LoadBalancerServiceClient) Update(existing *LoadBalancerService, updates interface{}) (*LoadBalancerService, error) {
	resp := &LoadBalancerService{}
	err := c.rancherClient.doUpdate(LOAD_BALANCER_SERVICE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *LoadBalancerServiceClient) List(opts *ListOpts) (*LoadBalancerServiceCollection, error) {
	resp := &LoadBalancerServiceCollection{}
	err := c.rancherClient.doList(LOAD_BALANCER_SERVICE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *LoadBalancerServiceCollection) Next() (*LoadBalancerServiceCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &LoadBalancerServiceCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *LoadBalancerServiceClient) ById(id string) (*LoadBalancerService, error) {
	resp := &LoadBalancerService{}
	err := c.rancherClient.doById(LOAD_BALANCER_SERVICE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *LoadBalancerServiceClient) Delete(container *LoadBalancerService) error {
	return c.rancherClient.doResourceDelete(LOAD_BALANCER_SERVICE_TYPE, &container.Resource)
}

func (c *LoadBalancerServiceClient) ActionActivate(resource *LoadBalancerService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(LOAD_BALANCER_SERVICE_TYPE, "activate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LoadBalancerServiceClient) ActionCancelupgrade(resource *LoadBalancerService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(LOAD_BALANCER_SERVICE_TYPE, "cancelupgrade", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LoadBalancerServiceClient) ActionCreate(resource *LoadBalancerService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(LOAD_BALANCER_SERVICE_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LoadBalancerServiceClient) ActionDeactivate(resource *LoadBalancerService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(LOAD_BALANCER_SERVICE_TYPE, "deactivate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LoadBalancerServiceClient) ActionError(resource *LoadBalancerService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(LOAD_BALANCER_SERVICE_TYPE, "error", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LoadBalancerServiceClient) ActionFinishupgrade(resource *LoadBalancerService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(LOAD_BALANCER_SERVICE_TYPE, "finishupgrade", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LoadBalancerServiceClient) ActionGarbagecollect(resource *LoadBalancerService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(LOAD_BALANCER_SERVICE_TYPE, "garbagecollect", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LoadBalancerServiceClient) ActionPause(resource *LoadBalancerService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(LOAD_BALANCER_SERVICE_TYPE, "pause", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LoadBalancerServiceClient) ActionRemove(resource *LoadBalancerService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(LOAD_BALANCER_SERVICE_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LoadBalancerServiceClient) ActionRestart(resource *LoadBalancerService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(LOAD_BALANCER_SERVICE_TYPE, "restart", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LoadBalancerServiceClient) ActionRollback(resource *LoadBalancerService, input *ServiceRollback) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(LOAD_BALANCER_SERVICE_TYPE, "rollback", &resource.Resource, input, resp)

	return resp, err
}

func (c *LoadBalancerServiceClient) ActionUpdate(resource *LoadBalancerService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(LOAD_BALANCER_SERVICE_TYPE, "update", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LoadBalancerServiceClient) ActionUpgrade(resource *LoadBalancerService, input *ServiceUpgrade) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(LOAD_BALANCER_SERVICE_TYPE, "upgrade", &resource.Resource, input, resp)

	return resp, err
}

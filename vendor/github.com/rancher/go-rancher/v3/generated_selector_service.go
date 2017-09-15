package client

const (
	SELECTOR_SERVICE_TYPE = "selectorService"
)

type SelectorService struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	AssignServiceIpAddress bool `json:"assignServiceIpAddress,omitempty" yaml:"assign_service_ip_address,omitempty"`

	BatchSize int64 `json:"batchSize,omitempty" yaml:"batch_size,omitempty"`

	ClusterId string `json:"clusterId,omitempty" yaml:"cluster_id,omitempty"`

	CompleteUpdate bool `json:"completeUpdate,omitempty" yaml:"complete_update,omitempty"`

	CreateIndex int64 `json:"createIndex,omitempty" yaml:"create_index,omitempty"`

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

	SecondaryLaunchConfigs []LaunchConfig `json:"secondaryLaunchConfigs,omitempty" yaml:"secondary_launch_configs,omitempty"`

	Selector string `json:"selector,omitempty" yaml:"selector,omitempty"`

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

type SelectorServiceCollection struct {
	Collection
	Data   []SelectorService `json:"data,omitempty"`
	client *SelectorServiceClient
}

type SelectorServiceClient struct {
	rancherClient *RancherClient
}

type SelectorServiceOperations interface {
	List(opts *ListOpts) (*SelectorServiceCollection, error)
	Create(opts *SelectorService) (*SelectorService, error)
	Update(existing *SelectorService, updates interface{}) (*SelectorService, error)
	ById(id string) (*SelectorService, error)
	Delete(container *SelectorService) error

	ActionActivate(*SelectorService) (*Service, error)

	ActionCancelupgrade(*SelectorService) (*Service, error)

	ActionCreate(*SelectorService) (*Service, error)

	ActionDeactivate(*SelectorService) (*Service, error)

	ActionError(*SelectorService) (*Service, error)

	ActionFinishupgrade(*SelectorService) (*Service, error)

	ActionGarbagecollect(*SelectorService) (*Service, error)

	ActionPause(*SelectorService) (*Service, error)

	ActionRemove(*SelectorService) (*Service, error)

	ActionRestart(*SelectorService) (*Service, error)

	ActionRollback(*SelectorService, *ServiceRollback) (*Service, error)

	ActionUpdate(*SelectorService) (*Service, error)

	ActionUpgrade(*SelectorService, *ServiceUpgrade) (*Service, error)
}

func newSelectorServiceClient(rancherClient *RancherClient) *SelectorServiceClient {
	return &SelectorServiceClient{
		rancherClient: rancherClient,
	}
}

func (c *SelectorServiceClient) Create(container *SelectorService) (*SelectorService, error) {
	resp := &SelectorService{}
	err := c.rancherClient.doCreate(SELECTOR_SERVICE_TYPE, container, resp)
	return resp, err
}

func (c *SelectorServiceClient) Update(existing *SelectorService, updates interface{}) (*SelectorService, error) {
	resp := &SelectorService{}
	err := c.rancherClient.doUpdate(SELECTOR_SERVICE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *SelectorServiceClient) List(opts *ListOpts) (*SelectorServiceCollection, error) {
	resp := &SelectorServiceCollection{}
	err := c.rancherClient.doList(SELECTOR_SERVICE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *SelectorServiceCollection) Next() (*SelectorServiceCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &SelectorServiceCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *SelectorServiceClient) ById(id string) (*SelectorService, error) {
	resp := &SelectorService{}
	err := c.rancherClient.doById(SELECTOR_SERVICE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *SelectorServiceClient) Delete(container *SelectorService) error {
	return c.rancherClient.doResourceDelete(SELECTOR_SERVICE_TYPE, &container.Resource)
}

func (c *SelectorServiceClient) ActionActivate(resource *SelectorService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SELECTOR_SERVICE_TYPE, "activate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *SelectorServiceClient) ActionCancelupgrade(resource *SelectorService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SELECTOR_SERVICE_TYPE, "cancelupgrade", &resource.Resource, nil, resp)

	return resp, err
}

func (c *SelectorServiceClient) ActionCreate(resource *SelectorService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SELECTOR_SERVICE_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *SelectorServiceClient) ActionDeactivate(resource *SelectorService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SELECTOR_SERVICE_TYPE, "deactivate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *SelectorServiceClient) ActionError(resource *SelectorService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SELECTOR_SERVICE_TYPE, "error", &resource.Resource, nil, resp)

	return resp, err
}

func (c *SelectorServiceClient) ActionFinishupgrade(resource *SelectorService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SELECTOR_SERVICE_TYPE, "finishupgrade", &resource.Resource, nil, resp)

	return resp, err
}

func (c *SelectorServiceClient) ActionGarbagecollect(resource *SelectorService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SELECTOR_SERVICE_TYPE, "garbagecollect", &resource.Resource, nil, resp)

	return resp, err
}

func (c *SelectorServiceClient) ActionPause(resource *SelectorService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SELECTOR_SERVICE_TYPE, "pause", &resource.Resource, nil, resp)

	return resp, err
}

func (c *SelectorServiceClient) ActionRemove(resource *SelectorService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SELECTOR_SERVICE_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}

func (c *SelectorServiceClient) ActionRestart(resource *SelectorService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SELECTOR_SERVICE_TYPE, "restart", &resource.Resource, nil, resp)

	return resp, err
}

func (c *SelectorServiceClient) ActionRollback(resource *SelectorService, input *ServiceRollback) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SELECTOR_SERVICE_TYPE, "rollback", &resource.Resource, input, resp)

	return resp, err
}

func (c *SelectorServiceClient) ActionUpdate(resource *SelectorService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SELECTOR_SERVICE_TYPE, "update", &resource.Resource, nil, resp)

	return resp, err
}

func (c *SelectorServiceClient) ActionUpgrade(resource *SelectorService, input *ServiceUpgrade) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SELECTOR_SERVICE_TYPE, "upgrade", &resource.Resource, input, resp)

	return resp, err
}

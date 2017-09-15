package client

const (
	SERVICE_TYPE = "service"
)

type Service struct {
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

	ExternalIpAddresses []string `json:"externalIpAddresses,omitempty" yaml:"external_ip_addresses,omitempty"`

	Fqdn string `json:"fqdn,omitempty" yaml:"fqdn,omitempty"`

	HealthCheck *InstanceHealthCheck `json:"healthCheck,omitempty" yaml:"health_check,omitempty"`

	HealthState string `json:"healthState,omitempty" yaml:"health_state,omitempty"`

	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty"`

	InstanceIds []string `json:"instanceIds,omitempty" yaml:"instance_ids,omitempty"`

	IntervalMillis int64 `json:"intervalMillis,omitempty" yaml:"interval_millis,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	LaunchConfig *LaunchConfig `json:"launchConfig,omitempty" yaml:"launch_config,omitempty"`

	LbConfig *LbConfig `json:"lbConfig,omitempty" yaml:"lb_config,omitempty"`

	LbTargetConfig *LbTargetConfig `json:"lbTargetConfig,omitempty" yaml:"lb_target_config,omitempty"`

	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	NetworkDriver *NetworkDriver `json:"networkDriver,omitempty" yaml:"network_driver,omitempty"`

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

	StorageDriver *StorageDriver `json:"storageDriver,omitempty" yaml:"storage_driver,omitempty"`

	Transitioning string `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`

	TransitioningMessage string `json:"transitioningMessage,omitempty" yaml:"transitioning_message,omitempty"`

	Upgrade *ServiceUpgrade `json:"upgrade,omitempty" yaml:"upgrade,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`

	Vip string `json:"vip,omitempty" yaml:"vip,omitempty"`
}

type ServiceCollection struct {
	Collection
	Data   []Service `json:"data,omitempty"`
	client *ServiceClient
}

type ServiceClient struct {
	rancherClient *RancherClient
}

type ServiceOperations interface {
	List(opts *ListOpts) (*ServiceCollection, error)
	Create(opts *Service) (*Service, error)
	Update(existing *Service, updates interface{}) (*Service, error)
	ById(id string) (*Service, error)
	Delete(container *Service) error

	ActionActivate(*Service) (*Service, error)

	ActionCancelupgrade(*Service) (*Service, error)

	ActionCreate(*Service) (*Service, error)

	ActionDeactivate(*Service) (*Service, error)

	ActionError(*Service) (*Service, error)

	ActionFinishupgrade(*Service) (*Service, error)

	ActionGarbagecollect(*Service) (*Service, error)

	ActionPause(*Service) (*Service, error)

	ActionRemove(*Service) (*Service, error)

	ActionRestart(*Service) (*Service, error)

	ActionRollback(*Service, *ServiceRollback) (*Service, error)

	ActionUpdate(*Service) (*Service, error)

	ActionUpgrade(*Service, *ServiceUpgrade) (*Service, error)
}

func newServiceClient(rancherClient *RancherClient) *ServiceClient {
	return &ServiceClient{
		rancherClient: rancherClient,
	}
}

func (c *ServiceClient) Create(container *Service) (*Service, error) {
	resp := &Service{}
	err := c.rancherClient.doCreate(SERVICE_TYPE, container, resp)
	return resp, err
}

func (c *ServiceClient) Update(existing *Service, updates interface{}) (*Service, error) {
	resp := &Service{}
	err := c.rancherClient.doUpdate(SERVICE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ServiceClient) List(opts *ListOpts) (*ServiceCollection, error) {
	resp := &ServiceCollection{}
	err := c.rancherClient.doList(SERVICE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ServiceCollection) Next() (*ServiceCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ServiceCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ServiceClient) ById(id string) (*Service, error) {
	resp := &Service{}
	err := c.rancherClient.doById(SERVICE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *ServiceClient) Delete(container *Service) error {
	return c.rancherClient.doResourceDelete(SERVICE_TYPE, &container.Resource)
}

func (c *ServiceClient) ActionActivate(resource *Service) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SERVICE_TYPE, "activate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ServiceClient) ActionCancelupgrade(resource *Service) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SERVICE_TYPE, "cancelupgrade", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ServiceClient) ActionCreate(resource *Service) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SERVICE_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ServiceClient) ActionDeactivate(resource *Service) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SERVICE_TYPE, "deactivate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ServiceClient) ActionError(resource *Service) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SERVICE_TYPE, "error", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ServiceClient) ActionFinishupgrade(resource *Service) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SERVICE_TYPE, "finishupgrade", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ServiceClient) ActionGarbagecollect(resource *Service) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SERVICE_TYPE, "garbagecollect", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ServiceClient) ActionPause(resource *Service) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SERVICE_TYPE, "pause", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ServiceClient) ActionRemove(resource *Service) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SERVICE_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ServiceClient) ActionRestart(resource *Service) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SERVICE_TYPE, "restart", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ServiceClient) ActionRollback(resource *Service, input *ServiceRollback) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SERVICE_TYPE, "rollback", &resource.Resource, input, resp)

	return resp, err
}

func (c *ServiceClient) ActionUpdate(resource *Service) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SERVICE_TYPE, "update", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ServiceClient) ActionUpgrade(resource *Service, input *ServiceUpgrade) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SERVICE_TYPE, "upgrade", &resource.Resource, input, resp)

	return resp, err
}

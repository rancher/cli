package client

const (
	NETWORK_DRIVER_SERVICE_TYPE = "networkDriverService"
)

type NetworkDriverService struct {
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

	NetworkDriver NetworkDriver `json:"networkDriver,omitempty" yaml:"network_driver,omitempty"`

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

type NetworkDriverServiceCollection struct {
	Collection
	Data   []NetworkDriverService `json:"data,omitempty"`
	client *NetworkDriverServiceClient
}

type NetworkDriverServiceClient struct {
	rancherClient *RancherClient
}

type NetworkDriverServiceOperations interface {
	List(opts *ListOpts) (*NetworkDriverServiceCollection, error)
	Create(opts *NetworkDriverService) (*NetworkDriverService, error)
	Update(existing *NetworkDriverService, updates interface{}) (*NetworkDriverService, error)
	ById(id string) (*NetworkDriverService, error)
	Delete(container *NetworkDriverService) error

	ActionActivate(*NetworkDriverService) (*Service, error)

	ActionCancelupgrade(*NetworkDriverService) (*Service, error)

	ActionCreate(*NetworkDriverService) (*Service, error)

	ActionDeactivate(*NetworkDriverService) (*Service, error)

	ActionError(*NetworkDriverService) (*Service, error)

	ActionFinishupgrade(*NetworkDriverService) (*Service, error)

	ActionGarbagecollect(*NetworkDriverService) (*Service, error)

	ActionPause(*NetworkDriverService) (*Service, error)

	ActionRemove(*NetworkDriverService) (*Service, error)

	ActionRestart(*NetworkDriverService) (*Service, error)

	ActionRollback(*NetworkDriverService, *ServiceRollback) (*Service, error)

	ActionUpdate(*NetworkDriverService) (*Service, error)

	ActionUpgrade(*NetworkDriverService, *ServiceUpgrade) (*Service, error)
}

func newNetworkDriverServiceClient(rancherClient *RancherClient) *NetworkDriverServiceClient {
	return &NetworkDriverServiceClient{
		rancherClient: rancherClient,
	}
}

func (c *NetworkDriverServiceClient) Create(container *NetworkDriverService) (*NetworkDriverService, error) {
	resp := &NetworkDriverService{}
	err := c.rancherClient.doCreate(NETWORK_DRIVER_SERVICE_TYPE, container, resp)
	return resp, err
}

func (c *NetworkDriverServiceClient) Update(existing *NetworkDriverService, updates interface{}) (*NetworkDriverService, error) {
	resp := &NetworkDriverService{}
	err := c.rancherClient.doUpdate(NETWORK_DRIVER_SERVICE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *NetworkDriverServiceClient) List(opts *ListOpts) (*NetworkDriverServiceCollection, error) {
	resp := &NetworkDriverServiceCollection{}
	err := c.rancherClient.doList(NETWORK_DRIVER_SERVICE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *NetworkDriverServiceCollection) Next() (*NetworkDriverServiceCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &NetworkDriverServiceCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *NetworkDriverServiceClient) ById(id string) (*NetworkDriverService, error) {
	resp := &NetworkDriverService{}
	err := c.rancherClient.doById(NETWORK_DRIVER_SERVICE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *NetworkDriverServiceClient) Delete(container *NetworkDriverService) error {
	return c.rancherClient.doResourceDelete(NETWORK_DRIVER_SERVICE_TYPE, &container.Resource)
}

func (c *NetworkDriverServiceClient) ActionActivate(resource *NetworkDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "activate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionCancelupgrade(resource *NetworkDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "cancelupgrade", &resource.Resource, nil, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionCreate(resource *NetworkDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionDeactivate(resource *NetworkDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "deactivate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionError(resource *NetworkDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "error", &resource.Resource, nil, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionFinishupgrade(resource *NetworkDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "finishupgrade", &resource.Resource, nil, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionGarbagecollect(resource *NetworkDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "garbagecollect", &resource.Resource, nil, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionPause(resource *NetworkDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "pause", &resource.Resource, nil, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionRemove(resource *NetworkDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionRestart(resource *NetworkDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "restart", &resource.Resource, nil, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionRollback(resource *NetworkDriverService, input *ServiceRollback) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "rollback", &resource.Resource, input, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionUpdate(resource *NetworkDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "update", &resource.Resource, nil, resp)

	return resp, err
}

func (c *NetworkDriverServiceClient) ActionUpgrade(resource *NetworkDriverService, input *ServiceUpgrade) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(NETWORK_DRIVER_SERVICE_TYPE, "upgrade", &resource.Resource, input, resp)

	return resp, err
}

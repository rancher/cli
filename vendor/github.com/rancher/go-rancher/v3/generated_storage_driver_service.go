package client

const (
	STORAGE_DRIVER_SERVICE_TYPE = "storageDriverService"
)

type StorageDriverService struct {
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

	StorageDriver StorageDriver `json:"storageDriver,omitempty" yaml:"storage_driver,omitempty"`

	Transitioning string `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`

	TransitioningMessage string `json:"transitioningMessage,omitempty" yaml:"transitioning_message,omitempty"`

	Upgrade *ServiceUpgrade `json:"upgrade,omitempty" yaml:"upgrade,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`

	Vip string `json:"vip,omitempty" yaml:"vip,omitempty"`
}

type StorageDriverServiceCollection struct {
	Collection
	Data   []StorageDriverService `json:"data,omitempty"`
	client *StorageDriverServiceClient
}

type StorageDriverServiceClient struct {
	rancherClient *RancherClient
}

type StorageDriverServiceOperations interface {
	List(opts *ListOpts) (*StorageDriverServiceCollection, error)
	Create(opts *StorageDriverService) (*StorageDriverService, error)
	Update(existing *StorageDriverService, updates interface{}) (*StorageDriverService, error)
	ById(id string) (*StorageDriverService, error)
	Delete(container *StorageDriverService) error

	ActionActivate(*StorageDriverService) (*Service, error)

	ActionCancelupgrade(*StorageDriverService) (*Service, error)

	ActionCreate(*StorageDriverService) (*Service, error)

	ActionDeactivate(*StorageDriverService) (*Service, error)

	ActionError(*StorageDriverService) (*Service, error)

	ActionFinishupgrade(*StorageDriverService) (*Service, error)

	ActionGarbagecollect(*StorageDriverService) (*Service, error)

	ActionPause(*StorageDriverService) (*Service, error)

	ActionRemove(*StorageDriverService) (*Service, error)

	ActionRestart(*StorageDriverService) (*Service, error)

	ActionRollback(*StorageDriverService, *ServiceRollback) (*Service, error)

	ActionUpdate(*StorageDriverService) (*Service, error)

	ActionUpgrade(*StorageDriverService, *ServiceUpgrade) (*Service, error)
}

func newStorageDriverServiceClient(rancherClient *RancherClient) *StorageDriverServiceClient {
	return &StorageDriverServiceClient{
		rancherClient: rancherClient,
	}
}

func (c *StorageDriverServiceClient) Create(container *StorageDriverService) (*StorageDriverService, error) {
	resp := &StorageDriverService{}
	err := c.rancherClient.doCreate(STORAGE_DRIVER_SERVICE_TYPE, container, resp)
	return resp, err
}

func (c *StorageDriverServiceClient) Update(existing *StorageDriverService, updates interface{}) (*StorageDriverService, error) {
	resp := &StorageDriverService{}
	err := c.rancherClient.doUpdate(STORAGE_DRIVER_SERVICE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *StorageDriverServiceClient) List(opts *ListOpts) (*StorageDriverServiceCollection, error) {
	resp := &StorageDriverServiceCollection{}
	err := c.rancherClient.doList(STORAGE_DRIVER_SERVICE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *StorageDriverServiceCollection) Next() (*StorageDriverServiceCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &StorageDriverServiceCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *StorageDriverServiceClient) ById(id string) (*StorageDriverService, error) {
	resp := &StorageDriverService{}
	err := c.rancherClient.doById(STORAGE_DRIVER_SERVICE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *StorageDriverServiceClient) Delete(container *StorageDriverService) error {
	return c.rancherClient.doResourceDelete(STORAGE_DRIVER_SERVICE_TYPE, &container.Resource)
}

func (c *StorageDriverServiceClient) ActionActivate(resource *StorageDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(STORAGE_DRIVER_SERVICE_TYPE, "activate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *StorageDriverServiceClient) ActionCancelupgrade(resource *StorageDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(STORAGE_DRIVER_SERVICE_TYPE, "cancelupgrade", &resource.Resource, nil, resp)

	return resp, err
}

func (c *StorageDriverServiceClient) ActionCreate(resource *StorageDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(STORAGE_DRIVER_SERVICE_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *StorageDriverServiceClient) ActionDeactivate(resource *StorageDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(STORAGE_DRIVER_SERVICE_TYPE, "deactivate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *StorageDriverServiceClient) ActionError(resource *StorageDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(STORAGE_DRIVER_SERVICE_TYPE, "error", &resource.Resource, nil, resp)

	return resp, err
}

func (c *StorageDriverServiceClient) ActionFinishupgrade(resource *StorageDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(STORAGE_DRIVER_SERVICE_TYPE, "finishupgrade", &resource.Resource, nil, resp)

	return resp, err
}

func (c *StorageDriverServiceClient) ActionGarbagecollect(resource *StorageDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(STORAGE_DRIVER_SERVICE_TYPE, "garbagecollect", &resource.Resource, nil, resp)

	return resp, err
}

func (c *StorageDriverServiceClient) ActionPause(resource *StorageDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(STORAGE_DRIVER_SERVICE_TYPE, "pause", &resource.Resource, nil, resp)

	return resp, err
}

func (c *StorageDriverServiceClient) ActionRemove(resource *StorageDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(STORAGE_DRIVER_SERVICE_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}

func (c *StorageDriverServiceClient) ActionRestart(resource *StorageDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(STORAGE_DRIVER_SERVICE_TYPE, "restart", &resource.Resource, nil, resp)

	return resp, err
}

func (c *StorageDriverServiceClient) ActionRollback(resource *StorageDriverService, input *ServiceRollback) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(STORAGE_DRIVER_SERVICE_TYPE, "rollback", &resource.Resource, input, resp)

	return resp, err
}

func (c *StorageDriverServiceClient) ActionUpdate(resource *StorageDriverService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(STORAGE_DRIVER_SERVICE_TYPE, "update", &resource.Resource, nil, resp)

	return resp, err
}

func (c *StorageDriverServiceClient) ActionUpgrade(resource *StorageDriverService, input *ServiceUpgrade) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(STORAGE_DRIVER_SERVICE_TYPE, "upgrade", &resource.Resource, input, resp)

	return resp, err
}

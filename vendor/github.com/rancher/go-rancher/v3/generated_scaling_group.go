package client

const (
	SCALING_GROUP_TYPE = "scalingGroup"
)

type ScalingGroup struct {
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

type ScalingGroupCollection struct {
	Collection
	Data   []ScalingGroup `json:"data,omitempty"`
	client *ScalingGroupClient
}

type ScalingGroupClient struct {
	rancherClient *RancherClient
}

type ScalingGroupOperations interface {
	List(opts *ListOpts) (*ScalingGroupCollection, error)
	Create(opts *ScalingGroup) (*ScalingGroup, error)
	Update(existing *ScalingGroup, updates interface{}) (*ScalingGroup, error)
	ById(id string) (*ScalingGroup, error)
	Delete(container *ScalingGroup) error

	ActionActivate(*ScalingGroup) (*Service, error)

	ActionCancelupgrade(*ScalingGroup) (*Service, error)

	ActionCreate(*ScalingGroup) (*Service, error)

	ActionDeactivate(*ScalingGroup) (*Service, error)

	ActionError(*ScalingGroup) (*Service, error)

	ActionFinishupgrade(*ScalingGroup) (*Service, error)

	ActionGarbagecollect(*ScalingGroup) (*Service, error)

	ActionPause(*ScalingGroup) (*Service, error)

	ActionRemove(*ScalingGroup) (*Service, error)

	ActionRestart(*ScalingGroup) (*Service, error)

	ActionRollback(*ScalingGroup, *ServiceRollback) (*Service, error)

	ActionUpdate(*ScalingGroup) (*Service, error)

	ActionUpgrade(*ScalingGroup, *ServiceUpgrade) (*Service, error)
}

func newScalingGroupClient(rancherClient *RancherClient) *ScalingGroupClient {
	return &ScalingGroupClient{
		rancherClient: rancherClient,
	}
}

func (c *ScalingGroupClient) Create(container *ScalingGroup) (*ScalingGroup, error) {
	resp := &ScalingGroup{}
	err := c.rancherClient.doCreate(SCALING_GROUP_TYPE, container, resp)
	return resp, err
}

func (c *ScalingGroupClient) Update(existing *ScalingGroup, updates interface{}) (*ScalingGroup, error) {
	resp := &ScalingGroup{}
	err := c.rancherClient.doUpdate(SCALING_GROUP_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ScalingGroupClient) List(opts *ListOpts) (*ScalingGroupCollection, error) {
	resp := &ScalingGroupCollection{}
	err := c.rancherClient.doList(SCALING_GROUP_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ScalingGroupCollection) Next() (*ScalingGroupCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ScalingGroupCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ScalingGroupClient) ById(id string) (*ScalingGroup, error) {
	resp := &ScalingGroup{}
	err := c.rancherClient.doById(SCALING_GROUP_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *ScalingGroupClient) Delete(container *ScalingGroup) error {
	return c.rancherClient.doResourceDelete(SCALING_GROUP_TYPE, &container.Resource)
}

func (c *ScalingGroupClient) ActionActivate(resource *ScalingGroup) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SCALING_GROUP_TYPE, "activate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ScalingGroupClient) ActionCancelupgrade(resource *ScalingGroup) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SCALING_GROUP_TYPE, "cancelupgrade", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ScalingGroupClient) ActionCreate(resource *ScalingGroup) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SCALING_GROUP_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ScalingGroupClient) ActionDeactivate(resource *ScalingGroup) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SCALING_GROUP_TYPE, "deactivate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ScalingGroupClient) ActionError(resource *ScalingGroup) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SCALING_GROUP_TYPE, "error", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ScalingGroupClient) ActionFinishupgrade(resource *ScalingGroup) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SCALING_GROUP_TYPE, "finishupgrade", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ScalingGroupClient) ActionGarbagecollect(resource *ScalingGroup) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SCALING_GROUP_TYPE, "garbagecollect", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ScalingGroupClient) ActionPause(resource *ScalingGroup) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SCALING_GROUP_TYPE, "pause", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ScalingGroupClient) ActionRemove(resource *ScalingGroup) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SCALING_GROUP_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ScalingGroupClient) ActionRestart(resource *ScalingGroup) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SCALING_GROUP_TYPE, "restart", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ScalingGroupClient) ActionRollback(resource *ScalingGroup, input *ServiceRollback) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SCALING_GROUP_TYPE, "rollback", &resource.Resource, input, resp)

	return resp, err
}

func (c *ScalingGroupClient) ActionUpdate(resource *ScalingGroup) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SCALING_GROUP_TYPE, "update", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ScalingGroupClient) ActionUpgrade(resource *ScalingGroup, input *ServiceUpgrade) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(SCALING_GROUP_TYPE, "upgrade", &resource.Resource, input, resp)

	return resp, err
}

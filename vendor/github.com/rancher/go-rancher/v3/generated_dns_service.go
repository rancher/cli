package client

const (
	DNS_SERVICE_TYPE = "dnsService"
)

type DnsService struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	AssignServiceIpAddress bool `json:"assignServiceIpAddress,omitempty" yaml:"assign_service_ip_address,omitempty"`

	ClusterId string `json:"clusterId,omitempty" yaml:"cluster_id,omitempty"`

	CompleteUpdate bool `json:"completeUpdate,omitempty" yaml:"complete_update,omitempty"`

	CreateOnly bool `json:"createOnly,omitempty" yaml:"create_only,omitempty"`

	Created string `json:"created,omitempty" yaml:"created,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	ExternalId string `json:"externalId,omitempty" yaml:"external_id,omitempty"`

	Fqdn string `json:"fqdn,omitempty" yaml:"fqdn,omitempty"`

	HealthState string `json:"healthState,omitempty" yaml:"health_state,omitempty"`

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

type DnsServiceCollection struct {
	Collection
	Data   []DnsService `json:"data,omitempty"`
	client *DnsServiceClient
}

type DnsServiceClient struct {
	rancherClient *RancherClient
}

type DnsServiceOperations interface {
	List(opts *ListOpts) (*DnsServiceCollection, error)
	Create(opts *DnsService) (*DnsService, error)
	Update(existing *DnsService, updates interface{}) (*DnsService, error)
	ById(id string) (*DnsService, error)
	Delete(container *DnsService) error

	ActionActivate(*DnsService) (*Service, error)

	ActionCancelupgrade(*DnsService) (*Service, error)

	ActionCreate(*DnsService) (*Service, error)

	ActionDeactivate(*DnsService) (*Service, error)

	ActionError(*DnsService) (*Service, error)

	ActionFinishupgrade(*DnsService) (*Service, error)

	ActionGarbagecollect(*DnsService) (*Service, error)

	ActionPause(*DnsService) (*Service, error)

	ActionRemove(*DnsService) (*Service, error)

	ActionRestart(*DnsService) (*Service, error)

	ActionRollback(*DnsService, *ServiceRollback) (*Service, error)

	ActionUpdate(*DnsService) (*Service, error)

	ActionUpgrade(*DnsService, *ServiceUpgrade) (*Service, error)
}

func newDnsServiceClient(rancherClient *RancherClient) *DnsServiceClient {
	return &DnsServiceClient{
		rancherClient: rancherClient,
	}
}

func (c *DnsServiceClient) Create(container *DnsService) (*DnsService, error) {
	resp := &DnsService{}
	err := c.rancherClient.doCreate(DNS_SERVICE_TYPE, container, resp)
	return resp, err
}

func (c *DnsServiceClient) Update(existing *DnsService, updates interface{}) (*DnsService, error) {
	resp := &DnsService{}
	err := c.rancherClient.doUpdate(DNS_SERVICE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *DnsServiceClient) List(opts *ListOpts) (*DnsServiceCollection, error) {
	resp := &DnsServiceCollection{}
	err := c.rancherClient.doList(DNS_SERVICE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *DnsServiceCollection) Next() (*DnsServiceCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &DnsServiceCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *DnsServiceClient) ById(id string) (*DnsService, error) {
	resp := &DnsService{}
	err := c.rancherClient.doById(DNS_SERVICE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *DnsServiceClient) Delete(container *DnsService) error {
	return c.rancherClient.doResourceDelete(DNS_SERVICE_TYPE, &container.Resource)
}

func (c *DnsServiceClient) ActionActivate(resource *DnsService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(DNS_SERVICE_TYPE, "activate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *DnsServiceClient) ActionCancelupgrade(resource *DnsService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(DNS_SERVICE_TYPE, "cancelupgrade", &resource.Resource, nil, resp)

	return resp, err
}

func (c *DnsServiceClient) ActionCreate(resource *DnsService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(DNS_SERVICE_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *DnsServiceClient) ActionDeactivate(resource *DnsService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(DNS_SERVICE_TYPE, "deactivate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *DnsServiceClient) ActionError(resource *DnsService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(DNS_SERVICE_TYPE, "error", &resource.Resource, nil, resp)

	return resp, err
}

func (c *DnsServiceClient) ActionFinishupgrade(resource *DnsService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(DNS_SERVICE_TYPE, "finishupgrade", &resource.Resource, nil, resp)

	return resp, err
}

func (c *DnsServiceClient) ActionGarbagecollect(resource *DnsService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(DNS_SERVICE_TYPE, "garbagecollect", &resource.Resource, nil, resp)

	return resp, err
}

func (c *DnsServiceClient) ActionPause(resource *DnsService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(DNS_SERVICE_TYPE, "pause", &resource.Resource, nil, resp)

	return resp, err
}

func (c *DnsServiceClient) ActionRemove(resource *DnsService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(DNS_SERVICE_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}

func (c *DnsServiceClient) ActionRestart(resource *DnsService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(DNS_SERVICE_TYPE, "restart", &resource.Resource, nil, resp)

	return resp, err
}

func (c *DnsServiceClient) ActionRollback(resource *DnsService, input *ServiceRollback) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(DNS_SERVICE_TYPE, "rollback", &resource.Resource, input, resp)

	return resp, err
}

func (c *DnsServiceClient) ActionUpdate(resource *DnsService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(DNS_SERVICE_TYPE, "update", &resource.Resource, nil, resp)

	return resp, err
}

func (c *DnsServiceClient) ActionUpgrade(resource *DnsService, input *ServiceUpgrade) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(DNS_SERVICE_TYPE, "upgrade", &resource.Resource, input, resp)

	return resp, err
}

package client

const (
	CLUSTER_TYPE = "cluster"
)

type Cluster struct {
	Resource

	Created string `json:"created,omitempty" yaml:"created,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	Embedded bool `json:"embedded,omitempty" yaml:"embedded,omitempty"`

	HostRemoveDelaySeconds int64 `json:"hostRemoveDelaySeconds,omitempty" yaml:"host_remove_delay_seconds,omitempty"`

	Identity ClusterIdentity `json:"identity,omitempty" yaml:"identity,omitempty"`

	K8sClientConfig *K8sClientConfig `json:"k8sClientConfig,omitempty" yaml:"k8s_client_config,omitempty"`

	K8sServerConfig *K8sServerConfig `json:"k8sServerConfig,omitempty" yaml:"k8s_server_config,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	Orchestration string `json:"orchestration,omitempty" yaml:"orchestration,omitempty"`

	RegistrationToken *RegistrationToken `json:"registrationToken,omitempty" yaml:"registration_token,omitempty"`

	RemoveTime string `json:"removeTime,omitempty" yaml:"remove_time,omitempty"`

	Removed string `json:"removed,omitempty" yaml:"removed,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	SystemStacks []Stack `json:"systemStacks,omitempty" yaml:"system_stacks,omitempty"`

	Transitioning string `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`

	TransitioningMessage string `json:"transitioningMessage,omitempty" yaml:"transitioning_message,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type ClusterCollection struct {
	Collection
	Data   []Cluster `json:"data,omitempty"`
	client *ClusterClient
}

type ClusterClient struct {
	rancherClient *RancherClient
}

type ClusterOperations interface {
	List(opts *ListOpts) (*ClusterCollection, error)
	Create(opts *Cluster) (*Cluster, error)
	Update(existing *Cluster, updates interface{}) (*Cluster, error)
	ById(id string) (*Cluster, error)
	Delete(container *Cluster) error

	ActionActivate(*Cluster) (*Cluster, error)

	ActionCreate(*Cluster) (*Cluster, error)

	ActionError(*Cluster) (*Cluster, error)

	ActionRemove(*Cluster) (*Cluster, error)

	ActionUpdate(*Cluster) (*Cluster, error)
}

func newClusterClient(rancherClient *RancherClient) *ClusterClient {
	return &ClusterClient{
		rancherClient: rancherClient,
	}
}

func (c *ClusterClient) Create(container *Cluster) (*Cluster, error) {
	resp := &Cluster{}
	err := c.rancherClient.doCreate(CLUSTER_TYPE, container, resp)
	return resp, err
}

func (c *ClusterClient) Update(existing *Cluster, updates interface{}) (*Cluster, error) {
	resp := &Cluster{}
	err := c.rancherClient.doUpdate(CLUSTER_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ClusterClient) List(opts *ListOpts) (*ClusterCollection, error) {
	resp := &ClusterCollection{}
	err := c.rancherClient.doList(CLUSTER_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ClusterCollection) Next() (*ClusterCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ClusterCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ClusterClient) ById(id string) (*Cluster, error) {
	resp := &Cluster{}
	err := c.rancherClient.doById(CLUSTER_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *ClusterClient) Delete(container *Cluster) error {
	return c.rancherClient.doResourceDelete(CLUSTER_TYPE, &container.Resource)
}

func (c *ClusterClient) ActionActivate(resource *Cluster) (*Cluster, error) {

	resp := &Cluster{}

	err := c.rancherClient.doAction(CLUSTER_TYPE, "activate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ClusterClient) ActionCreate(resource *Cluster) (*Cluster, error) {

	resp := &Cluster{}

	err := c.rancherClient.doAction(CLUSTER_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ClusterClient) ActionError(resource *Cluster) (*Cluster, error) {

	resp := &Cluster{}

	err := c.rancherClient.doAction(CLUSTER_TYPE, "error", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ClusterClient) ActionRemove(resource *Cluster) (*Cluster, error) {

	resp := &Cluster{}

	err := c.rancherClient.doAction(CLUSTER_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ClusterClient) ActionUpdate(resource *Cluster) (*Cluster, error) {

	resp := &Cluster{}

	err := c.rancherClient.doAction(CLUSTER_TYPE, "update", &resource.Resource, nil, resp)

	return resp, err
}

package client

const (
	CONTAINER_EVENT_TYPE = "containerEvent"
)

type ContainerEvent struct {
	Resource

	ClusterId int64 `json:"clusterId,omitempty" yaml:"cluster_id,omitempty"`

	ContainerUuid string `json:"containerUuid,omitempty" yaml:"container_uuid,omitempty"`

	DockerInspect interface{} `json:"dockerInspect,omitempty" yaml:"docker_inspect,omitempty"`

	ExternalId string `json:"externalId,omitempty" yaml:"external_id,omitempty"`

	ExternalStatus string `json:"externalStatus,omitempty" yaml:"external_status,omitempty"`

	HostId string `json:"hostId,omitempty" yaml:"host_id,omitempty"`

	ReportedHostUuid string `json:"reportedHostUuid,omitempty" yaml:"reported_host_uuid,omitempty"`
}

type ContainerEventCollection struct {
	Collection
	Data   []ContainerEvent `json:"data,omitempty"`
	client *ContainerEventClient
}

type ContainerEventClient struct {
	rancherClient *RancherClient
}

type ContainerEventOperations interface {
	List(opts *ListOpts) (*ContainerEventCollection, error)
	Create(opts *ContainerEvent) (*ContainerEvent, error)
	Update(existing *ContainerEvent, updates interface{}) (*ContainerEvent, error)
	ById(id string) (*ContainerEvent, error)
	Delete(container *ContainerEvent) error
}

func newContainerEventClient(rancherClient *RancherClient) *ContainerEventClient {
	return &ContainerEventClient{
		rancherClient: rancherClient,
	}
}

func (c *ContainerEventClient) Create(container *ContainerEvent) (*ContainerEvent, error) {
	resp := &ContainerEvent{}
	err := c.rancherClient.doCreate(CONTAINER_EVENT_TYPE, container, resp)
	return resp, err
}

func (c *ContainerEventClient) Update(existing *ContainerEvent, updates interface{}) (*ContainerEvent, error) {
	resp := &ContainerEvent{}
	err := c.rancherClient.doUpdate(CONTAINER_EVENT_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ContainerEventClient) List(opts *ListOpts) (*ContainerEventCollection, error) {
	resp := &ContainerEventCollection{}
	err := c.rancherClient.doList(CONTAINER_EVENT_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ContainerEventCollection) Next() (*ContainerEventCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ContainerEventCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ContainerEventClient) ById(id string) (*ContainerEvent, error) {
	resp := &ContainerEvent{}
	err := c.rancherClient.doById(CONTAINER_EVENT_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *ContainerEventClient) Delete(container *ContainerEvent) error {
	return c.rancherClient.doResourceDelete(CONTAINER_EVENT_TYPE, &container.Resource)
}

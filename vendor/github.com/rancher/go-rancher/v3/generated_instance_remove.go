package client

const (
	INSTANCE_REMOVE_TYPE = "instanceRemove"
)

type InstanceRemove struct {
	Resource

	RemoveSource string `json:"removeSource,omitempty" yaml:"remove_source,omitempty"`
}

type InstanceRemoveCollection struct {
	Collection
	Data   []InstanceRemove `json:"data,omitempty"`
	client *InstanceRemoveClient
}

type InstanceRemoveClient struct {
	rancherClient *RancherClient
}

type InstanceRemoveOperations interface {
	List(opts *ListOpts) (*InstanceRemoveCollection, error)
	Create(opts *InstanceRemove) (*InstanceRemove, error)
	Update(existing *InstanceRemove, updates interface{}) (*InstanceRemove, error)
	ById(id string) (*InstanceRemove, error)
	Delete(container *InstanceRemove) error
}

func newInstanceRemoveClient(rancherClient *RancherClient) *InstanceRemoveClient {
	return &InstanceRemoveClient{
		rancherClient: rancherClient,
	}
}

func (c *InstanceRemoveClient) Create(container *InstanceRemove) (*InstanceRemove, error) {
	resp := &InstanceRemove{}
	err := c.rancherClient.doCreate(INSTANCE_REMOVE_TYPE, container, resp)
	return resp, err
}

func (c *InstanceRemoveClient) Update(existing *InstanceRemove, updates interface{}) (*InstanceRemove, error) {
	resp := &InstanceRemove{}
	err := c.rancherClient.doUpdate(INSTANCE_REMOVE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *InstanceRemoveClient) List(opts *ListOpts) (*InstanceRemoveCollection, error) {
	resp := &InstanceRemoveCollection{}
	err := c.rancherClient.doList(INSTANCE_REMOVE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *InstanceRemoveCollection) Next() (*InstanceRemoveCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &InstanceRemoveCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *InstanceRemoveClient) ById(id string) (*InstanceRemove, error) {
	resp := &InstanceRemove{}
	err := c.rancherClient.doById(INSTANCE_REMOVE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *InstanceRemoveClient) Delete(container *InstanceRemove) error {
	return c.rancherClient.doResourceDelete(INSTANCE_REMOVE_TYPE, &container.Resource)
}

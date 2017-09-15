package client

const (
	CONTAINER_UPGRADE_TYPE = "containerUpgrade"
)

type ContainerUpgrade struct {
	Resource

	Config ContainerConfig `json:"config,omitempty" yaml:"config,omitempty"`
}

type ContainerUpgradeCollection struct {
	Collection
	Data   []ContainerUpgrade `json:"data,omitempty"`
	client *ContainerUpgradeClient
}

type ContainerUpgradeClient struct {
	rancherClient *RancherClient
}

type ContainerUpgradeOperations interface {
	List(opts *ListOpts) (*ContainerUpgradeCollection, error)
	Create(opts *ContainerUpgrade) (*ContainerUpgrade, error)
	Update(existing *ContainerUpgrade, updates interface{}) (*ContainerUpgrade, error)
	ById(id string) (*ContainerUpgrade, error)
	Delete(container *ContainerUpgrade) error
}

func newContainerUpgradeClient(rancherClient *RancherClient) *ContainerUpgradeClient {
	return &ContainerUpgradeClient{
		rancherClient: rancherClient,
	}
}

func (c *ContainerUpgradeClient) Create(container *ContainerUpgrade) (*ContainerUpgrade, error) {
	resp := &ContainerUpgrade{}
	err := c.rancherClient.doCreate(CONTAINER_UPGRADE_TYPE, container, resp)
	return resp, err
}

func (c *ContainerUpgradeClient) Update(existing *ContainerUpgrade, updates interface{}) (*ContainerUpgrade, error) {
	resp := &ContainerUpgrade{}
	err := c.rancherClient.doUpdate(CONTAINER_UPGRADE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ContainerUpgradeClient) List(opts *ListOpts) (*ContainerUpgradeCollection, error) {
	resp := &ContainerUpgradeCollection{}
	err := c.rancherClient.doList(CONTAINER_UPGRADE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ContainerUpgradeCollection) Next() (*ContainerUpgradeCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ContainerUpgradeCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ContainerUpgradeClient) ById(id string) (*ContainerUpgrade, error) {
	resp := &ContainerUpgrade{}
	err := c.rancherClient.doById(CONTAINER_UPGRADE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *ContainerUpgradeClient) Delete(container *ContainerUpgrade) error {
	return c.rancherClient.doResourceDelete(CONTAINER_UPGRADE_TYPE, &container.Resource)
}

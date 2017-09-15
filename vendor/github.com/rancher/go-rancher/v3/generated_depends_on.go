package client

const (
	DEPENDS_ON_TYPE = "dependsOn"
)

type DependsOn struct {
	Resource

	Condition string `json:"condition,omitempty" yaml:"condition,omitempty"`

	Container string `json:"container,omitempty" yaml:"container,omitempty"`

	Service string `json:"service,omitempty" yaml:"service,omitempty"`
}

type DependsOnCollection struct {
	Collection
	Data   []DependsOn `json:"data,omitempty"`
	client *DependsOnClient
}

type DependsOnClient struct {
	rancherClient *RancherClient
}

type DependsOnOperations interface {
	List(opts *ListOpts) (*DependsOnCollection, error)
	Create(opts *DependsOn) (*DependsOn, error)
	Update(existing *DependsOn, updates interface{}) (*DependsOn, error)
	ById(id string) (*DependsOn, error)
	Delete(container *DependsOn) error
}

func newDependsOnClient(rancherClient *RancherClient) *DependsOnClient {
	return &DependsOnClient{
		rancherClient: rancherClient,
	}
}

func (c *DependsOnClient) Create(container *DependsOn) (*DependsOn, error) {
	resp := &DependsOn{}
	err := c.rancherClient.doCreate(DEPENDS_ON_TYPE, container, resp)
	return resp, err
}

func (c *DependsOnClient) Update(existing *DependsOn, updates interface{}) (*DependsOn, error) {
	resp := &DependsOn{}
	err := c.rancherClient.doUpdate(DEPENDS_ON_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *DependsOnClient) List(opts *ListOpts) (*DependsOnCollection, error) {
	resp := &DependsOnCollection{}
	err := c.rancherClient.doList(DEPENDS_ON_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *DependsOnCollection) Next() (*DependsOnCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &DependsOnCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *DependsOnClient) ById(id string) (*DependsOn, error) {
	resp := &DependsOn{}
	err := c.rancherClient.doById(DEPENDS_ON_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *DependsOnClient) Delete(container *DependsOn) error {
	return c.rancherClient.doResourceDelete(DEPENDS_ON_TYPE, &container.Resource)
}

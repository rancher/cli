package client

const (
	SET_COMPUTE_FLAVOR_INPUT_TYPE = "setComputeFlavorInput"
)

type SetComputeFlavorInput struct {
	Resource

	ComputeFlavor string `json:"computeFlavor,omitempty" yaml:"compute_flavor,omitempty"`
}

type SetComputeFlavorInputCollection struct {
	Collection
	Data   []SetComputeFlavorInput `json:"data,omitempty"`
	client *SetComputeFlavorInputClient
}

type SetComputeFlavorInputClient struct {
	rancherClient *RancherClient
}

type SetComputeFlavorInputOperations interface {
	List(opts *ListOpts) (*SetComputeFlavorInputCollection, error)
	Create(opts *SetComputeFlavorInput) (*SetComputeFlavorInput, error)
	Update(existing *SetComputeFlavorInput, updates interface{}) (*SetComputeFlavorInput, error)
	ById(id string) (*SetComputeFlavorInput, error)
	Delete(container *SetComputeFlavorInput) error
}

func newSetComputeFlavorInputClient(rancherClient *RancherClient) *SetComputeFlavorInputClient {
	return &SetComputeFlavorInputClient{
		rancherClient: rancherClient,
	}
}

func (c *SetComputeFlavorInputClient) Create(container *SetComputeFlavorInput) (*SetComputeFlavorInput, error) {
	resp := &SetComputeFlavorInput{}
	err := c.rancherClient.doCreate(SET_COMPUTE_FLAVOR_INPUT_TYPE, container, resp)
	return resp, err
}

func (c *SetComputeFlavorInputClient) Update(existing *SetComputeFlavorInput, updates interface{}) (*SetComputeFlavorInput, error) {
	resp := &SetComputeFlavorInput{}
	err := c.rancherClient.doUpdate(SET_COMPUTE_FLAVOR_INPUT_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *SetComputeFlavorInputClient) List(opts *ListOpts) (*SetComputeFlavorInputCollection, error) {
	resp := &SetComputeFlavorInputCollection{}
	err := c.rancherClient.doList(SET_COMPUTE_FLAVOR_INPUT_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *SetComputeFlavorInputCollection) Next() (*SetComputeFlavorInputCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &SetComputeFlavorInputCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *SetComputeFlavorInputClient) ById(id string) (*SetComputeFlavorInput, error) {
	resp := &SetComputeFlavorInput{}
	err := c.rancherClient.doById(SET_COMPUTE_FLAVOR_INPUT_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *SetComputeFlavorInputClient) Delete(container *SetComputeFlavorInput) error {
	return c.rancherClient.doResourceDelete(SET_COMPUTE_FLAVOR_INPUT_TYPE, &container.Resource)
}

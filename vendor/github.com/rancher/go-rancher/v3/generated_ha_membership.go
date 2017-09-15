package client

const (
	HA_MEMBERSHIP_TYPE = "haMembership"
)

type HaMembership struct {
	Resource

	Clustered bool `json:"clustered,omitempty" yaml:"clustered,omitempty"`

	Config string `json:"config,omitempty" yaml:"config,omitempty"`

	Heartbeat int64 `json:"heartbeat,omitempty" yaml:"heartbeat,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type HaMembershipCollection struct {
	Collection
	Data   []HaMembership `json:"data,omitempty"`
	client *HaMembershipClient
}

type HaMembershipClient struct {
	rancherClient *RancherClient
}

type HaMembershipOperations interface {
	List(opts *ListOpts) (*HaMembershipCollection, error)
	Create(opts *HaMembership) (*HaMembership, error)
	Update(existing *HaMembership, updates interface{}) (*HaMembership, error)
	ById(id string) (*HaMembership, error)
	Delete(container *HaMembership) error
}

func newHaMembershipClient(rancherClient *RancherClient) *HaMembershipClient {
	return &HaMembershipClient{
		rancherClient: rancherClient,
	}
}

func (c *HaMembershipClient) Create(container *HaMembership) (*HaMembership, error) {
	resp := &HaMembership{}
	err := c.rancherClient.doCreate(HA_MEMBERSHIP_TYPE, container, resp)
	return resp, err
}

func (c *HaMembershipClient) Update(existing *HaMembership, updates interface{}) (*HaMembership, error) {
	resp := &HaMembership{}
	err := c.rancherClient.doUpdate(HA_MEMBERSHIP_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *HaMembershipClient) List(opts *ListOpts) (*HaMembershipCollection, error) {
	resp := &HaMembershipCollection{}
	err := c.rancherClient.doList(HA_MEMBERSHIP_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *HaMembershipCollection) Next() (*HaMembershipCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &HaMembershipCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *HaMembershipClient) ById(id string) (*HaMembership, error) {
	resp := &HaMembership{}
	err := c.rancherClient.doById(HA_MEMBERSHIP_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *HaMembershipClient) Delete(container *HaMembership) error {
	return c.rancherClient.doResourceDelete(HA_MEMBERSHIP_TYPE, &container.Resource)
}

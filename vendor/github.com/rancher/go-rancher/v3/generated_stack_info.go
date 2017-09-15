package client

const (
	STACK_INFO_TYPE = "stackInfo"
)

type StackInfo struct {
	Resource

	EnvironmentUuid string `json:"environmentUuid,omitempty" yaml:"environment_uuid,omitempty"`

	HealthState string `json:"healthState,omitempty" yaml:"health_state,omitempty"`

	InfoType string `json:"infoType,omitempty" yaml:"info_type,omitempty"`

	InfoTypeId string `json:"infoTypeId,omitempty" yaml:"info_type_id,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type StackInfoCollection struct {
	Collection
	Data   []StackInfo `json:"data,omitempty"`
	client *StackInfoClient
}

type StackInfoClient struct {
	rancherClient *RancherClient
}

type StackInfoOperations interface {
	List(opts *ListOpts) (*StackInfoCollection, error)
	Create(opts *StackInfo) (*StackInfo, error)
	Update(existing *StackInfo, updates interface{}) (*StackInfo, error)
	ById(id string) (*StackInfo, error)
	Delete(container *StackInfo) error
}

func newStackInfoClient(rancherClient *RancherClient) *StackInfoClient {
	return &StackInfoClient{
		rancherClient: rancherClient,
	}
}

func (c *StackInfoClient) Create(container *StackInfo) (*StackInfo, error) {
	resp := &StackInfo{}
	err := c.rancherClient.doCreate(STACK_INFO_TYPE, container, resp)
	return resp, err
}

func (c *StackInfoClient) Update(existing *StackInfo, updates interface{}) (*StackInfo, error) {
	resp := &StackInfo{}
	err := c.rancherClient.doUpdate(STACK_INFO_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *StackInfoClient) List(opts *ListOpts) (*StackInfoCollection, error) {
	resp := &StackInfoCollection{}
	err := c.rancherClient.doList(STACK_INFO_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *StackInfoCollection) Next() (*StackInfoCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &StackInfoCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *StackInfoClient) ById(id string) (*StackInfo, error) {
	resp := &StackInfo{}
	err := c.rancherClient.doById(STACK_INFO_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *StackInfoClient) Delete(container *StackInfo) error {
	return c.rancherClient.doResourceDelete(STACK_INFO_TYPE, &container.Resource)
}

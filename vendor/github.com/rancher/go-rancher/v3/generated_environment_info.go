package client

const (
	ENVIRONMENT_INFO_TYPE = "environmentInfo"
)

type EnvironmentInfo struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	EnvironmentUuid string `json:"environmentUuid,omitempty" yaml:"environment_uuid,omitempty"`

	ExternalId string `json:"externalId,omitempty" yaml:"external_id,omitempty"`

	InfoType string `json:"infoType,omitempty" yaml:"info_type,omitempty"`

	InfoTypeId string `json:"infoTypeId,omitempty" yaml:"info_type_id,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	System bool `json:"system,omitempty" yaml:"system,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type EnvironmentInfoCollection struct {
	Collection
	Data   []EnvironmentInfo `json:"data,omitempty"`
	client *EnvironmentInfoClient
}

type EnvironmentInfoClient struct {
	rancherClient *RancherClient
}

type EnvironmentInfoOperations interface {
	List(opts *ListOpts) (*EnvironmentInfoCollection, error)
	Create(opts *EnvironmentInfo) (*EnvironmentInfo, error)
	Update(existing *EnvironmentInfo, updates interface{}) (*EnvironmentInfo, error)
	ById(id string) (*EnvironmentInfo, error)
	Delete(container *EnvironmentInfo) error
}

func newEnvironmentInfoClient(rancherClient *RancherClient) *EnvironmentInfoClient {
	return &EnvironmentInfoClient{
		rancherClient: rancherClient,
	}
}

func (c *EnvironmentInfoClient) Create(container *EnvironmentInfo) (*EnvironmentInfo, error) {
	resp := &EnvironmentInfo{}
	err := c.rancherClient.doCreate(ENVIRONMENT_INFO_TYPE, container, resp)
	return resp, err
}

func (c *EnvironmentInfoClient) Update(existing *EnvironmentInfo, updates interface{}) (*EnvironmentInfo, error) {
	resp := &EnvironmentInfo{}
	err := c.rancherClient.doUpdate(ENVIRONMENT_INFO_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *EnvironmentInfoClient) List(opts *ListOpts) (*EnvironmentInfoCollection, error) {
	resp := &EnvironmentInfoCollection{}
	err := c.rancherClient.doList(ENVIRONMENT_INFO_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *EnvironmentInfoCollection) Next() (*EnvironmentInfoCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &EnvironmentInfoCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *EnvironmentInfoClient) ById(id string) (*EnvironmentInfo, error) {
	resp := &EnvironmentInfo{}
	err := c.rancherClient.doById(ENVIRONMENT_INFO_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *EnvironmentInfoClient) Delete(container *EnvironmentInfo) error {
	return c.rancherClient.doResourceDelete(ENVIRONMENT_INFO_TYPE, &container.Resource)
}

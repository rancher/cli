package client

const (
	METADATA_OBJECT_TYPE = "metadataObject"
)

type MetadataObject struct {
	Resource

	EnvironmentUuid string `json:"environmentUuid,omitempty" yaml:"environment_uuid,omitempty"`

	InfoType string `json:"infoType,omitempty" yaml:"info_type,omitempty"`

	InfoTypeId int64 `json:"infoTypeId,omitempty" yaml:"info_type_id,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type MetadataObjectCollection struct {
	Collection
	Data   []MetadataObject `json:"data,omitempty"`
	client *MetadataObjectClient
}

type MetadataObjectClient struct {
	rancherClient *RancherClient
}

type MetadataObjectOperations interface {
	List(opts *ListOpts) (*MetadataObjectCollection, error)
	Create(opts *MetadataObject) (*MetadataObject, error)
	Update(existing *MetadataObject, updates interface{}) (*MetadataObject, error)
	ById(id string) (*MetadataObject, error)
	Delete(container *MetadataObject) error
}

func newMetadataObjectClient(rancherClient *RancherClient) *MetadataObjectClient {
	return &MetadataObjectClient{
		rancherClient: rancherClient,
	}
}

func (c *MetadataObjectClient) Create(container *MetadataObject) (*MetadataObject, error) {
	resp := &MetadataObject{}
	err := c.rancherClient.doCreate(METADATA_OBJECT_TYPE, container, resp)
	return resp, err
}

func (c *MetadataObjectClient) Update(existing *MetadataObject, updates interface{}) (*MetadataObject, error) {
	resp := &MetadataObject{}
	err := c.rancherClient.doUpdate(METADATA_OBJECT_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *MetadataObjectClient) List(opts *ListOpts) (*MetadataObjectCollection, error) {
	resp := &MetadataObjectCollection{}
	err := c.rancherClient.doList(METADATA_OBJECT_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *MetadataObjectCollection) Next() (*MetadataObjectCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &MetadataObjectCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *MetadataObjectClient) ById(id string) (*MetadataObject, error) {
	resp := &MetadataObject{}
	err := c.rancherClient.doById(METADATA_OBJECT_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *MetadataObjectClient) Delete(container *MetadataObject) error {
	return c.rancherClient.doResourceDelete(METADATA_OBJECT_TYPE, &container.Resource)
}

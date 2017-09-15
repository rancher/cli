package client

const (
	SERVICE_ROLLBACK_TYPE = "serviceRollback"
)

type ServiceRollback struct {
	Resource

	RevisionId string `json:"revisionId,omitempty" yaml:"revision_id,omitempty"`
}

type ServiceRollbackCollection struct {
	Collection
	Data   []ServiceRollback `json:"data,omitempty"`
	client *ServiceRollbackClient
}

type ServiceRollbackClient struct {
	rancherClient *RancherClient
}

type ServiceRollbackOperations interface {
	List(opts *ListOpts) (*ServiceRollbackCollection, error)
	Create(opts *ServiceRollback) (*ServiceRollback, error)
	Update(existing *ServiceRollback, updates interface{}) (*ServiceRollback, error)
	ById(id string) (*ServiceRollback, error)
	Delete(container *ServiceRollback) error
}

func newServiceRollbackClient(rancherClient *RancherClient) *ServiceRollbackClient {
	return &ServiceRollbackClient{
		rancherClient: rancherClient,
	}
}

func (c *ServiceRollbackClient) Create(container *ServiceRollback) (*ServiceRollback, error) {
	resp := &ServiceRollback{}
	err := c.rancherClient.doCreate(SERVICE_ROLLBACK_TYPE, container, resp)
	return resp, err
}

func (c *ServiceRollbackClient) Update(existing *ServiceRollback, updates interface{}) (*ServiceRollback, error) {
	resp := &ServiceRollback{}
	err := c.rancherClient.doUpdate(SERVICE_ROLLBACK_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ServiceRollbackClient) List(opts *ListOpts) (*ServiceRollbackCollection, error) {
	resp := &ServiceRollbackCollection{}
	err := c.rancherClient.doList(SERVICE_ROLLBACK_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ServiceRollbackCollection) Next() (*ServiceRollbackCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ServiceRollbackCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ServiceRollbackClient) ById(id string) (*ServiceRollback, error) {
	resp := &ServiceRollback{}
	err := c.rancherClient.doById(SERVICE_ROLLBACK_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *ServiceRollbackClient) Delete(container *ServiceRollback) error {
	return c.rancherClient.doResourceDelete(SERVICE_ROLLBACK_TYPE, &container.Resource)
}

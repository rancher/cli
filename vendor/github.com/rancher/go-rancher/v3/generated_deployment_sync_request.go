package client

const (
	DEPLOYMENT_SYNC_REQUEST_TYPE = "deploymentSyncRequest"
)

type DeploymentSyncRequest struct {
	Resource

	ClusterId string `json:"clusterId,omitempty" yaml:"cluster_id,omitempty"`

	Containers []Container `json:"containers,omitempty" yaml:"containers,omitempty"`

	DeploymentUnitUuid string `json:"deploymentUnitUuid,omitempty" yaml:"deployment_unit_uuid,omitempty"`

	ExternalId string `json:"externalId,omitempty" yaml:"external_id,omitempty"`

	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`

	Networks []Network `json:"networks,omitempty" yaml:"networks,omitempty"`

	NodeName string `json:"nodeName,omitempty" yaml:"node_name,omitempty"`

	RegistryCredentials []Credential `json:"registryCredentials,omitempty" yaml:"registry_credentials,omitempty"`

	Revision string `json:"revision,omitempty" yaml:"revision,omitempty"`

	Volumes []Volume `json:"volumes,omitempty" yaml:"volumes,omitempty"`
}

type DeploymentSyncRequestCollection struct {
	Collection
	Data   []DeploymentSyncRequest `json:"data,omitempty"`
	client *DeploymentSyncRequestClient
}

type DeploymentSyncRequestClient struct {
	rancherClient *RancherClient
}

type DeploymentSyncRequestOperations interface {
	List(opts *ListOpts) (*DeploymentSyncRequestCollection, error)
	Create(opts *DeploymentSyncRequest) (*DeploymentSyncRequest, error)
	Update(existing *DeploymentSyncRequest, updates interface{}) (*DeploymentSyncRequest, error)
	ById(id string) (*DeploymentSyncRequest, error)
	Delete(container *DeploymentSyncRequest) error
}

func newDeploymentSyncRequestClient(rancherClient *RancherClient) *DeploymentSyncRequestClient {
	return &DeploymentSyncRequestClient{
		rancherClient: rancherClient,
	}
}

func (c *DeploymentSyncRequestClient) Create(container *DeploymentSyncRequest) (*DeploymentSyncRequest, error) {
	resp := &DeploymentSyncRequest{}
	err := c.rancherClient.doCreate(DEPLOYMENT_SYNC_REQUEST_TYPE, container, resp)
	return resp, err
}

func (c *DeploymentSyncRequestClient) Update(existing *DeploymentSyncRequest, updates interface{}) (*DeploymentSyncRequest, error) {
	resp := &DeploymentSyncRequest{}
	err := c.rancherClient.doUpdate(DEPLOYMENT_SYNC_REQUEST_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *DeploymentSyncRequestClient) List(opts *ListOpts) (*DeploymentSyncRequestCollection, error) {
	resp := &DeploymentSyncRequestCollection{}
	err := c.rancherClient.doList(DEPLOYMENT_SYNC_REQUEST_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *DeploymentSyncRequestCollection) Next() (*DeploymentSyncRequestCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &DeploymentSyncRequestCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *DeploymentSyncRequestClient) ById(id string) (*DeploymentSyncRequest, error) {
	resp := &DeploymentSyncRequest{}
	err := c.rancherClient.doById(DEPLOYMENT_SYNC_REQUEST_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *DeploymentSyncRequestClient) Delete(container *DeploymentSyncRequest) error {
	return c.rancherClient.doResourceDelete(DEPLOYMENT_SYNC_REQUEST_TYPE, &container.Resource)
}

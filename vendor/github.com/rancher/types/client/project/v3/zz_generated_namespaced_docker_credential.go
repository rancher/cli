package client

import (
	"github.com/rancher/norman/types"
)

const (
	NamespacedDockerCredentialType                 = "namespacedDockerCredential"
	NamespacedDockerCredentialFieldAnnotations     = "annotations"
	NamespacedDockerCredentialFieldCreated         = "created"
	NamespacedDockerCredentialFieldCreatorID       = "creatorId"
	NamespacedDockerCredentialFieldDescription     = "description"
	NamespacedDockerCredentialFieldLabels          = "labels"
	NamespacedDockerCredentialFieldName            = "name"
	NamespacedDockerCredentialFieldNamespaceId     = "namespaceId"
	NamespacedDockerCredentialFieldOwnerReferences = "ownerReferences"
	NamespacedDockerCredentialFieldProjectID       = "projectId"
	NamespacedDockerCredentialFieldRegistries      = "registries"
	NamespacedDockerCredentialFieldRemoved         = "removed"
	NamespacedDockerCredentialFieldUuid            = "uuid"
)

type NamespacedDockerCredential struct {
	types.Resource
	Annotations     map[string]string             `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	Created         string                        `json:"created,omitempty" yaml:"created,omitempty"`
	CreatorID       string                        `json:"creatorId,omitempty" yaml:"creatorId,omitempty"`
	Description     string                        `json:"description,omitempty" yaml:"description,omitempty"`
	Labels          map[string]string             `json:"labels,omitempty" yaml:"labels,omitempty"`
	Name            string                        `json:"name,omitempty" yaml:"name,omitempty"`
	NamespaceId     string                        `json:"namespaceId,omitempty" yaml:"namespaceId,omitempty"`
	OwnerReferences []OwnerReference              `json:"ownerReferences,omitempty" yaml:"ownerReferences,omitempty"`
	ProjectID       string                        `json:"projectId,omitempty" yaml:"projectId,omitempty"`
	Registries      map[string]RegistryCredential `json:"registries,omitempty" yaml:"registries,omitempty"`
	Removed         string                        `json:"removed,omitempty" yaml:"removed,omitempty"`
	Uuid            string                        `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}
type NamespacedDockerCredentialCollection struct {
	types.Collection
	Data   []NamespacedDockerCredential `json:"data,omitempty"`
	client *NamespacedDockerCredentialClient
}

type NamespacedDockerCredentialClient struct {
	apiClient *Client
}

type NamespacedDockerCredentialOperations interface {
	List(opts *types.ListOpts) (*NamespacedDockerCredentialCollection, error)
	Create(opts *NamespacedDockerCredential) (*NamespacedDockerCredential, error)
	Update(existing *NamespacedDockerCredential, updates interface{}) (*NamespacedDockerCredential, error)
	ByID(id string) (*NamespacedDockerCredential, error)
	Delete(container *NamespacedDockerCredential) error
}

func newNamespacedDockerCredentialClient(apiClient *Client) *NamespacedDockerCredentialClient {
	return &NamespacedDockerCredentialClient{
		apiClient: apiClient,
	}
}

func (c *NamespacedDockerCredentialClient) Create(container *NamespacedDockerCredential) (*NamespacedDockerCredential, error) {
	resp := &NamespacedDockerCredential{}
	err := c.apiClient.Ops.DoCreate(NamespacedDockerCredentialType, container, resp)
	return resp, err
}

func (c *NamespacedDockerCredentialClient) Update(existing *NamespacedDockerCredential, updates interface{}) (*NamespacedDockerCredential, error) {
	resp := &NamespacedDockerCredential{}
	err := c.apiClient.Ops.DoUpdate(NamespacedDockerCredentialType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *NamespacedDockerCredentialClient) List(opts *types.ListOpts) (*NamespacedDockerCredentialCollection, error) {
	resp := &NamespacedDockerCredentialCollection{}
	err := c.apiClient.Ops.DoList(NamespacedDockerCredentialType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *NamespacedDockerCredentialCollection) Next() (*NamespacedDockerCredentialCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &NamespacedDockerCredentialCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *NamespacedDockerCredentialClient) ByID(id string) (*NamespacedDockerCredential, error) {
	resp := &NamespacedDockerCredential{}
	err := c.apiClient.Ops.DoByID(NamespacedDockerCredentialType, id, resp)
	return resp, err
}

func (c *NamespacedDockerCredentialClient) Delete(container *NamespacedDockerCredential) error {
	return c.apiClient.Ops.DoResourceDelete(NamespacedDockerCredentialType, &container.Resource)
}

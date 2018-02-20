package client

import (
	"github.com/rancher/norman/types"
)

const (
	StorageClassType                      = "storageClass"
	StorageClassFieldAllowVolumeExpansion = "allowVolumeExpansion"
	StorageClassFieldAnnotations          = "annotations"
	StorageClassFieldCreated              = "created"
	StorageClassFieldCreatorID            = "creatorId"
	StorageClassFieldLabels               = "labels"
	StorageClassFieldMountOptions         = "mountOptions"
	StorageClassFieldName                 = "name"
	StorageClassFieldOwnerReferences      = "ownerReferences"
	StorageClassFieldParameters           = "parameters"
	StorageClassFieldProvisioner          = "provisioner"
	StorageClassFieldReclaimPolicy        = "reclaimPolicy"
	StorageClassFieldRemoved              = "removed"
	StorageClassFieldUuid                 = "uuid"
)

type StorageClass struct {
	types.Resource
	AllowVolumeExpansion *bool             `json:"allowVolumeExpansion,omitempty"`
	Annotations          map[string]string `json:"annotations,omitempty"`
	Created              string            `json:"created,omitempty"`
	CreatorID            string            `json:"creatorId,omitempty"`
	Labels               map[string]string `json:"labels,omitempty"`
	MountOptions         []string          `json:"mountOptions,omitempty"`
	Name                 string            `json:"name,omitempty"`
	OwnerReferences      []OwnerReference  `json:"ownerReferences,omitempty"`
	Parameters           map[string]string `json:"parameters,omitempty"`
	Provisioner          string            `json:"provisioner,omitempty"`
	ReclaimPolicy        string            `json:"reclaimPolicy,omitempty"`
	Removed              string            `json:"removed,omitempty"`
	Uuid                 string            `json:"uuid,omitempty"`
}
type StorageClassCollection struct {
	types.Collection
	Data   []StorageClass `json:"data,omitempty"`
	client *StorageClassClient
}

type StorageClassClient struct {
	apiClient *Client
}

type StorageClassOperations interface {
	List(opts *types.ListOpts) (*StorageClassCollection, error)
	Create(opts *StorageClass) (*StorageClass, error)
	Update(existing *StorageClass, updates interface{}) (*StorageClass, error)
	ByID(id string) (*StorageClass, error)
	Delete(container *StorageClass) error
}

func newStorageClassClient(apiClient *Client) *StorageClassClient {
	return &StorageClassClient{
		apiClient: apiClient,
	}
}

func (c *StorageClassClient) Create(container *StorageClass) (*StorageClass, error) {
	resp := &StorageClass{}
	err := c.apiClient.Ops.DoCreate(StorageClassType, container, resp)
	return resp, err
}

func (c *StorageClassClient) Update(existing *StorageClass, updates interface{}) (*StorageClass, error) {
	resp := &StorageClass{}
	err := c.apiClient.Ops.DoUpdate(StorageClassType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *StorageClassClient) List(opts *types.ListOpts) (*StorageClassCollection, error) {
	resp := &StorageClassCollection{}
	err := c.apiClient.Ops.DoList(StorageClassType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *StorageClassCollection) Next() (*StorageClassCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &StorageClassCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *StorageClassClient) ByID(id string) (*StorageClass, error) {
	resp := &StorageClass{}
	err := c.apiClient.Ops.DoByID(StorageClassType, id, resp)
	return resp, err
}

func (c *StorageClassClient) Delete(container *StorageClass) error {
	return c.apiClient.Ops.DoResourceDelete(StorageClassType, &container.Resource)
}

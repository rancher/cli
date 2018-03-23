package client

import (
	"github.com/rancher/norman/types"
)

const (
	TemplateVersionType                       = "templateVersion"
	TemplateVersionFieldAnnotations           = "annotations"
	TemplateVersionFieldCreated               = "created"
	TemplateVersionFieldCreatorID             = "creatorId"
	TemplateVersionFieldExternalID            = "externalId"
	TemplateVersionFieldFiles                 = "files"
	TemplateVersionFieldLabels                = "labels"
	TemplateVersionFieldMaximumRancherVersion = "maximumRancherVersion"
	TemplateVersionFieldMinimumRancherVersion = "minimumRancherVersion"
	TemplateVersionFieldName                  = "name"
	TemplateVersionFieldOwnerReferences       = "ownerReferences"
	TemplateVersionFieldQuestions             = "questions"
	TemplateVersionFieldReadme                = "readme"
	TemplateVersionFieldRemoved               = "removed"
	TemplateVersionFieldRevision              = "revision"
	TemplateVersionFieldState                 = "state"
	TemplateVersionFieldStatus                = "status"
	TemplateVersionFieldTransitioning         = "transitioning"
	TemplateVersionFieldTransitioningMessage  = "transitioningMessage"
	TemplateVersionFieldUpgradeFrom           = "upgradeFrom"
	TemplateVersionFieldUpgradeVersionLinks   = "upgradeVersionLinks"
	TemplateVersionFieldUuid                  = "uuid"
	TemplateVersionFieldVersion               = "version"
)

type TemplateVersion struct {
	types.Resource
	Annotations           map[string]string      `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	Created               string                 `json:"created,omitempty" yaml:"created,omitempty"`
	CreatorID             string                 `json:"creatorId,omitempty" yaml:"creatorId,omitempty"`
	ExternalID            string                 `json:"externalId,omitempty" yaml:"externalId,omitempty"`
	Files                 []File                 `json:"files,omitempty" yaml:"files,omitempty"`
	Labels                map[string]string      `json:"labels,omitempty" yaml:"labels,omitempty"`
	MaximumRancherVersion string                 `json:"maximumRancherVersion,omitempty" yaml:"maximumRancherVersion,omitempty"`
	MinimumRancherVersion string                 `json:"minimumRancherVersion,omitempty" yaml:"minimumRancherVersion,omitempty"`
	Name                  string                 `json:"name,omitempty" yaml:"name,omitempty"`
	OwnerReferences       []OwnerReference       `json:"ownerReferences,omitempty" yaml:"ownerReferences,omitempty"`
	Questions             []Question             `json:"questions,omitempty" yaml:"questions,omitempty"`
	Readme                string                 `json:"readme,omitempty" yaml:"readme,omitempty"`
	Removed               string                 `json:"removed,omitempty" yaml:"removed,omitempty"`
	Revision              *int64                 `json:"revision,omitempty" yaml:"revision,omitempty"`
	State                 string                 `json:"state,omitempty" yaml:"state,omitempty"`
	Status                *TemplateVersionStatus `json:"status,omitempty" yaml:"status,omitempty"`
	Transitioning         string                 `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`
	TransitioningMessage  string                 `json:"transitioningMessage,omitempty" yaml:"transitioningMessage,omitempty"`
	UpgradeFrom           string                 `json:"upgradeFrom,omitempty" yaml:"upgradeFrom,omitempty"`
	UpgradeVersionLinks   map[string]string      `json:"upgradeVersionLinks,omitempty" yaml:"upgradeVersionLinks,omitempty"`
	Uuid                  string                 `json:"uuid,omitempty" yaml:"uuid,omitempty"`
	Version               string                 `json:"version,omitempty" yaml:"version,omitempty"`
}
type TemplateVersionCollection struct {
	types.Collection
	Data   []TemplateVersion `json:"data,omitempty"`
	client *TemplateVersionClient
}

type TemplateVersionClient struct {
	apiClient *Client
}

type TemplateVersionOperations interface {
	List(opts *types.ListOpts) (*TemplateVersionCollection, error)
	Create(opts *TemplateVersion) (*TemplateVersion, error)
	Update(existing *TemplateVersion, updates interface{}) (*TemplateVersion, error)
	ByID(id string) (*TemplateVersion, error)
	Delete(container *TemplateVersion) error
}

func newTemplateVersionClient(apiClient *Client) *TemplateVersionClient {
	return &TemplateVersionClient{
		apiClient: apiClient,
	}
}

func (c *TemplateVersionClient) Create(container *TemplateVersion) (*TemplateVersion, error) {
	resp := &TemplateVersion{}
	err := c.apiClient.Ops.DoCreate(TemplateVersionType, container, resp)
	return resp, err
}

func (c *TemplateVersionClient) Update(existing *TemplateVersion, updates interface{}) (*TemplateVersion, error) {
	resp := &TemplateVersion{}
	err := c.apiClient.Ops.DoUpdate(TemplateVersionType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *TemplateVersionClient) List(opts *types.ListOpts) (*TemplateVersionCollection, error) {
	resp := &TemplateVersionCollection{}
	err := c.apiClient.Ops.DoList(TemplateVersionType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *TemplateVersionCollection) Next() (*TemplateVersionCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &TemplateVersionCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *TemplateVersionClient) ByID(id string) (*TemplateVersion, error) {
	resp := &TemplateVersion{}
	err := c.apiClient.Ops.DoByID(TemplateVersionType, id, resp)
	return resp, err
}

func (c *TemplateVersionClient) Delete(container *TemplateVersion) error {
	return c.apiClient.Ops.DoResourceDelete(TemplateVersionType, &container.Resource)
}

package client

import (
	"github.com/rancher/norman/types"
)

const (
	PodSecurityPolicyTemplateType                                 = "podSecurityPolicyTemplate"
	PodSecurityPolicyTemplateFieldAllowPrivilegeEscalation        = "allowPrivilegeEscalation"
	PodSecurityPolicyTemplateFieldAllowedCapabilities             = "allowedCapabilities"
	PodSecurityPolicyTemplateFieldAllowedHostPaths                = "allowedHostPaths"
	PodSecurityPolicyTemplateFieldAnnotations                     = "annotations"
	PodSecurityPolicyTemplateFieldCreated                         = "created"
	PodSecurityPolicyTemplateFieldCreatorID                       = "creatorId"
	PodSecurityPolicyTemplateFieldDefaultAddCapabilities          = "defaultAddCapabilities"
	PodSecurityPolicyTemplateFieldDefaultAllowPrivilegeEscalation = "defaultAllowPrivilegeEscalation"
	PodSecurityPolicyTemplateFieldDescription                     = "description"
	PodSecurityPolicyTemplateFieldFSGroup                         = "fsGroup"
	PodSecurityPolicyTemplateFieldHostIPC                         = "hostIPC"
	PodSecurityPolicyTemplateFieldHostNetwork                     = "hostNetwork"
	PodSecurityPolicyTemplateFieldHostPID                         = "hostPID"
	PodSecurityPolicyTemplateFieldHostPorts                       = "hostPorts"
	PodSecurityPolicyTemplateFieldLabels                          = "labels"
	PodSecurityPolicyTemplateFieldName                            = "name"
	PodSecurityPolicyTemplateFieldOwnerReferences                 = "ownerReferences"
	PodSecurityPolicyTemplateFieldPrivileged                      = "privileged"
	PodSecurityPolicyTemplateFieldReadOnlyRootFilesystem          = "readOnlyRootFilesystem"
	PodSecurityPolicyTemplateFieldRemoved                         = "removed"
	PodSecurityPolicyTemplateFieldRequiredDropCapabilities        = "requiredDropCapabilities"
	PodSecurityPolicyTemplateFieldRunAsUser                       = "runAsUser"
	PodSecurityPolicyTemplateFieldSELinux                         = "seLinux"
	PodSecurityPolicyTemplateFieldSupplementalGroups              = "supplementalGroups"
	PodSecurityPolicyTemplateFieldUuid                            = "uuid"
	PodSecurityPolicyTemplateFieldVolumes                         = "volumes"
)

type PodSecurityPolicyTemplate struct {
	types.Resource
	AllowPrivilegeEscalation        *bool                              `json:"allowPrivilegeEscalation,omitempty"`
	AllowedCapabilities             []string                           `json:"allowedCapabilities,omitempty"`
	AllowedHostPaths                []AllowedHostPath                  `json:"allowedHostPaths,omitempty"`
	Annotations                     map[string]string                  `json:"annotations,omitempty"`
	Created                         string                             `json:"created,omitempty"`
	CreatorID                       string                             `json:"creatorId,omitempty"`
	DefaultAddCapabilities          []string                           `json:"defaultAddCapabilities,omitempty"`
	DefaultAllowPrivilegeEscalation *bool                              `json:"defaultAllowPrivilegeEscalation,omitempty"`
	Description                     string                             `json:"description,omitempty"`
	FSGroup                         *FSGroupStrategyOptions            `json:"fsGroup,omitempty"`
	HostIPC                         bool                               `json:"hostIPC,omitempty"`
	HostNetwork                     bool                               `json:"hostNetwork,omitempty"`
	HostPID                         bool                               `json:"hostPID,omitempty"`
	HostPorts                       []HostPortRange                    `json:"hostPorts,omitempty"`
	Labels                          map[string]string                  `json:"labels,omitempty"`
	Name                            string                             `json:"name,omitempty"`
	OwnerReferences                 []OwnerReference                   `json:"ownerReferences,omitempty"`
	Privileged                      bool                               `json:"privileged,omitempty"`
	ReadOnlyRootFilesystem          bool                               `json:"readOnlyRootFilesystem,omitempty"`
	Removed                         string                             `json:"removed,omitempty"`
	RequiredDropCapabilities        []string                           `json:"requiredDropCapabilities,omitempty"`
	RunAsUser                       *RunAsUserStrategyOptions          `json:"runAsUser,omitempty"`
	SELinux                         *SELinuxStrategyOptions            `json:"seLinux,omitempty"`
	SupplementalGroups              *SupplementalGroupsStrategyOptions `json:"supplementalGroups,omitempty"`
	Uuid                            string                             `json:"uuid,omitempty"`
	Volumes                         []string                           `json:"volumes,omitempty"`
}
type PodSecurityPolicyTemplateCollection struct {
	types.Collection
	Data   []PodSecurityPolicyTemplate `json:"data,omitempty"`
	client *PodSecurityPolicyTemplateClient
}

type PodSecurityPolicyTemplateClient struct {
	apiClient *Client
}

type PodSecurityPolicyTemplateOperations interface {
	List(opts *types.ListOpts) (*PodSecurityPolicyTemplateCollection, error)
	Create(opts *PodSecurityPolicyTemplate) (*PodSecurityPolicyTemplate, error)
	Update(existing *PodSecurityPolicyTemplate, updates interface{}) (*PodSecurityPolicyTemplate, error)
	ByID(id string) (*PodSecurityPolicyTemplate, error)
	Delete(container *PodSecurityPolicyTemplate) error
}

func newPodSecurityPolicyTemplateClient(apiClient *Client) *PodSecurityPolicyTemplateClient {
	return &PodSecurityPolicyTemplateClient{
		apiClient: apiClient,
	}
}

func (c *PodSecurityPolicyTemplateClient) Create(container *PodSecurityPolicyTemplate) (*PodSecurityPolicyTemplate, error) {
	resp := &PodSecurityPolicyTemplate{}
	err := c.apiClient.Ops.DoCreate(PodSecurityPolicyTemplateType, container, resp)
	return resp, err
}

func (c *PodSecurityPolicyTemplateClient) Update(existing *PodSecurityPolicyTemplate, updates interface{}) (*PodSecurityPolicyTemplate, error) {
	resp := &PodSecurityPolicyTemplate{}
	err := c.apiClient.Ops.DoUpdate(PodSecurityPolicyTemplateType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *PodSecurityPolicyTemplateClient) List(opts *types.ListOpts) (*PodSecurityPolicyTemplateCollection, error) {
	resp := &PodSecurityPolicyTemplateCollection{}
	err := c.apiClient.Ops.DoList(PodSecurityPolicyTemplateType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *PodSecurityPolicyTemplateCollection) Next() (*PodSecurityPolicyTemplateCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &PodSecurityPolicyTemplateCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *PodSecurityPolicyTemplateClient) ByID(id string) (*PodSecurityPolicyTemplate, error) {
	resp := &PodSecurityPolicyTemplate{}
	err := c.apiClient.Ops.DoByID(PodSecurityPolicyTemplateType, id, resp)
	return resp, err
}

func (c *PodSecurityPolicyTemplateClient) Delete(container *PodSecurityPolicyTemplate) error {
	return c.apiClient.Ops.DoResourceDelete(PodSecurityPolicyTemplateType, &container.Resource)
}

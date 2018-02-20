package client

import (
	"github.com/rancher/norman/types"
)

const (
	UserType                    = "user"
	UserFieldAnnotations        = "annotations"
	UserFieldCreated            = "created"
	UserFieldCreatorID          = "creatorId"
	UserFieldDescription        = "description"
	UserFieldLabels             = "labels"
	UserFieldMe                 = "me"
	UserFieldMustChangePassword = "mustChangePassword"
	UserFieldName               = "name"
	UserFieldOwnerReferences    = "ownerReferences"
	UserFieldPassword           = "password"
	UserFieldPrincipalIDs       = "principalIds"
	UserFieldRemoved            = "removed"
	UserFieldUsername           = "username"
	UserFieldUuid               = "uuid"
)

type User struct {
	types.Resource
	Annotations        map[string]string `json:"annotations,omitempty"`
	Created            string            `json:"created,omitempty"`
	CreatorID          string            `json:"creatorId,omitempty"`
	Description        string            `json:"description,omitempty"`
	Labels             map[string]string `json:"labels,omitempty"`
	Me                 bool              `json:"me,omitempty"`
	MustChangePassword bool              `json:"mustChangePassword,omitempty"`
	Name               string            `json:"name,omitempty"`
	OwnerReferences    []OwnerReference  `json:"ownerReferences,omitempty"`
	Password           string            `json:"password,omitempty"`
	PrincipalIDs       []string          `json:"principalIds,omitempty"`
	Removed            string            `json:"removed,omitempty"`
	Username           string            `json:"username,omitempty"`
	Uuid               string            `json:"uuid,omitempty"`
}
type UserCollection struct {
	types.Collection
	Data   []User `json:"data,omitempty"`
	client *UserClient
}

type UserClient struct {
	apiClient *Client
}

type UserOperations interface {
	List(opts *types.ListOpts) (*UserCollection, error)
	Create(opts *User) (*User, error)
	Update(existing *User, updates interface{}) (*User, error)
	ByID(id string) (*User, error)
	Delete(container *User) error

	ActionSetpassword(*User, *SetPasswordInput) (*User, error)
}

func newUserClient(apiClient *Client) *UserClient {
	return &UserClient{
		apiClient: apiClient,
	}
}

func (c *UserClient) Create(container *User) (*User, error) {
	resp := &User{}
	err := c.apiClient.Ops.DoCreate(UserType, container, resp)
	return resp, err
}

func (c *UserClient) Update(existing *User, updates interface{}) (*User, error) {
	resp := &User{}
	err := c.apiClient.Ops.DoUpdate(UserType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *UserClient) List(opts *types.ListOpts) (*UserCollection, error) {
	resp := &UserCollection{}
	err := c.apiClient.Ops.DoList(UserType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *UserCollection) Next() (*UserCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &UserCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *UserClient) ByID(id string) (*User, error) {
	resp := &User{}
	err := c.apiClient.Ops.DoByID(UserType, id, resp)
	return resp, err
}

func (c *UserClient) Delete(container *User) error {
	return c.apiClient.Ops.DoResourceDelete(UserType, &container.Resource)
}

func (c *UserClient) ActionSetpassword(resource *User, input *SetPasswordInput) (*User, error) {

	resp := &User{}

	err := c.apiClient.Ops.DoAction(UserType, "setpassword", &resource.Resource, input, resp)

	return resp, err
}

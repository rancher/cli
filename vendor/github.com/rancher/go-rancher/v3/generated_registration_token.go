package client

const (
	REGISTRATION_TOKEN_TYPE = "registrationToken"
)

type RegistrationToken struct {
	Resource

	ClusterCommand string `json:"clusterCommand,omitempty" yaml:"cluster_command,omitempty"`

	HostCommand string `json:"hostCommand,omitempty" yaml:"host_command,omitempty"`

	Image string `json:"image,omitempty" yaml:"image,omitempty"`

	RegistrationUrl string `json:"registrationUrl,omitempty" yaml:"registration_url,omitempty"`

	Token string `json:"token,omitempty" yaml:"token,omitempty"`

	WindowsCommand string `json:"windowsCommand,omitempty" yaml:"windows_command,omitempty"`
}

type RegistrationTokenCollection struct {
	Collection
	Data   []RegistrationToken `json:"data,omitempty"`
	client *RegistrationTokenClient
}

type RegistrationTokenClient struct {
	rancherClient *RancherClient
}

type RegistrationTokenOperations interface {
	List(opts *ListOpts) (*RegistrationTokenCollection, error)
	Create(opts *RegistrationToken) (*RegistrationToken, error)
	Update(existing *RegistrationToken, updates interface{}) (*RegistrationToken, error)
	ById(id string) (*RegistrationToken, error)
	Delete(container *RegistrationToken) error
}

func newRegistrationTokenClient(rancherClient *RancherClient) *RegistrationTokenClient {
	return &RegistrationTokenClient{
		rancherClient: rancherClient,
	}
}

func (c *RegistrationTokenClient) Create(container *RegistrationToken) (*RegistrationToken, error) {
	resp := &RegistrationToken{}
	err := c.rancherClient.doCreate(REGISTRATION_TOKEN_TYPE, container, resp)
	return resp, err
}

func (c *RegistrationTokenClient) Update(existing *RegistrationToken, updates interface{}) (*RegistrationToken, error) {
	resp := &RegistrationToken{}
	err := c.rancherClient.doUpdate(REGISTRATION_TOKEN_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *RegistrationTokenClient) List(opts *ListOpts) (*RegistrationTokenCollection, error) {
	resp := &RegistrationTokenCollection{}
	err := c.rancherClient.doList(REGISTRATION_TOKEN_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *RegistrationTokenCollection) Next() (*RegistrationTokenCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &RegistrationTokenCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *RegistrationTokenClient) ById(id string) (*RegistrationToken, error) {
	resp := &RegistrationToken{}
	err := c.rancherClient.doById(REGISTRATION_TOKEN_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *RegistrationTokenClient) Delete(container *RegistrationToken) error {
	return c.rancherClient.doResourceDelete(REGISTRATION_TOKEN_TYPE, &container.Resource)
}

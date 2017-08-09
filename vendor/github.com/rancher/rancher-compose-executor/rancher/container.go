package rancher

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/rancher/go-rancher/v2"
	"github.com/rancher/rancher-compose-executor/config"
	"github.com/rancher/rancher-compose-executor/project"
	"github.com/rancher/rancher-compose-executor/project/options"
)

type RancherContainer struct {
	name          string
	serviceConfig *config.ServiceConfig
	context       *Context
}

func (r *RancherContainer) ID() string {
	return ""
}

func (r *RancherContainer) Name() string {
	return r.name
}

func (r *RancherContainer) Config() *config.ServiceConfig {
	return r.serviceConfig
}

func (r *RancherContainer) Context() *Context {
	return r.context
}

func NewContainer(name string, config *config.ServiceConfig, context *Context) *RancherContainer {
	return &RancherContainer{
		name:          name,
		serviceConfig: config,
		context:       context,
	}
}

func (r *RancherContainer) Create(ctx context.Context, options options.Create) error {
	fmt.Println(r.Name(), "Create")
	return nil
}

func (r *RancherContainer) Up(ctx context.Context, options options.Up) error {
	fmt.Println(r.Name(), "Up")
	return nil
}

func (r *RancherContainer) Build(ctx context.Context, buildOptions options.Build) error {
	fmt.Println(r.Name(), "Build")
	return nil
}

func (r *RancherContainer) Log(ctx context.Context, follow bool) error {
	fmt.Println(r.Name(), "Log")
	return nil
}

func (r *RancherContainer) DependentServices() []project.ServiceRelationship {
	return []project.ServiceRelationship{}
}

func (r *RancherContainer) Client() *client.RancherClient {
	return r.context.Client
}

func (r *RancherContainer) Pull(ctx context.Context) error {
	fmt.Println(r.Name(), "Pull")
	return nil
}

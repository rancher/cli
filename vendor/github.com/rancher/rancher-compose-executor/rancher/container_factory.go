package rancher

import (
	"github.com/rancher/rancher-compose-executor/config"
	"github.com/rancher/rancher-compose-executor/project"
)

type RancherContainerFactory struct {
	Context *Context
}

func (r *RancherContainerFactory) Create(project *project.Project, name string, serviceConfig *config.ServiceConfig) (project.Service, error) {
	return NewContainer(name, serviceConfig, r.Context), nil
}

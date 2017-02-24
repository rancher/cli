package rancher

import (
	"github.com/rancher/rancher-compose-executor/config"
	"github.com/rancher/rancher-compose-executor/project"
)

type RancherServiceFactory struct {
	Context *Context
}

func (r *RancherServiceFactory) Create(project *project.Project, name string, serviceConfig *config.ServiceConfig) (project.Service, error) {
	if len(r.Context.SidekickInfo.sidekickToPrimaries[name]) > 0 {
		return NewSidekick(name, serviceConfig, r.Context), nil
	} else {
		return NewService(name, serviceConfig, r.Context), nil
	}
}

package rancher

import (
	"golang.org/x/net/context"

	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
)

type Sidekick struct {
	project.EmptyService

	name          string
	serviceConfig *config.ServiceConfig
	context       *Context
}

func NewSidekick(name string, serviceConfig *config.ServiceConfig, context *Context) *Sidekick {
	return &Sidekick{
		name:          name,
		serviceConfig: serviceConfig,
		context:       context,
	}
}

func (s *Sidekick) Name() string {
	return s.name
}

func (s *Sidekick) primaries() []string {
	return s.context.SidekickInfo.sidekickToPrimaries[s.name]
}

func (s *Sidekick) Config() *config.ServiceConfig {
	links := []string{}

	for _, primary := range s.primaries() {
		links = append(links, primary)
	}

	config := *s.serviceConfig
	config.Links = links
	config.VolumesFrom = []string{}

	return &config
}

func (s *Sidekick) DependentServices() []project.ServiceRelationship {
	dependentServices := project.DefaultDependentServices(s.context.Project, s)
	for i, dependentService := range dependentServices {
		if dependentService.Type == project.RelTypeLink {
			dependentServices[i].Optional = true
		}
	}

	return dependentServices
}

func (s *Sidekick) Log(ctx context.Context, follow bool) error {
	return nil
}

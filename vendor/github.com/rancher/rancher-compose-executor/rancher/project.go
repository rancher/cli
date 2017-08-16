package rancher

import (
	"github.com/Sirupsen/logrus"
	"github.com/rancher/rancher-compose-executor/project"
)

func NewProject(context *Context) (*project.Project, error) {
	context.ServiceFactory = &RancherServiceFactory{
		Context: context,
	}

	context.ContainerFactory = &RancherContainerFactory{
		Context: context,
	}

	context.DependenciesFactory = &RancherDependenciesFactory{
		Context: context,
	}

	context.VolumesFactory = &RancherVolumesFactory{
		Context: context,
	}

	context.HostsFactory = &RancherHostsFactory{
		Context: context,
	}

	context.SecretsFactory = &RancherSecretsFactory{
		Context: context,
	}

	p := project.NewProject(&context.Context)
	err := p.Open()
	if err != nil {
		return nil, err
	}

	if err := context.open(); err != nil {
		logrus.Errorf("Failed to open project %s: %v", p.Name, err)
		return nil, err
	}

	if err := p.Parse(); err != nil {
		return nil, err
	}

	context.SidekickInfo = NewSidekickInfo(p)

	return p, nil
}

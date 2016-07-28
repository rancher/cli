package rancher

import (
	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	"github.com/rancher/rancher-compose/preprocess"
)

func NewProject(context *Context) (*project.Project, error) {
	context.ServiceFactory = &RancherServiceFactory{
		Context: context,
	}

	p := project.NewProject(&context.Context, nil, &config.ParseOptions{
		Interpolate: true,
		Validate:    true,
		Preprocess:  preprocess.PreprocessServiceMap,
	})

	err := p.Parse()
	if err != nil {
		return nil, err
	}

	if err = context.open(); err != nil {
		logrus.Errorf("Failed to open project %s: %v", p.Name, err)
		return nil, err
	}

	p.Name = context.ProjectName

	context.SidekickInfo = NewSidekickInfo(p)

	return p, err
}

package rancher

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	"github.com/rancher/rancher-compose/preprocess"
	"io/ioutil"
)

func NewProject(context *Context) (*project.Project, error) {
	context.ServiceFactory = &RancherServiceFactory{
		Context: context,
	}

	context.VolumesFactory = &RancherVolumesFactory{
		Context: context,
	}

	if context.Binding != nil {
		bindingBytes, err := json.Marshal(context.Binding)
		if err != nil {
			return nil, err
		}
		context.BindingsBytes = bindingBytes
	}

	if context.BindingsBytes == nil {
		if context.BindingsFile != "" {
			bindingsContent, err := ioutil.ReadFile(context.BindingsFile)
			if err != nil {
				return nil, err
			}
			context.BindingsBytes = bindingsContent
		}
	}

	preProcessServiceMap := preprocess.PreprocessServiceMap(context.BindingsBytes)
	p := project.NewProject(&context.Context, nil, &config.ParseOptions{
		Interpolate: true,
		Validate:    true,
		Preprocess:  preProcessServiceMap,
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

package rancher

import (
	"github.com/Sirupsen/logrus"
	"github.com/rancher/rancher-compose-executor/project"
)

func pruneServices(context *Context, project *project.Project) error {
	if context.Stack == nil {
		return nil
	}

	for _, serviceId := range context.Stack.ServiceIds {
		service, err := context.Client.Service.ById(serviceId)
		if err != nil {
			return err
		}
		if !project.ServiceConfigs.Has(service.Name) {
			err = context.Client.Service.Delete(service)
			if err != nil {
				return err
			}
			logrus.Infof("Deleting service %s", service.Name)
		}
	}

	return nil
}

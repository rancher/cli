package rancher

import (
	"time"

	rancherClient "github.com/rancher/go-rancher/client"
)

func (r *RancherService) WaitFor(resource *rancherClient.Resource, output interface{}, transitioning func() string) error {
	for {
		if transitioning() != "yes" {
			return nil
		}

		time.Sleep(150 * time.Millisecond)

		err := r.context.Client.Reload(resource, output)
		if err != nil {
			return err
		}
	}
}

func (r *RancherService) Wait(service *rancherClient.Service) error {
	return r.WaitFor(&service.Resource, service, func() string {
		return service.Transitioning
	})
}

func (r *RancherService) waitInstance(instance *rancherClient.Instance) error {
	return r.WaitFor(&instance.Resource, instance, func() string {
		return instance.Transitioning
	})
}

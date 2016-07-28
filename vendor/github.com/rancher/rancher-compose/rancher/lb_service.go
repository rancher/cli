package rancher

import rancherClient "github.com/rancher/go-rancher/client"

func populateLbFields(r *RancherService, launchConfig *rancherClient.LaunchConfig, service *CompositeService) error {
	config, ok := r.context.RancherConfig[r.name]
	if ok {
		service.LoadBalancerConfig = config.LoadBalancerConfig
	}

	if err := populateCerts(r.context.Client, service, &config); err != nil {
		return err
	}

	if FindServiceType(r) == LbServiceType {
		launchConfig.ImageUuid = ""
		// Write back to the ports passed in because the Docker parsing logic changes then
		launchConfig.Ports = r.serviceConfig.Ports
		launchConfig.Expose = r.serviceConfig.Expose
	}

	return nil
}

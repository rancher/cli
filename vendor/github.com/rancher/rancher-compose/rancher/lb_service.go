package rancher

import "github.com/rancher/go-rancher/v2"

const (
	defaultLoadBalancerImage = ""
)

func populateLbFields(r *RancherService, launchConfig *client.LaunchConfig, service *CompositeService) error {
	config, ok := r.context.RancherConfig[r.name]
	if !ok {
		return nil
	}

	if config.LbConfig == nil {
		return nil
	}

	// TODO: certs, defaultcert ?
	service.LbConfig = &client.LbConfig{
		Config: config.LbConfig.Config,
		StickinessPolicy: &client.LoadBalancerCookieStickinessPolicy{
			Name:     config.LbConfig.StickinessPolicy.Name,
			Cookie:   config.LbConfig.StickinessPolicy.Cookie,
			Domain:   config.LbConfig.StickinessPolicy.Domain,
			Indirect: config.LbConfig.StickinessPolicy.Indirect,
			Nocache:  config.LbConfig.StickinessPolicy.Nocache,
			Postonly: config.LbConfig.StickinessPolicy.Postonly,
			Mode:     config.LbConfig.StickinessPolicy.Mode,
		},
	}
	for _, portRule := range config.LbConfig.PortRules {
		targetService, err := r.FindExisting(portRule.Service)
		if err != nil {
			return err
		}
		service.LbConfig.PortRules = append(service.LbConfig.PortRules, client.PortRule{
			SourcePort:  int64(portRule.SourcePort),
			Protocol:    portRule.Protocol,
			Path:        portRule.Path,
			Hostname:    portRule.Hostname,
			ServiceId:   targetService.Id,
			TargetPort:  int64(portRule.TargetPort),
			Priority:    int64(portRule.Priority),
			BackendName: portRule.BackendName,
			Selector:    portRule.Selector,
		})
	}

	var defaultCert string
	var certs []string

	serviceType := FindServiceType(r)
	if serviceType == LegacyLbServiceType {
		launchConfig.ImageUuid = ""
		// Write back to the ports passed in because the Docker parsing logic changes then
		launchConfig.Ports = r.serviceConfig.Ports
		launchConfig.Expose = r.serviceConfig.Expose

		defaultCert = config.DefaultCert
		certs = config.Certs
	} else if serviceType == LbServiceType {
		// TODO: need this for v2?
		launchConfig.Ports = r.serviceConfig.Ports
		launchConfig.Expose = r.serviceConfig.Expose

		defaultCert = config.LbConfig.DefaultCert
		certs = config.LbConfig.Certs
	}

	if err := populateCerts(r.context.Client, service, defaultCert, certs); err != nil {
		return err
	}

	return nil
}

func convert(ports, links, externalLinks []string) ([]PortRule, error) {
	portRules := []PortRule{}
	for _, port := range ports {
		_ = port
		for _, link := range links {
			_ = link
			portRules = append(portRules, PortRule{
				SourcePort: 0,
				TargetPort: 0,
				Service:    "",
			})
		}
		for _, externalLink := range externalLinks {
			_ = externalLink
			portRules = append(portRules, PortRule{
				SourcePort: 0,
				TargetPort: 0,
				Service:    "",
			})
		}
	}
	return portRules, nil
}

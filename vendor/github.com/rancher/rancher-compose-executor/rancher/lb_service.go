package rancher

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	legacyClient "github.com/rancher/go-rancher/client"
	"github.com/rancher/go-rancher/v2"
)

func populateLbFields(r *RancherService, launchConfig *client.LaunchConfig, service *CompositeService) error {
	serviceType := FindServiceType(r)

	config, ok := r.context.RancherConfig[r.name]
	if ok {
		if serviceType == RancherType && config.LbConfig != nil {
			service.LbConfig = &client.LbTargetConfig{}
			service.LbConfig.PortRules = []client.TargetPortRule{}
			for _, portRule := range config.LbConfig.PortRules {
				service.LbConfig.PortRules = append(service.LbConfig.PortRules, client.TargetPortRule{
					BackendName: portRule.BackendName,
					Hostname:    portRule.Hostname,
					Path:        portRule.Path,
					TargetPort:  int64(portRule.TargetPort),
				})
			}
		}
	} else {
		if serviceType == LegacyLbServiceType {
			r.context.RancherConfig[r.name] = RancherConfig{}
			config = r.context.RancherConfig[r.name]
		} else {
			return nil
		}
	}

	// Write back to the ports passed in because the Docker parsing logic changes then
	launchConfig.Ports = r.serviceConfig.Ports
	launchConfig.Expose = r.serviceConfig.Expose

	if serviceType == LegacyLbServiceType {
		existingHAProxyConfig := ""
		var legacyStickinessPolicy *legacyClient.LoadBalancerCookieStickinessPolicy
		if config.LegacyLoadBalancerConfig != nil {
			legacyStickinessPolicy = config.LegacyLoadBalancerConfig.LbCookieStickinessPolicy
			if config.LegacyLoadBalancerConfig.HaproxyConfig != nil {
				existingHAProxyConfig = generateHAProxyConf(config.LegacyLoadBalancerConfig.HaproxyConfig.Global, config.LegacyLoadBalancerConfig.HaproxyConfig.Defaults)
			}
		}
		service.RealLbConfig = &client.LbConfig{
			CertificateIds:       config.Certs,
			Config:               string(existingHAProxyConfig),
			DefaultCertificateId: config.DefaultCert,
		}
		if legacyStickinessPolicy != nil {
			service.RealLbConfig.StickinessPolicy = &client.LoadBalancerCookieStickinessPolicy{
				Cookie:   legacyStickinessPolicy.Cookie,
				Domain:   legacyStickinessPolicy.Domain,
				Indirect: legacyStickinessPolicy.Indirect,
				Mode:     legacyStickinessPolicy.Mode,
				Name:     legacyStickinessPolicy.Name,
				Nocache:  legacyStickinessPolicy.Nocache,
				Postonly: legacyStickinessPolicy.Postonly,
			}
		}
		portRules, err := convertLb(r.serviceConfig.Ports, r.serviceConfig.Links, r.serviceConfig.ExternalLinks, "")
		if err != nil {
			return err
		}
		exposeRules, err := convertLb(r.serviceConfig.Expose, r.serviceConfig.Links, r.serviceConfig.ExternalLinks, "")
		portRules = append(portRules, exposeRules...)
		labelName := "io.rancher.service.selector.link"
		if label, ok := r.serviceConfig.Labels[labelName]; ok {
			selectorPortRules, err := convertLb(r.serviceConfig.Ports, nil, nil, label)
			if err != nil {
				return err
			}
			portRules = append(portRules, selectorPortRules...)
			selectorExposeRules, err := convertLb(r.serviceConfig.Expose, nil, nil, label)
			if err != nil {
				return err
			}
			portRules = append(portRules, selectorExposeRules...)

		}

		links, err := r.getLinks()
		if err != nil {
			return err
		}
		for link := range links {
			labelName = "io.rancher.loadbalancer.target." + link.ServiceName
			if label, ok := r.serviceConfig.Labels[labelName]; ok {
				newPortRules, err := convertLbLabel(label)
				if err != nil {
					return err
				}
				for i := range newPortRules {
					newPortRules[i].Service = link.ServiceName
				}
				portRules = mergePortRules(portRules, newPortRules)
			}
		}
		labelName = "io.rancher.loadbalancer.ssl.ports"
		if label, ok := r.serviceConfig.Labels[labelName]; ok {
			split := strings.Split(label, ",")
			for _, portString := range split {
				port, err := strconv.ParseInt(portString, 10, 32)
				if err != nil {
					return err
				}
				for i, portRule := range portRules {
					if portRule.SourcePort == int(port) {
						portRules[i].Protocol = "https"
					}
				}
			}
		}
		labelName = "io.rancher.loadbalancer.proxy-protocol.ports"
		if label, ok := r.serviceConfig.Labels[labelName]; ok {
			split := strings.Split(label, ",")
			for _, portString := range split {
				service.RealLbConfig.Config += fmt.Sprintf(`
frontend %s
    accept-proxy`, portString)
			}
		}
		for _, portRule := range portRules {
			finalPortRule := client.PortRule{
				SourcePort:  int64(portRule.SourcePort),
				Protocol:    portRule.Protocol,
				Path:        portRule.Path,
				Hostname:    portRule.Hostname,
				TargetPort:  int64(portRule.TargetPort),
				Priority:    int64(portRule.Priority),
				BackendName: portRule.BackendName,
				Selector:    portRule.Selector,
			}
			if portRule.Service != "" {
				targetService, err := r.FindExisting(portRule.Service)
				if err != nil {
					return err
				}
				if targetService == nil {
					return fmt.Errorf("Failed to find existing service: %s", portRule.Service)
				}
				finalPortRule.ServiceId = targetService.Id
			}
			service.RealLbConfig.PortRules = append(service.RealLbConfig.PortRules, finalPortRule)
		}

		// Strip target ports from lb service config
		launchConfig.Ports, err = rewritePorts(r.serviceConfig.Ports)
		if err != nil {
			return err
		}
		// Remove expose from config
		launchConfig.Expose = nil

		return populateCerts(r.context.Client, service, config.DefaultCert, config.Certs)
	} else if serviceType == LbServiceType {
		service.RealLbConfig = &client.LbConfig{
			Config: config.LbConfig.Config,
		}
		stickinessPolicy := config.LbConfig.StickinessPolicy
		if stickinessPolicy != nil {
			service.RealLbConfig.StickinessPolicy = &client.LoadBalancerCookieStickinessPolicy{
				Name:     stickinessPolicy.Name,
				Cookie:   stickinessPolicy.Cookie,
				Domain:   stickinessPolicy.Domain,
				Indirect: stickinessPolicy.Indirect,
				Nocache:  stickinessPolicy.Nocache,
				Postonly: stickinessPolicy.Postonly,
				Mode:     stickinessPolicy.Mode,
			}
		}
		for _, portRule := range config.LbConfig.PortRules {
			finalPortRule := client.PortRule{
				SourcePort:  int64(portRule.SourcePort),
				Protocol:    portRule.Protocol,
				Path:        portRule.Path,
				Hostname:    portRule.Hostname,
				TargetPort:  int64(portRule.TargetPort),
				Priority:    int64(portRule.Priority),
				BackendName: portRule.BackendName,
				Selector:    portRule.Selector,
			}

			if portRule.Service != "" {
				targetService, err := r.FindExisting(portRule.Service)
				if err != nil {
					return err
				}
				if targetService == nil {
					return fmt.Errorf("Failed to find existing service: %s", portRule.Service)
				}
				finalPortRule.ServiceId = targetService.Id
			}

			service.RealLbConfig.PortRules = append(service.RealLbConfig.PortRules, finalPortRule)
		}

		launchConfig.Ports = r.serviceConfig.Ports
		launchConfig.Expose = r.serviceConfig.Expose

		return populateCerts(r.context.Client, service, config.LbConfig.DefaultCert, config.LbConfig.Certs)
	}

	return nil
}

func generateHAProxyConf(global, defaults string) string {
	conf := ""
	if global != "" {
		conf += "global"
		global = "\n" + global
		finalGlobal := ""
		for _, c := range global {
			if c == '\n' {
				finalGlobal += "\n    "
			} else {
				finalGlobal += string(c)
			}
		}
		conf += finalGlobal
		conf += "\n"
	}
	if defaults != "" {
		conf += "defaults"
		defaults = "\n" + defaults
		finalDefaults := ""
		for _, c := range defaults {
			if c == '\n' {
				finalDefaults += "\n    "
			} else {
				finalDefaults += string(c)
			}
		}
		conf += finalDefaults
	}
	return conf

}

func rewritePorts(ports []string) ([]string, error) {
	updatedPorts := []string{}

	for _, port := range ports {
		protocol := ""
		split := strings.Split(port, "/")
		if len(split) == 2 {
			protocol = split[1]
		}

		var source string
		var err error
		split = strings.Split(port, ":")
		if len(split) == 1 {
			source, _, err = readPort(split[0], 0)
			if err != nil {
				return nil, err
			}
		} else if len(split) == 2 {
			source = split[0]
		}

		if protocol == "" {
			updatedPorts = append(updatedPorts, source)
		} else {
			updatedPorts = append(updatedPorts, fmt.Sprintf("%s/%s", source, protocol))
		}
	}

	return updatedPorts, nil
}

func convertLb(ports, links, externalLinks []string, selector string) ([]PortRule, error) {
	portRules := []PortRule{}

	for _, port := range ports {
		protocol := "http"
		split := strings.Split(port, "/")
		if len(split) == 2 {
			protocol = split[1]
		}

		var sourcePort int64
		var targetPort int64
		var err error
		split = strings.Split(port, ":")
		if len(split) == 1 {
			singlePort, _, err := readPort(split[0], 0)
			if err != nil {
				return nil, err
			}
			sourcePort, err = strconv.ParseInt(singlePort, 10, 32)
			if err != nil {
				return nil, err
			}
			targetPort, err = strconv.ParseInt(singlePort, 10, 32)
			if err != nil {
				return nil, err
			}
		} else if len(split) == 2 {
			sourcePort, err = strconv.ParseInt(split[0], 10, 32)
			if err != nil {
				return nil, err
			}
			target, _, err := readPort(split[1], 0)
			if err != nil {
				return nil, err
			}
			targetPort, err = strconv.ParseInt(target, 10, 32)
			if err != nil {
				return nil, err
			}
		}
		for _, link := range links {
			split := strings.Split(link, ":")
			portRules = append(portRules, PortRule{
				SourcePort: int(sourcePort),
				TargetPort: int(targetPort),
				Service:    split[0],
				Protocol:   protocol,
			})
		}
		for _, externalLink := range externalLinks {
			split := strings.Split(externalLink, ":")
			portRules = append(portRules, PortRule{
				SourcePort: int(sourcePort),
				TargetPort: int(targetPort),
				Service:    split[0],
				Protocol:   protocol,
			})
		}
		if selector != "" {
			portRules = append(portRules, PortRule{
				SourcePort: int(sourcePort),
				TargetPort: int(targetPort),
				Selector:   selector,
				Protocol:   protocol,
			})
		}
	}

	return portRules, nil
}

func isNum(c uint8) bool {
	return c >= '0' && c <= '9'
}

func readHostname(label string, pos int) (string, int, error) {
	var hostname bytes.Buffer
	if isNum(label[pos]) {
		return hostname.String(), pos, nil
	}
	for ; pos < len(label); pos++ {
		c := label[pos]
		if c == '=' {
			return hostname.String(), pos + 1, nil
		}
		if c == ':' {
			return hostname.String(), pos + 1, nil
		}
		if c == '/' {
			return hostname.String(), pos, nil
		}
		hostname.WriteByte(c)
	}
	return hostname.String(), pos, nil
}

func readPort(label string, pos int) (string, int, error) {
	var port bytes.Buffer
	for ; pos < len(label); pos++ {
		c := label[pos]
		if !isNum(c) {
			return port.String(), pos, nil
		}
		port.WriteByte(c)
	}
	return port.String(), pos, nil
}

func readPath(label string, pos int) (string, int, error) {
	var path bytes.Buffer
	for ; pos < len(label); pos++ {
		c := label[pos]
		if c == '=' {
			return path.String(), pos + 1, nil
		}
		path.WriteByte(c)
	}
	return path.String(), pos, nil
}

func convertLbLabel(label string) ([]PortRule, error) {
	var portRules []PortRule

	labels := strings.Split(label, ",")
	for _, label := range labels {
		label = strings.Trim(label, " \t\n")

		hostname, pos, err := readHostname(label, 0)
		if err != nil {
			return nil, err
		}

		sourcePort, pos, err := readPort(label, pos)
		if err != nil {
			return nil, err
		}

		path, pos, err := readPath(label, pos)
		if err != nil {
			return nil, err
		}

		targetPort, pos, err := readPort(label, pos)
		if err != nil {
			return nil, err
		}

		var source int64
		if sourcePort == "" {
			source = 0
		} else {
			source, err = strconv.ParseInt(sourcePort, 10, 32)
			if err != nil {
				return nil, err
			}
		}

		var target int64
		if targetPort == "" {
			target = 0
		} else {
			target, err = strconv.ParseInt(targetPort, 10, 32)
			if err != nil {
				return nil, err
			}
		}

		if hostname == "" && path == "" && target == 0 {
			portRules = append(portRules, PortRule{
				TargetPort: int(source),
			})
			continue
		}

		if target == 0 && strings.Contains(label, "=") {
			portRules = append(portRules, PortRule{
				Hostname:   hostname,
				TargetPort: int(source),
			})
			continue
		}

		portRules = append(portRules, PortRule{
			Hostname:   hostname,
			SourcePort: int(source),
			Path:       path,
			TargetPort: int(target),
		})
	}

	return portRules, nil
}

func mergePortRules(baseRules, overrideRules []PortRule) []PortRule {
	newRules := []PortRule{}
	for _, baseRule := range baseRules {
		prevLength := len(newRules)
		for _, overrideRule := range overrideRules {
			if baseRule.Service == overrideRule.Service && (overrideRule.SourcePort == 0 || baseRule.SourcePort == overrideRule.SourcePort) {
				newRule := baseRule
				newRule.Path = overrideRule.Path
				newRule.Hostname = overrideRule.Hostname
				if overrideRule.TargetPort != 0 {
					newRule.TargetPort = overrideRule.TargetPort
				}
				newRules = append(newRules, newRule)
			}
		}
		// If no rules were overidden, just copy over base rule
		if len(newRules) == prevLength {
			newRules = append(newRules, baseRule)
		}
	}
	return newRules
}

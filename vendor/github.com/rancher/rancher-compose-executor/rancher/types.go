package rancher

import (
	"github.com/rancher/go-rancher/v2"
	"github.com/rancher/rancher-compose-executor/config"
)

const (
	LB_IMAGE       = "rancher/load-balancer-service"
	DNS_IMAGE      = "rancher/dns-service"
	EXTERNAL_IMAGE = "rancher/external-service"

	RancherType         = ServiceType(iota)
	LegacyLbServiceType = ServiceType(iota)
	LbServiceType       = ServiceType(iota)
	DnsServiceType      = ServiceType(iota)
	ExternalServiceType = ServiceType(iota)
	StorageDriverType   = ServiceType(iota)
	NetworkDriverType   = ServiceType(iota)
)

type ServiceType int

func FindServiceType(r *RancherService) ServiceType {
	if r.serviceConfig.Image == EXTERNAL_IMAGE {
		return ExternalServiceType
	} else if r.serviceConfig.Image == LB_IMAGE {
		return LegacyLbServiceType
	} else if isLbServiceType(r.serviceConfig.LbConfig) {
		return LbServiceType
	} else if r.serviceConfig.Image == DNS_IMAGE {
		return DnsServiceType
	} else if r.serviceConfig.NetworkDriver != nil {
		return NetworkDriverType
	} else if r.serviceConfig.StorageDriver != nil {
		return StorageDriverType
	}

	return RancherType
}

func isLbServiceType(lbConfig *config.LBConfig) bool {
	if lbConfig == nil {
		return false
	}

	for _, portRule := range lbConfig.PortRules {
		if portRule.SourcePort != 0 {
			return true
		}
	}

	return false
}

type CompositeService struct {
	client.Service

	StorageDriver *client.StorageDriver `json:"storageDriver,omitempty" yaml:"storageDriver,omitempty"`
	NetworkDriver *client.NetworkDriver `json:"networkDriver,omitempty" yaml:"networkDriver,omitempty"`
	RealLbConfig  *client.LbConfig      `json:"lbConfig,omitempty" yaml:"lb_config,omitempty"`

	// External Service Fields
	ExternalIpAddresses []string                    `json:"externalIpAddresses,omitempty" yaml:"external_ip_addresses,omitempty"`
	Hostname            string                      `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	HealthCheck         *client.InstanceHealthCheck `json:"healthCheck,omitempty" yaml:"health_check,omitempty"`
}

package rancher

import rancherClient "github.com/rancher/go-rancher/client"

const (
	LB_IMAGE       = "rancher/load-balancer-service"
	DNS_IMAGE      = "rancher/dns-service"
	EXTERNAL_IMAGE = "rancher/external-service"

	RancherType         = ServiceType(iota)
	LbServiceType       = ServiceType(iota)
	DnsServiceType      = ServiceType(iota)
	ExternalServiceType = ServiceType(iota)
)

type ServiceType int

func FindServiceType(r *RancherService) ServiceType {
	rancherConfig := r.RancherConfig()

	if len(rancherConfig.ExternalIps) > 0 || rancherConfig.Hostname != "" {
		return ExternalServiceType
	} else if r.serviceConfig.Image == LB_IMAGE {
		return LbServiceType
	} else if r.serviceConfig.Image == DNS_IMAGE {
		return DnsServiceType
	}

	return RancherType
}

type CompositeService struct {
	rancherClient.Service

	//LoadBalancer Fields
	CertificateIds       []string                          `json:"certificateIds,omitempty" yaml:"certificate_ids,omitempty"`
	DefaultCertificateId string                            `json:"defaultCertificateId,omitempty" yaml:"default_certificate_id,omitempty"`
	LoadBalancerConfig   *rancherClient.LoadBalancerConfig `json:"loadBalancerConfig,omitempty" yaml:"load_balancer_config,omitempty"`

	// External Service Fields
	ExternalIpAddresses []string                           `json:"externalIpAddresses,omitempty" yaml:"external_ip_addresses,omitempty"`
	Hostname            string                             `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	HealthCheck         *rancherClient.InstanceHealthCheck `json:"healthCheck,omitempty" yaml:"health_check,omitempty"`
}

package rancher

import (
	"fmt"

	"github.com/rancher/go-rancher/v2"
)

func populateCerts(apiClient *client.RancherClient, lbService *CompositeService, rancherConfig *RancherConfig) error {
	if rancherConfig.DefaultCert != "" {
		if certId, err := findCertByName(apiClient, rancherConfig.DefaultCert); err != nil {
			return err
		} else {
			lbService.DefaultCertificateId = certId
		}
	}

	lbService.CertificateIds = []string{}
	for _, certName := range rancherConfig.Certs {
		if certId, err := findCertByName(apiClient, certName); err != nil {
			return err

		} else {
			lbService.CertificateIds = append(lbService.CertificateIds, certId)
		}
	}

	return nil
}

func findCertByName(apiClient *client.RancherClient, name string) (string, error) {
	certs, err := apiClient.Certificate.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"removed_null": nil,
			"name":         name,
		},
	})

	if err != nil {
		return "", err
	}

	if len(certs.Data) == 0 {
		return "", fmt.Errorf("Failed to find certificate %s", name)
	}

	return certs.Data[0].Id, nil
}

package rancher

import (
	"fmt"

	rancherClient "github.com/rancher/go-rancher/client"
)

func populateCerts(client *rancherClient.RancherClient, lbService *CompositeService, rancherConfig *RancherConfig) error {
	if rancherConfig.DefaultCert != "" {
		if certId, err := findCertByName(client, rancherConfig.DefaultCert); err != nil {
			return err
		} else {
			lbService.DefaultCertificateId = certId
		}
	}

	lbService.CertificateIds = []string{}
	for _, certName := range rancherConfig.Certs {
		if certId, err := findCertByName(client, certName); err != nil {
			return err

		} else {
			lbService.CertificateIds = append(lbService.CertificateIds, certId)
		}
	}

	return nil
}

func findCertByName(client *rancherClient.RancherClient, name string) (string, error) {
	certs, err := client.Certificate.List(&rancherClient.ListOpts{
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

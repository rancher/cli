package rancher

import (
	"fmt"

	"github.com/rancher/go-rancher/v2"
)

func populateCerts(apiClient *client.RancherClient, lbService *CompositeService, defaultCert string, certs []string) error {
	if defaultCert != "" {
		certId, err := findCertByName(apiClient, defaultCert)
		if err != nil {
			return err
		}
		lbService.RealLbConfig.DefaultCertificateId = certId
	}

	certIds := []string{}
	for _, certName := range certs {
		certId, err := findCertByName(apiClient, certName)
		if err != nil {
			return err
		}
		certIds = append(certIds, certId)
	}
	lbService.RealLbConfig.CertificateIds = certIds

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

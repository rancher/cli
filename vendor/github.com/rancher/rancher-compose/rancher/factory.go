package rancher

import "github.com/rancher/rancher-compose/digest"

type Factory interface {
	Hash(service *RancherService) (digest.ServiceHash, error)
	Create(service *RancherService) error
	Upgrade(r *RancherService, force bool, selected []string) error
	Rollback(r *RancherService) error
}

func GetFactory(service *RancherService) (Factory, error) {
	return &NormalFactory{}, nil
}

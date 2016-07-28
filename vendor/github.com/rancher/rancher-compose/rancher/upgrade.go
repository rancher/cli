package rancher

import (
	"github.com/Sirupsen/logrus"
	rancherClient "github.com/rancher/go-rancher/client"
	"github.com/rancher/rancher-compose/digest"
)

func (r *RancherService) upgrade(service *rancherClient.Service, force bool, selected []string) (*rancherClient.Service, error) {
	factory, err := GetFactory(r)
	if err != nil {
		return nil, err
	}

	if err := factory.Upgrade(r, force, selected); err != nil {
		return nil, err
	}

	return r.FindExisting(r.name)
}

func (r *RancherService) rollback(service *rancherClient.Service) (*rancherClient.Service, error) {
	factory, err := GetFactory(r)
	if err != nil {
		return nil, err
	}

	if err := factory.Rollback(r); err != nil {
		return nil, err
	}

	return r.FindExisting(r.name)
}

func (r *RancherService) shouldUpgrade(service *rancherClient.Service) bool {
	switch FindServiceType(r) {
	case ExternalServiceType:
		return false
	case DnsServiceType:
		return false
	}

	if service == nil {
		return false
	}

	if r.context.ForceUpgrade {
		return true
	}

	if r.isOutOfSync(service) {
		if r.context.Upgrade {
			return true
		} else if hasOldHash(service) {
			logrus.Warnf("Service %s is out of sync with local configuration file", r.name)
			return false
		}
	}

	return false
}

func (r *RancherService) isOutOfSync(service *rancherClient.Service) bool {
	if service == nil {
		return false
	}

	hash, ok := digest.LookupHash(service)
	if !ok {
		return true
	}

	factory, err := GetFactory(r)
	if err != nil {
		logrus.Errorf("Failed to find factory to service %s: %v", r.name, err)
		return false
	}

	newHash, err := factory.Hash(r)
	if err != nil {
		logrus.Errorf("Failed to calculate hash for service %s: %v", r.name, err)
		return false
	}

	logrus.Debugf("Comparing hashes for %s: old: %#v new: %#v", r.name, hash, newHash)
	return !hash.Equals(newHash)
}

func hasOldHash(service *rancherClient.Service) bool {
	_, ok := digest.LookupHash(service)
	return ok
}

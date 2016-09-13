package rancher

import (
	"github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher/v2"
	"github.com/rancher/rancher-compose/digest"
)

func (r *RancherService) upgrade(service *client.Service, force bool, selected []string) (*client.Service, error) {
	factory, err := GetFactory(r)
	if err != nil {
		return nil, err
	}

	if err := factory.Upgrade(r, force, selected); err != nil {
		return nil, err
	}

	return r.FindExisting(r.name)
}

func (r *RancherService) rollback(service *client.Service) (*client.Service, error) {
	factory, err := GetFactory(r)
	if err != nil {
		return nil, err
	}

	if err := factory.Rollback(r); err != nil {
		return nil, err
	}

	return r.FindExisting(r.name)
}

func (r *RancherService) shouldUpgrade(service *client.Service) bool {
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

func (r *RancherService) isOutOfSync(service *client.Service) bool {
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

func hasOldHash(service *client.Service) bool {
	_, ok := digest.LookupHash(service)
	return ok
}

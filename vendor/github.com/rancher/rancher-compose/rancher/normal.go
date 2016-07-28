package rancher

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/utils"
	rancherClient "github.com/rancher/go-rancher/client"
	"github.com/rancher/rancher-compose/digest"
)

type NormalFactory struct {
}

func (f *NormalFactory) Hash(service *RancherService) (digest.ServiceHash, error) {
	hash, _, err := f.configAndHash(service)
	return hash, err
}

func (f *NormalFactory) configAndHash(r *RancherService) (digest.ServiceHash, *CompositeService, error) {
	rancherService, launchConfig, secondaryLaunchConfigs, err := f.config(r)
	if err != nil {
		return digest.ServiceHash{}, nil, err
	}

	hash, err := digest.CreateServiceHash(rancherService, launchConfig, secondaryLaunchConfigs)
	if err != nil {
		return digest.ServiceHash{}, nil, err
	}

	rancherService.LaunchConfig = launchConfig
	rancherService.LaunchConfig.Labels[digest.ServiceHashKey] = hash.LaunchConfig
	rancherService.SecondaryLaunchConfigs = []interface{}{}
	rancherService.Metadata[digest.ServiceHashKey] = hash.Service

	for _, secondaryLaunchConfig := range secondaryLaunchConfigs {
		secondaryLaunchConfig.Labels[digest.ServiceHashKey] = hash.SecondaryLaunchConfigs[secondaryLaunchConfig.Name]
		rancherService.SecondaryLaunchConfigs = append(rancherService.SecondaryLaunchConfigs, secondaryLaunchConfig)
	}

	return hash, rancherService, nil
}

func (f *NormalFactory) config(r *RancherService) (*CompositeService, *rancherClient.LaunchConfig, []rancherClient.SecondaryLaunchConfig, error) {
	launchConfig, secondaryLaunchConfigs, err := createLaunchConfigs(r)
	if err != nil {
		return nil, nil, nil, err
	}

	rancherConfig, _ := r.context.RancherConfig[r.name]

	service := &CompositeService{
		Service: rancherClient.Service{
			Name:              r.name,
			Metadata:          r.Metadata(),
			Scale:             int64(r.getConfiguredScale()),
			ScalePolicy:       rancherConfig.ScalePolicy,
			RetainIp:          rancherConfig.RetainIp,
			EnvironmentId:     r.Context().Environment.Id,
			SelectorContainer: r.SelectorContainer(),
			SelectorLink:      r.SelectorLink(),
		},
		ExternalIpAddresses: rancherConfig.ExternalIps,
		Hostname:            rancherConfig.Hostname,
		HealthCheck:         r.HealthCheck(""),
	}

	if err := populateLbFields(r, &launchConfig, service); err != nil {
		return nil, nil, nil, err
	}

	return service, &launchConfig, secondaryLaunchConfigs, nil
}

func (f *NormalFactory) Create(r *RancherService) error {
	hash, service, err := f.configAndHash(r)
	if err != nil {
		return err
	}

	logrus.Debugf("Creating service %s with hash: %#v", r.name, hash)
	switch FindServiceType(r) {
	case ExternalServiceType:
		return r.context.Client.Create(rancherClient.EXTERNAL_SERVICE_TYPE, &service, nil)
	case DnsServiceType:
		return r.context.Client.Create(rancherClient.DNS_SERVICE_TYPE, &service, nil)
	case LbServiceType:
		return r.context.Client.Create(rancherClient.LOAD_BALANCER_SERVICE_TYPE, &service, nil)
	default:
		_, err = r.context.Client.Service.Create(&service.Service)
	}

	return err
}

func (f *NormalFactory) Rollback(r *RancherService) error {
	existingService, err := r.FindExisting(r.Name())
	if err != nil || existingService == nil {
		return err
	}

	existingService, err = r.Client().Service.ActionRollback(existingService)
	if err != nil {
		return err
	}

	return r.Wait(existingService)
}

func isForce(name string, force bool, selected []string) bool {
	if !force {
		return false
	}
	if len(selected) == 0 {
		return true
	}

	return utils.Contains(selected, name)
}

func (f *NormalFactory) Upgrade(r *RancherService, force bool, selected []string) error {
	existingService, err := r.FindExisting(r.Name())
	if err != nil || existingService == nil {
		return err
	}

	if existingService.State != "active" && existingService.State != "inactive" {
		return fmt.Errorf("Service %s must be state=active or inactive to upgrade, currently: state=%s", r.Name(), existingService.State)
	}

	existingHash, _ := digest.LookupHash(existingService)
	secondaryNames := []string{}
	removedSecondaryNames := []string{}

	hash, err := f.Hash(r)
	if err != nil {
		return err
	}

	service := hash.Service != existingHash.Service || isForce(r.Name(), force, selected)
	launchConfig := hash.LaunchConfig != existingHash.LaunchConfig || isForce(r.Name(), force, selected)
	for oldSecondary, _ := range existingHash.SecondaryLaunchConfigs {
		if _, ok := hash.SecondaryLaunchConfigs[oldSecondary]; !ok {
			removedSecondaryNames = append(removedSecondaryNames, oldSecondary)
		}
	}
	for newSecondary, newHash := range hash.SecondaryLaunchConfigs {
		if oldHash, ok := existingHash.SecondaryLaunchConfigs[newSecondary]; ok {
			if oldHash != newHash || isForce(newSecondary, force, selected) {
				secondaryNames = append(secondaryNames, newSecondary)
			}
		} else {
			secondaryNames = append(secondaryNames, newSecondary)
		}
	}

	return f.upgrade(r, existingService, service, launchConfig, secondaryNames, removedSecondaryNames)
}

func (f *NormalFactory) upgrade(r *RancherService, existingService *rancherClient.Service, service, launchConfig bool, secondaryNames, removedSecondaryNames []string) error {
	_, config, err := f.configAndHash(r)
	if err != nil {
		return err
	}

	serviceUpgrade := &rancherClient.ServiceUpgrade{
		InServiceStrategy: &rancherClient.InServiceUpgradeStrategy{
			BatchSize:      r.context.BatchSize,
			IntervalMillis: r.context.Interval,
			StartFirst:     r.RancherConfig().UpgradeStrategy.StartFirst,
		},
	}

	serviceUpgrade.InServiceStrategy.SecondaryLaunchConfigs = []interface{}{}

	if launchConfig {
		serviceUpgrade.InServiceStrategy.LaunchConfig = config.LaunchConfig
	}

	for _, name := range secondaryNames {
		for _, v := range config.SecondaryLaunchConfigs {
			if secondaryLaunchConfig, ok := v.(rancherClient.SecondaryLaunchConfig); ok {
				if secondaryLaunchConfig.Name == name {
					serviceUpgrade.InServiceStrategy.SecondaryLaunchConfigs = append(serviceUpgrade.InServiceStrategy.SecondaryLaunchConfigs, secondaryLaunchConfig)
				}
			}
		}
	}

	for _, removedSecondaryName := range removedSecondaryNames {
		serviceUpgrade.InServiceStrategy.SecondaryLaunchConfigs = append(serviceUpgrade.InServiceStrategy.SecondaryLaunchConfigs, &rancherClient.SecondaryLaunchConfig{
			Name:      removedSecondaryName,
			ImageUuid: "rancher/none",
		})
	}

	if service {
		// Scale must be changed through "scale" not "up", so always copy scale existing scale
		config.Scale = existingService.Scale

		logrus.Infof("Updating %s", r.Name())
		schemaType := rancherClient.SERVICE_TYPE
		switch FindServiceType(r) {
		case ExternalServiceType:
			schemaType = rancherClient.EXTERNAL_SERVICE_TYPE
		case DnsServiceType:
			schemaType = rancherClient.DNS_SERVICE_TYPE
		case LbServiceType:
			schemaType = rancherClient.LOAD_BALANCER_SERVICE_TYPE
		}

		if err := r.context.Client.Update(schemaType, &existingService.Resource, config, existingService); err != nil {
			return err
		}

		if err := r.Wait(existingService); err != nil {
			return err
		}
	}

	if launchConfig || len(secondaryNames) > 0 {
		logrus.Infof("Upgrading %s", r.Name())
		existingService, err = r.Client().Service.ActionUpgrade(existingService, serviceUpgrade)
		if err != nil {
			return err
		}
	}

	return r.Wait(existingService)
}

package rancher

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/utils"
	"github.com/rancher/go-rancher/v2"
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
	rancherService.Metadata[digest.ServiceHashKey] = hash.Service

	for _, secondaryLaunchConfig := range secondaryLaunchConfigs {
		secondaryLaunchConfig.Labels[digest.ServiceHashKey] = hash.SecondaryLaunchConfigs[secondaryLaunchConfig.Name]
		rancherService.SecondaryLaunchConfigs = append(rancherService.SecondaryLaunchConfigs, secondaryLaunchConfig)
	}

	return hash, rancherService, nil
}

func (f *NormalFactory) config(r *RancherService) (*CompositeService, *client.LaunchConfig, []client.SecondaryLaunchConfig, error) {
	launchConfig, secondaryLaunchConfigs, err := createLaunchConfigs(r)
	if err != nil {
		return nil, nil, nil, err
	}

	rancherConfig, _ := r.context.RancherConfig[r.name]

	service := &CompositeService{
		Service: client.Service{
			Name:              r.name,
			Metadata:          r.Metadata(),
			Scale:             int64(r.getConfiguredScale()),
			ScalePolicy:       rancherConfig.ScalePolicy,
			RetainIp:          rancherConfig.RetainIp,
			StackId:           r.Context().Stack.Id,
			SelectorContainer: r.SelectorContainer(),
			SelectorLink:      r.SelectorLink(),
		},
		ExternalIpAddresses: rancherConfig.ExternalIps,
		Hostname:            rancherConfig.Hostname,
		HealthCheck:         r.HealthCheck(""),
		StorageDriver:       rancherConfig.StorageDriver,
		NetworkDriver:       rancherConfig.NetworkDriver,
	}

	if service.NetworkDriver != nil {
		service.NetworkDriver.CniConfig = convertNestedMapKeysToStrings(service.NetworkDriver.CniConfig)
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
		return r.context.Client.Create(client.EXTERNAL_SERVICE_TYPE, &service, nil)
	case DnsServiceType:
		return r.context.Client.Create(client.DNS_SERVICE_TYPE, &service, nil)
	case LegacyLbServiceType:
		fallthrough
	case LbServiceType:
		return r.context.Client.Create(client.LOAD_BALANCER_SERVICE_TYPE, &service, nil)
	case StorageDriverType:
		return r.context.Client.Create(client.STORAGE_DRIVER_SERVICE_TYPE, &service, nil)
	case NetworkDriverType:
		return r.context.Client.Create(client.NETWORK_DRIVER_SERVICE_TYPE, &service, nil)
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

func (f *NormalFactory) upgrade(r *RancherService, existingService *client.Service, service, launchConfig bool, secondaryNames, removedSecondaryNames []string) error {
	_, config, err := f.configAndHash(r)
	if err != nil {
		return err
	}

	serviceUpgrade := &client.ServiceUpgrade{
		InServiceStrategy: &client.InServiceUpgradeStrategy{
			BatchSize:      r.context.BatchSize,
			IntervalMillis: r.context.Interval,
			StartFirst:     r.RancherConfig().UpgradeStrategy.StartFirst,
		},
	}

	if launchConfig {
		serviceUpgrade.InServiceStrategy.LaunchConfig = config.LaunchConfig
	}

	for _, name := range secondaryNames {
		for _, secondaryLaunchConfig := range config.SecondaryLaunchConfigs {
			if secondaryLaunchConfig.Name == name {
				serviceUpgrade.InServiceStrategy.SecondaryLaunchConfigs = append(serviceUpgrade.InServiceStrategy.SecondaryLaunchConfigs, secondaryLaunchConfig)
			}
		}
	}

	for _, removedSecondaryName := range removedSecondaryNames {
		serviceUpgrade.InServiceStrategy.SecondaryLaunchConfigs = append(serviceUpgrade.InServiceStrategy.SecondaryLaunchConfigs, client.SecondaryLaunchConfig{
			Name:      removedSecondaryName,
			ImageUuid: "rancher/none",
		})
	}

	if service {
		// Scale must be changed through "scale" not "up", so always copy scale existing scale
		config.Scale = existingService.Scale

		logrus.Infof("Updating %s", r.Name())
		schemaType := client.SERVICE_TYPE
		switch FindServiceType(r) {
		case ExternalServiceType:
			schemaType = client.EXTERNAL_SERVICE_TYPE
		case DnsServiceType:
			schemaType = client.DNS_SERVICE_TYPE
		case LbServiceType:
			schemaType = client.LOAD_BALANCER_SERVICE_TYPE
		case NetworkDriverType:
			schemaType = client.NETWORK_DRIVER_SERVICE_TYPE
		case StorageDriverType:
			schemaType = client.STORAGE_DRIVER_SERVICE_TYPE
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

func convertNestedMapKeysToStrings(service map[string]interface{}) map[string]interface{} {
	newService := map[string]interface{}{}

	for k, v := range service {
		newService[k] = convertKeysToStrings(v)
	}

	return newService
}

func convertKeysToStrings(item interface{}) interface{} {
	switch typedDatas := item.(type) {

	case map[interface{}]interface{}:
		newMap := make(map[string]interface{})

		for key, value := range typedDatas {
			stringKey := key.(string)
			newMap[stringKey] = convertKeysToStrings(value)
		}
		return newMap

	case []interface{}:
		// newArray := make([]interface{}, 0) will cause golint to complain
		var newArray []interface{}
		newArray = make([]interface{}, 0)

		for _, value := range typedDatas {
			newArray = append(newArray, convertKeysToStrings(value))
		}
		return newArray

	default:
		return item
	}
}

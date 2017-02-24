package rancher

import (
	"fmt"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/libcompose/utils"
	"github.com/docker/libcompose/yaml"
	"github.com/rancher/go-rancher/v2"
	"github.com/rancher/rancher-compose-executor/config"
	"github.com/rancher/rancher-compose-executor/convert"
)

func createLaunchConfigs(r *RancherService) (client.LaunchConfig, []client.SecondaryLaunchConfig, error) {
	secondaryLaunchConfigs := []client.SecondaryLaunchConfig{}
	launchConfig, err := createLaunchConfig(r, r.Name(), r.Config())
	if err != nil {
		return launchConfig, nil, err
	}
	launchConfig.HealthCheck = r.HealthCheck("")

	if secondaries, ok := r.Context().SidekickInfo.primariesToSidekicks[r.Name()]; ok {
		for _, secondaryName := range secondaries {
			serviceConfig, ok := r.Context().Project.ServiceConfigs.Get(secondaryName)
			if !ok {
				return launchConfig, nil, fmt.Errorf("Failed to find sidekick: %s", secondaryName)
			}

			launchConfig, err := createLaunchConfig(r, secondaryName, serviceConfig)
			if err != nil {
				return launchConfig, nil, err
			}
			launchConfig.HealthCheck = r.HealthCheck(secondaryName)

			var secondaryLaunchConfig client.SecondaryLaunchConfig
			utils.Convert(launchConfig, &secondaryLaunchConfig)
			secondaryLaunchConfig.Name = secondaryName

			if secondaryLaunchConfig.Labels == nil {
				secondaryLaunchConfig.Labels = map[string]interface{}{}
			}
			secondaryLaunchConfigs = append(secondaryLaunchConfigs, secondaryLaunchConfig)
		}
	}

	return launchConfig, secondaryLaunchConfigs, nil
}

func createLaunchConfig(r *RancherService, name string, serviceConfig *config.ServiceConfig) (client.LaunchConfig, error) {
	var result client.LaunchConfig

	rancherConfig := r.context.RancherConfig[name]

	schemasUrl := strings.SplitN(r.Context().Client.GetSchemas().Links["self"], "/schemas", 2)[0]
	scriptsUrl := schemasUrl + "/scripts/transform"

	tempImage := serviceConfig.Image
	tempLabels := serviceConfig.Labels
	newLabels := yaml.SliceorMap{}
	if serviceConfig.Image == "rancher/load-balancer-service" {
		// Lookup default load balancer image
		lbImageSetting, err := r.Client().Setting.ById("lb.instance.image")
		if err != nil {
			return result, err
		}
		serviceConfig.Image = lbImageSetting.Value

		// Strip off legacy load balancer labels
		for k, v := range serviceConfig.Labels {
			if !strings.HasPrefix(k, "io.rancher.loadbalancer") && !strings.HasPrefix(k, "io.rancher.service.selector") {
				newLabels[k] = v
			}
		}
		serviceConfig.Labels = newLabels
	}

	config, hostConfig, err := convert.Convert(serviceConfig, r.context.Context)
	if err != nil {
		return result, err
	}

	serviceConfig.Image = tempImage
	serviceConfig.Labels = tempLabels

	dockerContainer := &ContainerInspect{
		Config:     config,
		HostConfig: hostConfig,
	}

	dockerContainer.HostConfig.NetworkMode = container.NetworkMode("")
	dockerContainer.Name = "/" + name

	err = r.Context().Client.Post(scriptsUrl, dockerContainer, &result)
	if err != nil {
		return result, err
	}

	result.VolumeDriver = hostConfig.VolumeDriver

	setupNetworking(serviceConfig.NetworkMode, &result)
	setupVolumesFrom(serviceConfig.VolumesFrom, &result)

	err = setupBuild(r, name, &result, serviceConfig)

	if result.Labels == nil {
		result.Labels = map[string]interface{}{}
	}

	result.Kind = rancherConfig.Type
	result.Vcpu = int64(rancherConfig.Vcpu)
	result.Userdata = rancherConfig.Userdata
	result.MemoryMb = int64(rancherConfig.Memory)
	result.Disks = rancherConfig.Disks

	if strings.EqualFold(result.Kind, "virtual_machine") || strings.EqualFold(result.Kind, "virtualmachine") {
		result.Kind = "virtualMachine"
	}

	if result.LogConfig.Config == nil {
		result.LogConfig.Config = map[string]interface{}{}
	}

	return result, err
}

func setupNetworking(netMode string, launchConfig *client.LaunchConfig) {
	if netMode == "" {
		launchConfig.NetworkMode = "managed"
	} else if container.IpcMode(netMode).IsContainer() {
		// For some reason NetworkMode object is gone runconfig, but IpcMode works the same for this
		launchConfig.NetworkMode = "container"
		launchConfig.NetworkLaunchConfig = strings.TrimPrefix(netMode, "container:")
	} else {
		launchConfig.NetworkMode = netMode
	}
}

func setupVolumesFrom(volumesFrom []string, launchConfig *client.LaunchConfig) {
	launchConfig.DataVolumesFromLaunchConfigs = volumesFrom
}

func setupBuild(r *RancherService, name string, result *client.LaunchConfig, serviceConfig *config.ServiceConfig) error {
	if serviceConfig.Build.Context != "" {
		result.Build = &client.DockerBuild{
			Remote:     serviceConfig.Build.Context,
			Dockerfile: serviceConfig.Build.Dockerfile,
		}

		needBuild := true
		if config.IsValidRemote(serviceConfig.Build.Context) {
			needBuild = false
		}

		if needBuild {
			image, url, err := Upload(r.Context(), name)
			if err != nil {
				return err
			}
			logrus.Infof("Build for %s available at %s", name, url)
			serviceConfig.Build.Context = url

			if serviceConfig.Image == "" {
				serviceConfig.Image = image
			}

			result.Build = &client.DockerBuild{
				Context:    url,
				Dockerfile: serviceConfig.Build.Dockerfile,
			}
			result.ImageUuid = "docker:" + image
		} else if result.ImageUuid == "" {
			result.ImageUuid = fmt.Sprintf("docker:%s_%s_%d", r.Context().ProjectName, name, time.Now().UnixNano()/int64(time.Millisecond))
		}
	}

	return nil
}

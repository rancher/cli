package rancher

import (
	"fmt"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/utils"
	rancherClient "github.com/rancher/go-rancher/client"
)

func createLaunchConfigs(r *RancherService) (rancherClient.LaunchConfig, []rancherClient.SecondaryLaunchConfig, error) {
	secondaryLaunchConfigs := []rancherClient.SecondaryLaunchConfig{}
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

			var secondaryLaunchConfig rancherClient.SecondaryLaunchConfig
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

func createLaunchConfig(r *RancherService, name string, serviceConfig *config.ServiceConfig) (rancherClient.LaunchConfig, error) {
	var result rancherClient.LaunchConfig

	rancherConfig := r.context.RancherConfig[name]

	schemasUrl := strings.SplitN(r.Context().Client.Schemas.Links["self"], "/schemas", 2)[0]
	scriptsUrl := schemasUrl + "/scripts/transform"

	config, hostConfig, err := docker.Convert(serviceConfig, r.context.Context)
	if err != nil {
		return result, err
	}

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
	result.Vcpu = rancherConfig.Vcpu
	result.Userdata = rancherConfig.Userdata
	result.MemoryMb = rancherConfig.Memory
	result.Disks = []interface{}{}
	for _, i := range rancherConfig.Disks {
		result.Disks = append(result.Disks, i)
	}

	if strings.EqualFold(result.Kind, "virtual_machine") || strings.EqualFold(result.Kind, "virtualmachine") {
		result.Kind = "virtualMachine"
	}

	if result.LogConfig.Config == nil {
		result.LogConfig.Config = map[string]interface{}{}
	}

	return result, err
}

func setupNetworking(netMode string, launchConfig *rancherClient.LaunchConfig) {
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

func setupVolumesFrom(volumesFrom []string, launchConfig *rancherClient.LaunchConfig) {
	launchConfig.DataVolumesFromLaunchConfigs = volumesFrom
}

func setupBuild(r *RancherService, name string, result *rancherClient.LaunchConfig, serviceConfig *config.ServiceConfig) error {
	if serviceConfig.Build.Context != "" {
		result.Build = &rancherClient.DockerBuild{
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

			result.Build = &rancherClient.DockerBuild{
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

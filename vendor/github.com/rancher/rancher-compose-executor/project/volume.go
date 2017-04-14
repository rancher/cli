package project

import (
	"golang.org/x/net/context"

	"github.com/rancher/rancher-compose-executor/config"
)

type Volumes interface {
	Initialize(ctx context.Context) error
}

type VolumesFactory interface {
	Create(projectName string, volumeConfigs map[string]*config.VolumeConfig, serviceConfigs *config.ServiceConfigs) (Volumes, error)
}

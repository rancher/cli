package project

import (
	"github.com/rancher/rancher-compose-executor/config"
	"golang.org/x/net/context"
)

type Hosts interface {
	Initialize(ctx context.Context) error
}

type HostsFactory interface {
	Create(projectName string, hostConfigs map[string]*config.HostConfig) (Hosts, error)
}

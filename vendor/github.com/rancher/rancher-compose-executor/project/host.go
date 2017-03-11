package project

import (
	"github.com/rancher/go-rancher/v2"
	"golang.org/x/net/context"
)

type Hosts interface {
	Initialize(ctx context.Context) error
}

type HostsFactory interface {
	Create(projectName string, hostConfigs map[string]*client.Host) (Hosts, error)
}

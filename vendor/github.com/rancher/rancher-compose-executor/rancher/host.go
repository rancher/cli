package rancher

import (
	"golang.org/x/net/context"

	"github.com/rancher/go-rancher/v2"
	"github.com/rancher/rancher-compose-executor/project"
)

type RancherHostsFactory struct {
	Context *Context
}

func (f *RancherHostsFactory) Create(projectName string, hostConfigs map[string]*client.Host) (project.Hosts, error) {
	hosts := make([]*Host, 0, len(hostConfigs))
	for name, config := range hostConfigs {
		hosts = append(hosts, &Host{
			context:     f.Context,
			name:        name,
			projectName: projectName,
			hostConfig:  config,
		})
	}
	return &Hosts{
		hosts: hosts,
	}, nil
}

type Hosts struct {
	hosts   []*Host
	Context *Context
}

func (h *Hosts) Initialize(ctx context.Context) error {
	for _, host := range h.hosts {
		if err := host.EnsureItExists(ctx); err != nil {
			return err
		}
	}
	return nil
}

type Host struct {
	context     *Context
	name        string
	projectName string
	hostConfig  *client.Host
}

func (h *Host) EnsureItExists(ctx context.Context) error {
	host := *h.hostConfig
	// TODO: is this the proper hostname?
	host.Hostname = h.name
	// TODO: what if host already exists?
	_, err := h.context.Client.Host.Create(&host)
	return err
}

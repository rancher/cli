package cmd

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/go-rancher/client"
	"github.com/urfave/cli"
)

func PsCommand() cli.Command {
	return cli.Command{
		Name:   "ps",
		Usage:  "Show services/containers",
		Action: servicePs,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "containers,c",
				Usage: "Display containers",
			},
			cli.BoolFlag{
				Name:  "quiet,q",
				Usage: "Only display IDs",
			},
			cli.StringFlag{
				Name:  "format",
				Usage: "'json' or Custom format: {{.Id}} {{.Name}",
			},
		},
	}
}

func GetStackMap(c *client.RancherClient) map[string]client.Environment {
	result := map[string]client.Environment{}

	stacks, err := c.Environment.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"limit": -1,
		},
	})

	if err != nil {
		return result
	}

	for _, stack := range stacks.Data {
		result[stack.Id] = stack
	}

	return result
}

type PsData struct {
	Service       client.Service
	Stack         client.Environment
	CombinedState string
	ID            string
}

type ContainerPsData struct {
	ID            string
	Container     client.Container
	CombinedState string
	DockerID      string
}

func servicePs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	if ctx.Bool("containers") {
		return hostContainerPs(ctx, c)
	}

	if len(ctx.Args()) > 0 {
		return serviceContainersPs(ctx, c, ctx.Args())
	}

	stackMap := GetStackMap(c)

	collection, err := c.Service.List(nil)
	if err != nil {
		return errors.Wrap(err, "service list failed")
	}

	writer := NewTableWriter([][]string{
		{"ID", "Service.Id"},
		{"TYPE", "Service.Type"},
		{"NAME", "{{.Stack.Name}}/{{.Service.Name}}"},
		{"IMAGE", "Service.LaunchConfig.ImageUuid"},
		{"STATE", "CombinedState"},
		{"SCALE", "Service.Scale"},
		{"ENDPOINTS", "{{endpoint .Service.PublicEndpoints}}"},
		{"DETAIL", "Service.TransitioningMessage"},
	}, ctx)

	defer writer.Close()

	for _, item := range collection.Data {
		if item.LaunchConfig != nil {
			item.LaunchConfig.ImageUuid = strings.TrimPrefix(item.LaunchConfig.ImageUuid, "docker:")
		}

		combined := item.HealthState
		if item.State != "active" || combined == "" {
			combined = item.State
		}
		if item.LaunchConfig == nil {
			item.LaunchConfig = &client.LaunchConfig{}
		}
		writer.Write(PsData{
			ID:            item.Id,
			Service:       item,
			Stack:         stackMap[item.EnvironmentId],
			CombinedState: combined,
		})
	}

	return writer.Err()
}

func serviceContainersPs(ctx *cli.Context, c *client.RancherClient, names []string) error {
	containerList := []client.Container{}

	for _, name := range names {
		service, err := Lookup(c, name, "service")
		if err != nil {
			return err
		}

		var containers client.ContainerCollection
		if err := c.GetLink(*service, "instances", &containers); err != nil {
			return err
		}

		containerList = append(containerList, containers.Data...)
	}

	return containerPs(ctx, containerList)
}

func hostContainerPs(ctx *cli.Context, c *client.RancherClient) error {
	if len(ctx.Args()) == 0 {
		containerList, err := c.Container.List(nil)
		if err != nil {
			return err
		}
		return containerPs(ctx, containerList.Data)
	}

	containers := []client.Container{}
	for _, hostname := range ctx.Args() {
		host, err := Lookup(c, hostname, "host")
		if err != nil {
			return err
		}

		var containerList client.ContainerCollection
		if err := c.GetLink(*host, "instances", &containerList); err != nil {
			return err
		}

		containers = append(containers, containerList.Data...)
	}

	return containerPs(ctx, containers)
}

func containerPs(ctx *cli.Context, containers []client.Container) error {
	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "Container.Name"},
		{"IMAGE", "Container.ImageUuid"},
		{"STATE", "CombinedState"},
		{"HOST", "Container.HostId"},
		{"IP", "Container.PrimaryIpAddress"},
		{"DOCKER", "DockerID"},
		{"DETAIL", "Container.TransitioningMessage"},
		//TODO: {"PORTS", "{{ports .Container.Ports}}"},
	}, ctx)
	defer writer.Close()

	for _, container := range containers {
		container.ImageUuid = strings.TrimPrefix(container.ImageUuid, "docker:")
		combined := container.HealthState
		if container.State != "running" || combined == "" {
			combined = container.State
		}
		containerID := container.ExternalId
		if len(containerID) > 12 {
			containerID = containerID[:12]
		}
		writer.Write(ContainerPsData{
			Container:     container,
			ID:            container.Id,
			DockerID:      containerID,
			CombinedState: combined,
		})
	}

	return writer.Err()
}

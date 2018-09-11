package cmd

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/go-rancher/v2"
	"github.com/urfave/cli"
)

func PsCommand() cli.Command {
	return cli.Command{
		Name:        "ps",
		Usage:       "Show services/containers",
		Description: "\nLists all services or containers in the current $RANCHER_ENVIRONMENT. Use `--env <envID>` or `--env <envName>` to select a different environment.\n\nExample:\n\t$ rancher ps\n\t$ rancher ps -c\n\t$ rancher --env 1a5 ps\n",
		ArgsUsage:   "None",
		Action:      servicePs,
		Flags: []cli.Flag{
			listAllFlag(),
			listSystemFlag(),
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
				Usage: "'json' or Custom format: '{{.Service.Id}} {{.Service.Name}} {{.Service.LaunchConfig.ImageUuid}}'",
			},
		},
	}
}

func GetStackMap(c *client.RancherClient) map[string]client.Stack {
	result := map[string]client.Stack{}
	stacks, err := c.Stack.List(baseListOpts())
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
	Name          string
	LaunchConfig  interface{}
	Stack         client.Stack
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

	collection, err := c.Service.List(defaultListOpts(ctx))
	if err != nil {
		return errors.Wrap(err, "service list failed")
	}

	collectiondata := collection.Data

	for {
		collection, _ = collection.Next()
		if collection == nil {
			break
		}
		collectiondata = append(collectiondata, collection.Data...)
		if !collection.Pagination.Partial {
			break
		}
	}

	writer := NewTableWriter([][]string{
		{"ID", "Service.Id"},
		{"TYPE", "Service.Type"},
		{"NAME", "Name"},
		{"IMAGE", "LaunchConfig.ImageUuid"},
		{"STATE", "CombinedState"},
		{"SCALE", "{{len .Service.InstanceIds}}/{{.Service.Scale}}"},
		{"SYSTEM", "Service.System"},
		{"ENDPOINTS", "{{endpoint .Service.PublicEndpoints}}"},
		{"DETAIL", "Service.TransitioningMessage"},
	}, ctx)

	defer writer.Close()

	for _, item := range collectiondata {
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
			Name:          fmt.Sprintf("%s/%s", stackMap[item.StackId].Name, item.Name),
			LaunchConfig:  *item.LaunchConfig,
			Stack:         stackMap[item.StackId],
			CombinedState: combined,
		})
		for _, sidekick := range item.SecondaryLaunchConfigs {
			sidekick.ImageUuid = strings.TrimPrefix(sidekick.ImageUuid, "docker:")
			item.Type = "sidekick"
			writer.Write(PsData{
				ID:      item.Id,
				Service: item,
				Name: fmt.Sprintf("%s/%s/%s", stackMap[item.StackId].Name, item.Name,
					sidekick.Name),
				LaunchConfig:  sidekick,
				Stack:         stackMap[item.StackId],
				CombinedState: combined,
			})
		}
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
		containerList, err := c.Container.List(defaultListOpts(ctx))
		if err != nil {
			return err
		}
		collectiondata := containerList.Data

		for {
			containerList, _ = containerList.Next()
			if containerList == nil {
				break
			}
			collectiondata = append(collectiondata, containerList.Data...)
			if !containerList.Pagination.Partial {
				break
			}
		}

		return containerPs(ctx, collectiondata)
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

package cmd

import (
	"strings"

	"github.com/codegangsta/cli"
	"github.com/rancher/go-rancher/client"
)

func PsCommand() cli.Command {
	return cli.Command{
		Name:   "ps",
		Usage:  "Show services/containers",
		Action: servicePs,
		Flags: []cli.Flag{
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

func servicePs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	env, err := GetEnvironment(c)
	if err != nil {
		return err
	}

	stackMap := GetStackMap(c)

	collection, err := c.Service.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"accountId": env.Id,
		},
	})
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "Service.Id"},
		{"TYPE", "Service.Type"},
		{"NAME", "{{.Stack.Name}}/{{.Service.Name}}"},
		{"IMAGE", "Service.LaunchConfig.ImageUuid"},
		{"STATE", "CombinedState"},
		{"SCALE", "Service.Scale"},
		{"ENDPOINTS", "{{endpoint .Service.PublicEndpoints}}"},
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
		writer.Write(PsData{
			ID:            item.Id,
			Service:       item,
			Stack:         stackMap[item.EnvironmentId],
			CombinedState: combined,
		})
	}

	return writer.Err()
}

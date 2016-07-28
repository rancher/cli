package cmd

import (
	"strings"

	"github.com/rancher/go-rancher/client"
	"github.com/urfave/cli"
)

func StackCommand() cli.Command {
	return cli.Command{
		Name:      "stacks",
		ShortName: "stack",
		Usage:     "Operations on stacks",
		Action:    stackLs,
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

type StackData struct {
	ID      string
	Catalog string
	Stack   client.Environment
	State   string
	System  bool
}

func stackLs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	collection, err := c.Environment.List(nil)
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "Stack.Name"},
		{"STATE", "State"},
		{"CATALOG", "Catalog"},
		{"SYSTEM", "System"},
		{"DETAIL", "Stack.TransitioningMessage"},
	}, ctx)

	defer writer.Close()

	for _, item := range collection.Data {
		system := strings.HasPrefix(item.ExternalId, "system://")
		if !system {
			system = strings.HasPrefix(item.ExternalId, "system-catalog://")
		}
		if !system {
			system = strings.HasPrefix(item.ExternalId, "kubernetes")
		}
		combined := item.HealthState
		if item.State != "active" || combined == "" {
			combined = item.State
		}
		writer.Write(&StackData{
			ID:      item.Id,
			Stack:   item,
			State:   combined,
			System:  system,
			Catalog: item.ExternalId,
		})
	}

	return writer.Err()
}

package cmd

import (
	"strings"

	"github.com/rancher/go-rancher/v2"
	"github.com/urfave/cli"
)

func StackCommand() cli.Command {
	stackLsFlags := []cli.Flag{
		cli.BoolFlag{
			Name:  "quiet,q",
			Usage: "Only display IDs",
		},
		cli.StringFlag{
			Name:  "format",
			Usage: "'json' or Custom format: {{.Id}} {{.Name}}",
		},
	}

	return cli.Command{
		Name:      "stacks",
		ShortName: "stack",
		Usage:     "Operations on stacks",
		Action:    defaultAction(stackLs),
		Flags:     stackLsFlags,
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "ls",
				Usage:       "List stacks",
				Description: "\nLists all stacks in the current $RANCHER_ENVIRONMENT. Use `--env <envID>` or `--env <envName>` to select a different environment.\n\nExample:\n\t$ rancher stacks ls\n\t$ rancher --env 1a5 stacks ls\n",
				ArgsUsage:   "None",
				Action:      stackLs,
				Flags:       stackLsFlags,
			},
		},
	}
}

type StackData struct {
	ID      string
	Catalog string
	Stack   client.Stack
	State   string
	System  bool
}

func stackLs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	collection, err := c.Stack.List(defaultListOpts(nil))
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

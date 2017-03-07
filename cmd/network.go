package cmd

import (
	"github.com/rancher/go-rancher/v2"
	"github.com/urfave/cli"
)

func NetworkCommand() cli.Command {
	networkLsFlags := []cli.Flag{
		listAllFlag(),
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
		Name:      "networks",
		ShortName: "network",
		Usage:     "Operations on networks",
		Action:    defaultAction(networkLs),
		Flags:     networkLsFlags,
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "ls",
				Usage:       "List networks",
				Description: "\nLists all networks in the current $RANCHER_ENVIRONMENT. Use `--env <envID>` or `--env <envName>` to select a different environment.\n\nExample:\n\t$ rancher networks ls\n\t$ rancher --env 1a5 networks ls\n",
				ArgsUsage:   "None",
				Action:      networkLs,
				Flags:       networkLsFlags,
			},
		},
	}
}

type NetworksData struct {
	ID                  string
	Network             client.Network
	State               string
	DefaultPolicyAction string
}

func getNetworkState(network *client.Network) string {
	state := network.State
	if state == "active" && network.State != "" {
		state = network.State
	}
	return state
}

func networkLs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	collection, err := c.Network.List(defaultListOpts(ctx))
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "Network.Id"},
		{"NAME", "Network.Name"},
		{"STATE", "State"},
		{"DEFAULTPOLICYACTION", "Network.DefaultPolicyAction"},
	}, ctx)

	defer writer.Close()

	for _, item := range collection.Data {
		writer.Write(&NetworksData{
			ID:                  item.Id,
			Network:             item,
			State:               getNetworkState(&item),
			DefaultPolicyAction: item.DefaultPolicyAction,
		})
	}

	return writer.Err()
}

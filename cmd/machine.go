package cmd

import (
	"context"
	"fmt"

	"github.com/rancher/cli/cliclient"
	capiClient "github.com/rancher/rancher/pkg/client/generated/cluster/v1beta2"
	"github.com/urfave/cli/v3"
)

type MachineData struct {
	ID      string
	Machine capiClient.Machine
	Name    string
}

func MachineCommand() *cli.Command {
	return &cli.Command{
		Name:    "machines",
		Aliases: []string{"machine"},
		Usage:   "Operations on machines",
		Action:  defaultAction(machineLs),
		Commands: []*cli.Command{
			{
				Name:        "ls",
				Usage:       "List machines",
				Description: "\nLists all machines in the current cluster.",
				ArgsUsage:   "None",
				Action:      machineLs,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "format",
						Usage: "'json', 'yaml' or Custom format: '{{.Machine.ID}} {{.Machine.Name}}'",
					},
					quietFlag,
				},
			},
		},
	}
}

func machineLs(ctx context.Context, cmd *cli.Command) error {
	c, err := GetClient(cmd)
	if err != nil {
		return err
	}

	collection, err := getMachinesList(cmd, c)
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "Name"},
		{"PHASE", "Machine.Status.Phase"},
	}, cmd)

	defer writer.Close()

	for _, item := range collection.Data {
		writer.Write(&MachineData{
			ID:      item.ID,
			Machine: item,
			Name:    getMachineName(item),
		})
	}

	return writer.Err()
}

func getMachinesList(
	cmd *cli.Command,
	c *cliclient.MasterClient,
) (*capiClient.MachineCollection, error) {
	filter := defaultListOpts(cmd)
	return c.CAPIClient.Machine.List(filter)
}

func getMachineByNodeName(
	cmd *cli.Command,
	c *cliclient.MasterClient,
	nodeName string,
) (capiClient.Machine, error) {
	machineCollection, err := getMachinesList(cmd, c)
	if err != nil {
		return capiClient.Machine{}, err
	}

	for _, machine := range machineCollection.Data {
		if machine.Status.NodeRef != nil && machine.Status.NodeRef.Name == nodeName {
			return machine, nil
		}
	}

	return capiClient.Machine{}, fmt.Errorf("no machine found associated with node [%s], run "+
		"`rancher machines` to see available nodes", nodeName)
}

func getMachineName(machine capiClient.Machine) string {
	if machine.Name != "" {
		return machine.Name
	} else if machine.Status.NodeRef != nil {
		return machine.Status.NodeRef.Name
	} else if machine.InfrastructureRef != nil {
		return machine.InfrastructureRef.Name
	}
	return machine.ID
}

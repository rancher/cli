package cmd

import (
	"strings"

	"github.com/rancher/go-rancher/v3"
	"github.com/urfave/cli"
)

var (
	startTypes = []string{"service", "container", "host", "stack"}
)

func StartCommand() cli.Command {
	return cli.Command{
		Name:        "start",
		ShortName:   "activate",
		Usage:       "Start or activate " + strings.Join(startTypes, ", "),
		Description: "\nStart resources by ID or name in the current $RANCHER_ENVIRONMENT.  Use `--env <envID>` or `--env <envName>` to select a different environment.\n\nExample:\n\t$ rancher start 1s70\n\t$ rancher --env 1a5 start stackName/serviceName \n",
		ArgsUsage:   "[ID NAME...]",
		Action:      startResources,
		Flags: []cli.Flag{
			typesStringFlag(startTypes),
		},
	}
}

func startResources(ctx *cli.Context) error {
	return forEachResource(ctx, startTypes, func(c *client.RancherClient, resource *client.Resource) (string, error) {
		action, err := pickAction(resource, "start", "activate", "activateservices")
		if err != nil {
			return "", err
		}
		err = c.Action(resource.Type, action, resource, nil, resource)
		return resource.Id, err
	})
}

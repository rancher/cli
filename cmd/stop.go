package cmd

import (
	"strings"

	"github.com/rancher/go-rancher/v2"
	"github.com/urfave/cli"
)

var (
	stopTypes = []string{"service", "container", "host", "stack"}
)

func StopCommand() cli.Command {
	return cli.Command{
		Name:        "stop",
		ShortName:   "deactivate",
		Usage:       "Stop or deactivate " + strings.Join(stopTypes, ", "),
		Description: "\nStop resources by ID or name in the current $RANCHER_ENVIRONMENT.  Use `--env <envID>` or `--env <envName>` to select a different environment.\n\nExample:\n\t$ rancher stop 1s70\n\t$ rancher --env 1a5 stop stackName/serviceName \n",
		ArgsUsage:   "[ID NAME...]",
		Action:      stopResources,
		Flags: []cli.Flag{
			typesStringFlag(stopTypes),
		},
	}
}

func stopResources(ctx *cli.Context) error {
	return forEachResource(ctx, stopTypes, func(c *client.RancherClient, resource *client.Resource) (string, error) {
		action, err := pickAction(resource, "stop", "deactivate", "deactivateservices")
		if err != nil {
			return "", err
		}
		return resource.Id, c.Action(resource.Type, action, resource, nil, resource)
	})
}

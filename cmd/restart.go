package cmd

import (
	"strings"

	"github.com/rancher/go-rancher/v3"
	"github.com/urfave/cli"
)

var (
	restartTypes = []string{"service", "container"}
)

func RestartCommand() cli.Command {
	return cli.Command{
		Name:        "restart",
		Usage:       "Restart " + strings.Join(restartTypes, ", "),
		Description: "\nRestart resources by ID or name in the current $RANCHER_ENVIRONMENT.  Use `--env <envID>` or `--env <envName>` to select a different environment.\n\nExample:\n\t$ rancher restart 1s70\n\t$ rancher --env 1a5 restart stackName/serviceName \n",
		ArgsUsage:   "[ID NAME...]",
		Action:      restartResources,
		Flags: []cli.Flag{
			typesStringFlag(restartTypes),
			cli.IntFlag{
				Name:  "batch-size",
				Usage: "Number of containers to restart at a time",
				Value: 1,
			},
			cli.IntFlag{
				Name:  "interval",
				Usage: "Interval in millisecond to wait between restarts",
				Value: 1000,
			},
		},
	}
}

func restartResources(ctx *cli.Context) error {
	return forEachResource(ctx, restartTypes, func(c *client.RancherClient, resource *client.Resource) (string, error) {
		action, err := pickAction(resource, "restart")
		if err != nil {
			return "", err
		}
		//todo: revisit restart policy
		err = c.Action(resource.Type, action, resource, nil, resource)
		return resource.Id, err
	})
}

package cmd

import (
	"fmt"
	"strings"

	"github.com/rancher/go-rancher/client"
	"github.com/urfave/cli"
)

var (
	restartTypes = cli.StringSlice([]string{"service", "container"})
)

func RestartCommand() cli.Command {
	return cli.Command{
		Name:   "restart",
		Usage:  "Restart " + strings.Join(restartTypes, ", "),
		Action: restartResources,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "type",
				Usage: "Restrict restart to specific types",
				Value: &restartTypes,
			},
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
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	types := ctx.StringSlice("type")

	var lastErr error
	for _, id := range ctx.Args() {
		resource, err := Lookup(c, id, types...)
		if err != nil {
			lastErr = err
			fmt.Println(lastErr)
			continue
		}

		if err := c.Action(resource.Type, "restart", resource, &client.ServiceRestart{
			RollingRestartStrategy: client.RollingRestartStrategy{
				BatchSize:      int64(ctx.Int("batch-size")),
				IntervalMillis: int64(ctx.Int("interval")),
			},
		}, resource); err != nil {
			lastErr = err
			fmt.Println(lastErr)
		} else {
			fmt.Println(resource.Id)
		}
	}

	return lastErr
}

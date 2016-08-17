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
		Name:        "restart",
		Usage:       "Restart " + strings.Join(restartTypes, ", "),
		Description: "\nRestart resources by ID or name in the current $RANCHER_ENVIRONMENT.  Use `--env <envID>` or `--env <envName>` to select a different environment.\n\nExample:\n\t$ rancher restart 1s70\n\t$ rancher --env 1a5 restart stackName/serviceName \n",
		ArgsUsage:   "[ID NAME...]",
		Action:      restartResources,
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

	w, err := NewWaiter(ctx)
	if err != nil {
		return err
	}

	types := ctx.StringSlice("type")

	var lastErr error
	var envErr error
	for _, id := range ctx.Args() {
		resource, err := Lookup(c, id, types...)
		if err != nil {
			lastErr = err
			if _, envErr = LookupEnvironment(c, id); envErr != nil {
				fmt.Println("Incorrect usage: Environments cannot be restarted.")
			} else {
				fmt.Println(lastErr)
			}
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
			w.Add(resource.Id)
			//fmt.Println(resource.Id)
		}

		if lastErr != nil && envErr == nil {
			return lastErr
		}
	}

	return w.Wait()
}

package cmd

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"
)

var (
	stopTypes = cli.StringSlice([]string{"service", "container", "host"})
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
			cli.StringSliceFlag{
				Name:  "type",
				Usage: "Restrict stop to specific types",
				Value: &stopTypes,
			},
		},
	}
}

func stopResources(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	types := ctx.StringSlice("type")

	w, err := NewWaiter(ctx)
	if err != nil {
		return err
	}

	var lastErr error
	var envErr error
	for _, id := range ctx.Args() {
		resource, err := Lookup(c, id, types...)
		if err != nil {
			lastErr = err
			if _, envErr = LookupEnvironment(c, id); envErr != nil {
				fmt.Println("Incorrect usage: Use `rancher env stop`.")
			} else {
				fmt.Println(lastErr)
			}
			continue
		}

		action := ""
		if _, ok := resource.Actions["stop"]; ok {
			action = "stop"
		} else if _, ok := resource.Actions["deactivate"]; ok {
			action = "deactivate"
		}

		if action == "" {
			lastErr = fmt.Errorf("stop or deactivate not available on %s", id)
			fmt.Println(lastErr)
		} else if err := c.Action(resource.Type, action, resource, nil, resource); err != nil {
			lastErr = err
			fmt.Println(lastErr)
		} else {
			w.Add(resource.Id)
		}
	}

	if lastErr != nil && envErr == nil {
		return lastErr
	}

	return w.Wait()
}

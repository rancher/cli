package cmd

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"
)

var (
	startTypes = cli.StringSlice([]string{"service", "container", "host"})
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
			cli.StringSliceFlag{
				Name:  "type",
				Usage: "Restrict start to specific types",
				Value: &startTypes,
			},
		},
	}
}

func startResources(ctx *cli.Context) error {
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
	for _, id := range ctx.Args() {
		resource, err := Lookup(c, id, types...)
		if err != nil {
			lastErr = err
			fmt.Println(lastErr)
			continue
		}

		action := "activate"
		if _, ok := resource.Actions["start"]; ok {
			action = "start"
		}

		if err := c.Action(resource.Type, action, resource, nil, resource); err != nil {
			lastErr = err
			fmt.Println(lastErr)
		} else {
			w.Add(resource.Id)
		}
	}

	if lastErr != nil {
		return lastErr
	}

	return w.Wait()
}

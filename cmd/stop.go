package cmd

import (
	"fmt"
	"strings"

	"github.com/codegangsta/cli"
)

var (
	stopTypes = cli.StringSlice([]string{"service", "container", "host"})
)

func StopCommand() cli.Command {
	return cli.Command{
		Name:      "stop",
		ShortName: "deactivate",
		Usage:     "Stop or deactivate " + strings.Join(stopTypes, ", "),
		Action:    stopResources,
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

	var lastErr error
	for _, id := range ctx.Args() {
		resource, err := Lookup(c, id, types...)
		if err != nil {
			lastErr = err
			fmt.Println(lastErr)
			continue
		}

		action := "deactivate"
		if _, ok := resource.Actions["stop"]; ok {
			action = "stop"
		}

		if err := c.Action(resource.Type, action, resource, nil, resource); err != nil {
			lastErr = err
			fmt.Println(lastErr)
		} else {
			fmt.Println(resource.Id)
		}
	}

	return lastErr
}

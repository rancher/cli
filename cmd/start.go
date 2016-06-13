package cmd

import (
	"fmt"
	"strings"

	"github.com/codegangsta/cli"
)

var (
	startTypes = cli.StringSlice([]string{"service", "container", "host"})
)

func StartCommand() cli.Command {
	return cli.Command{
		Name:      "start",
		ShortName: "activate",
		Usage:     "Start or activate " + strings.Join(startTypes, ", "),
		Action:    startResources,
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
			fmt.Println(resource.Id)
		}
	}

	return lastErr
}

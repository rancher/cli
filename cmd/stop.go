package cmd

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"
)

var (
	stopTypes = cli.StringSlice([]string{"service", "container", "host", "account"})
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

	if lastErr != nil {
		return lastErr
	}

	return w.Wait()
}

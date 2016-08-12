package cmd

import (
	"fmt"

	"github.com/urfave/cli"
)

var (
	rmTypes = []string{"service", "container", "host", "environment", "machine"}
)

func RmCommand() cli.Command {
	return cli.Command{
		Name:        "rm",
		Usage:       "Delete resources",
		Description: "\nDeletes resources by ID or name in the current $RANCHER_ENVIRONMENT.  Use `--env <envID>` or `--env <envName>` to select a different environment.\n\nExample:\n\t$ rancher rm 1s70\n\t$ rancher --env 1a5 rm stackName/serviceName \n",
		ArgsUsage:   "[ID NAME...]",
		Action:      deleteResources,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "type",
				Usage: "Restrict delete to specific types",
				Value: &cli.StringSlice{},
			},
		},
	}
}

func deleteResources(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	w, err := NewWaiter(ctx)
	if err != nil {
		return err
	}

	types := ctx.StringSlice("type")
	if len(types) == 0 {
		types = rmTypes
	}

	var lastErr error
	for _, id := range ctx.Args() {
		resource, err := Lookup(c, id, types...)
		if err != nil {
			lastErr = err
			fmt.Println(lastErr)
			continue
		}

		if err := c.Delete(resource); err != nil {
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

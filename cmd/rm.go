package cmd

import (
	"fmt"

	"github.com/codegangsta/cli"
)

var (
	rmTypes = cli.StringSlice([]string{"service", "container", "host", "environment"})
)

func RmCommand() cli.Command {
	return cli.Command{
		Name:   "rm",
		Usage:  "Delete resources",
		Action: deleteResources,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "type",
				Usage: "Restrict delete to specific types",
				Value: &rmTypes,
			},
		},
	}
}

func deleteResources(ctx *cli.Context) error {
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

		if err := c.Delete(resource); err != nil {
			lastErr = err
			fmt.Println(lastErr)
		} else {
			fmt.Println(resource.Id)
		}
	}

	return lastErr
}

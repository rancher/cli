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
		Name:   "rm",
		Usage:  "Delete resources",
		Action: deleteResources,
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

package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/go-rancher/v3"
	"github.com/urfave/cli"
)

func ScaleCommand() cli.Command {
	return cli.Command{
		Name:        "scale",
		Usage:       "Set number of containers to run for a service",
		Action:      serviceScale,
		Description: "\nNumbers are specified in the form `service=num` as arguments.\n\nExample:\n\t$ rancher scale web=2 worker=3\n",
		ArgsUsage:   "[SERVICE=NUM...]",
	}
}

type scaleUp struct {
	name     string
	resource *client.Resource
	scale    int
}

func serviceScale(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	servicesToScale := []scaleUp{}
	for _, arg := range ctx.Args() {
		scale := 1
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) > 1 {
			i, err := strconv.Atoi(parts[1])
			if err != nil {
				return fmt.Errorf("Invalid format for %s, expecting name=scale, example: web=2", arg)
			}
			scale = i
		}

		resource, err := Lookup(c, parts[0], "service", "container")
		if err != nil {
			return err
		}

		servicesToScale = append(servicesToScale, scaleUp{
			name:     parts[0],
			resource: resource,
			scale:    scale,
		})
	}

	w, err := NewWaiter(ctx)
	if err != nil {
		return err
	}
	for _, toScale := range servicesToScale {
		w.Add(toScale.name)

		if toScale.resource.Type == "service" {
			err := c.Update("service", toScale.resource, map[string]interface{}{
				"scale": toScale.scale,
			}, toScale.resource)
			if err != nil {
				return err
			}
		} else if toScale.resource.Type == "container" {
			// convert it into service
			container, err := c.Container.ById(toScale.resource.Id)
			if err != nil {
				return err
			}
			service, err := c.Container.ActionConverttoservice(container)
			if err != nil {
				return err
			}
			err = c.Update("service", &service.Resource, map[string]interface{}{
				"scale": toScale.scale,
			}, &service.Resource)
			if err != nil {
				return err
			}
		}
	}

	return w.Wait()
}

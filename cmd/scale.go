package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/rancher/go-rancher/client"
)

func ScaleCommand() cli.Command {
	return cli.Command{
		Name:   "scale",
		Usage:  "Scale a service",
		Action: serviceScale,
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

		resource, err := Lookup(c, parts[0], "service")
		if err != nil {
			return err
		}

		servicesToScale = append(servicesToScale, scaleUp{
			name:     parts[0],
			resource: resource,
			scale:    scale,
		})
	}

	for _, toScale := range servicesToScale {
		fmt.Printf("%s=%d\n", toScale.name, toScale.scale)

		err := c.Update("service", toScale.resource, map[string]interface{}{
			"scale": toScale.scale,
		}, toScale.resource)
		if err != nil {
			return err
		}
	}

	return nil
}

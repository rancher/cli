package cmd

import (
	"strings"

	"github.com/rancher/go-rancher/v2"
	"github.com/urfave/cli"
)

var (
	inspectTypes = []string{"service", "container", "host", "project", "stack", "volume"}
)

func InspectCommand() cli.Command {
	return cli.Command{
		Name:  "inspect",
		Usage: "View details for " + replaceTypeNames(strings.Join(inspectTypes, ", ")),
		Description: `
Inspect resources by ID or name in the current $RANCHER_ENVIRONMENT.  Use '--env <envID>' or '--env <envName>' to select a different environment.

Example:
	$ rancher inspect 1s70
`,
		ArgsUsage: "[ID NAME...]",
		Action:    inspectResources,
		Flags: []cli.Flag{
			typesStringFlag(stopTypes),
			cli.BoolFlag{
				Name:  "links",
				Usage: "Include URLs to actions and links in resource output",
			},
			cli.StringFlag{
				Name:  "format",
				Usage: "'json' or Custom format: {{.Id}} {{.Name}}",
				Value: "json",
			},
		},
	}
}

func inspectResources(ctx *cli.Context) error {
	writer := NewTableWriter(nil, ctx)
	forEachResource(ctx, inspectTypes, func(c *client.RancherClient, resource *client.Resource) (string, error) {
		mapResource := map[string]interface{}{}
		err := c.ById(resource.Type, resource.Id, &mapResource)
		if err != nil {
			return "-", err
		}
		if !ctx.Bool("links") {
			delete(mapResource, "links")
			delete(mapResource, "actions")
		}
		writer.Write(mapResource)
		return "-", nil
	})
	return writer.Err()
}

package cmd

import (
	"strings"

	"github.com/urfave/cli"
)

func InspectCommand() cli.Command {
	return cli.Command{
		Name:  "inspect",
		Usage: "View details of resources",
		Description: `
Inspect resources by name or ID in the current context. If the 'type' is not specified inspect will search: ` + strings.Join(listAllRoles(), ", ") + `
Examples:
	# Specify the type
	$ rancher inspect --type cluster clusterFoo

	# No type is specified so defaults are checked
	$ rancher inspect myvolume

	# Inspect a project and get the output in yaml format with the projects links
	$ rancher inspect --type project --format yaml --links projectFoo
`,
		ArgsUsage: "[RESOURCEID RESOURCENAME]",
		Action:    inspectResources,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "links",
				Usage: "Include URLs to actions and links in resource output",
			},
			cli.StringFlag{
				Name:  "type",
				Usage: "Specify the type of resource to inspect",
			},
			cli.StringFlag{
				Name:  "format",
				Usage: "'json', 'yaml' or Custom format: '{{.kind}}'",
				Value: "json",
			},
		},
	}
}

func inspectResources(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowCommandHelp(ctx, "inspect")
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	t := ctx.String("type")
	types := []string{}
	if t != "" {
		rt, err := GetResourceType(c, t)
		if err != nil {
			return err
		}
		types = append(types, rt)
	} else {
		types = listAllRoles()
	}

	resource, err := Lookup(c, ctx.Args().First(), types...)
	if nil != err {
		return err
	}
	mapResource := map[string]interface{}{}

	err = c.ByID(resource, &mapResource)
	if err != nil {
		return err
	}

	if !ctx.Bool("links") {
		delete(mapResource, "links")
		delete(mapResource, "actions")
	}
	writer := NewTableWriter(nil, ctx)
	writer.Write(mapResource)
	writer.Close()

	return writer.Err()
}

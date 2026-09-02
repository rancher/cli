package cmd

import (
	"context"
	"strings"

	"github.com/urfave/cli/v3"
)

func InspectCommand() *cli.Command {
	return &cli.Command{
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
			&cli.BoolFlag{
				Name:  "links",
				Usage: "Include URLs to actions and links in resource output",
			},
			&cli.StringFlag{
				Name:  "type",
				Usage: "Specify the type of resource to inspect",
			},
			&cli.StringFlag{
				Name:  "format",
				Usage: "'json', 'yaml' or Custom format: '{{.kind}}'",
				Value: "json",
			},
		},
	}
}

func inspectResources(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() == 0 {
		return cli.ShowCommandHelp(ctx, cmd, "inspect")
	}

	c, err := GetClient(cmd)
	if err != nil {
		return err
	}

	t := cmd.String("type")
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

	resource, err := Lookup(c, cmd.Args().First(), types...)
	if err != nil {
		return err
	}
	mapResource := map[string]interface{}{}

	err = c.ByID(resource, &mapResource)
	if err != nil {
		return err
	}

	if !cmd.Bool("links") {
		delete(mapResource, "links")
		delete(mapResource, "actions")
	}
	writer := NewTableWriter(nil, cmd)
	writer.Write(mapResource)
	writer.Close()

	return writer.Err()
}

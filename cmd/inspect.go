package cmd

import (
	"strings"

	"github.com/pkg/errors"
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
		types = append(types, t)
	} else {
		types = listAllRoles()
	}

	resource, err := Lookup(c, ctx.Args().First(), types...)
	if nil != err {
		return err
	}
	mapResource := map[string]interface{}{}

	if _, ok := c.ManagementClient.APIBaseClient.Types[resource.Type]; ok {
		err = c.ManagementClient.ByID(resource.Type, resource.ID, &mapResource)
		if err != nil {
			return err
		}
	} else if _, ok := c.ProjectClient.APIBaseClient.Types[resource.Type]; ok {
		err = c.ProjectClient.ByID(resource.Type, resource.ID, &mapResource)
		if err != nil {
			return err
		}
	} else if _, ok := c.ClusterClient.APIBaseClient.Types[resource.Type]; ok {
		err = c.ClusterClient.ByID(resource.Type, resource.ID, &mapResource)
		if err != nil {
			return err
		}
	} else {
		return errors.New("unkown resource type")
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

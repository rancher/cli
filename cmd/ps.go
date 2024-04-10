package cmd

import (
	"strconv"

	"github.com/rancher/cli/cliclient"
	"github.com/urfave/cli"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type PSHolder struct {
	NameSpace string
	Name      string
	Type      string
	State     string
	Image     string
	Scale     string
}

func PsCommand() cli.Command {
	return cli.Command{
		Name:  "ps",
		Usage: "Show workloads in a project",
		Description: `Show information on the workloads in a project. Defaults to the current context.
Examples:
	# Show workloads in the current context
	$ rancher ps

	# Show workloads in a specific project and output the results in yaml
	$ rancher ps --project projectFoo --format yaml
`,
		Action: psLs,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "project",
				Usage: "Optional project to show workloads for",
			},
			cli.StringFlag{
				Name:  "format",
				Usage: "'json', 'yaml' or Custom format: '{{.Name}} {{.Image}}'",
			},
		},
	}
}

func psLs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	if ctx.String("project") != "" {
		//Verify the project given is valid
		resource, err := Lookup(c, ctx.String("project"), "project")
		if err != nil {
			return err
		}

		sc, err := lookupConfig(ctx)
		if err != nil {
			return err
		}
		sc.Project = resource.ID

		projClient, err := cliclient.NewProjectClient(sc)
		if err != nil {
			return err
		}
		c.ProjectClient = projClient.ProjectClient
	}

	workLoads, err := c.ProjectClient.Workload.List(defaultListOpts(ctx))
	if err != nil {
		return err
	}

	wlWriter := NewTableWriter([][]string{
		{"NAMESPACE", "NameSpace"},
		{"NAME", "Name"},
		{"TYPE", "Type"},
		{"STATE", "State"},
		{"IMAGE", "Image"},
		{"SCALE", "Scale"},
	}, ctx)

	defer wlWriter.Close()

	titleCaser := cases.Title(language.Und)

	for _, item := range workLoads.Data {
		var scale string

		if item.Scale == nil {
			scale = "-"
		} else {
			scale = strconv.Itoa(int(*item.Scale))
		}

		item.Type = titleCaser.String(item.Type)

		wlWriter.Write(&PSHolder{
			NameSpace: item.NamespaceId,
			Name:      item.Name,
			Type:      item.Type,
			State:     item.State,
			Image:     item.Containers[0].Image,
			Scale:     scale,
		})
	}

	opts := defaultListOpts(ctx)
	opts.Filters["workloadId"] = ""

	orphanPods, err := c.ProjectClient.Pod.List(opts)
	if err != nil {
		return err
	}

	if len(orphanPods.Data) > 0 {
		for _, item := range orphanPods.Data {
			item.Type = titleCaser.String(item.Type)
			wlWriter.Write(&PSHolder{
				NameSpace: item.NamespaceId,
				Name:      item.Name,
				Type:      item.Type,
				State:     item.State,
				Image:     item.Containers[0].Image,
				Scale:     "Standalone", // a single pod doesn't have scale
			})
		}
	}

	return nil
}

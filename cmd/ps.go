package cmd

import (
	"strconv"
	"strings"

	"github.com/rancher/cli/cliclient"
	"github.com/urfave/cli"
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
		Name:        "ps",
		Usage:       "Show workloads and pods",
		Description: "Prints out a table of pods not associated with a workload then a table of workloads",
		Action:      psLs,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "project",
				Usage: "project to show workloads for",
			},
			cli.StringFlag{
				Name:  "format",
				Usage: "'json' or Custom format: '{{.Name}} {{.Image}}'",
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
		_, err := getProjectByID(c, ctx.String("project"))
		if nil != err {
			return err
		}

		sc, err := lookupConfig(ctx)
		if nil != err {
			return err
		}
		sc.Project = ctx.String("project")

		projClient, err := cliclient.NewProjectClient(sc)
		if nil != err {
			return err
		}
		c.ProjectClient = projClient.ProjectClient
	}

	workLoads, err := c.ProjectClient.Workload.List(defaultListOpts(ctx))
	if nil != err {
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

	for _, item := range workLoads.Data {
		var scale string

		if item.Scale == nil {
			scale = "-"
		} else {
			scale = strconv.Itoa(int(*item.Scale))
		}

		item.Type = strings.Title(item.Type)

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
	if nil != err {
		return err
	}

	if len(orphanPods.Data) > 0 {
		for _, item := range orphanPods.Data {
			item.Type = strings.Title(item.Type)
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

package cmd

import (
	"fmt"

	"github.com/rancher/cli/cliclient"
	projectClient "github.com/rancher/types/client/project/v3"
	"github.com/urfave/cli"
)

type WorkLoadPS struct {
	WorkLoad projectClient.Workload
	Name     string // this is built from namespace/name
}

type PodPS struct {
	Pod  projectClient.Pod
	Name string // this is built from namespace/name
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
		{"NAME", "Name"},
		{"STATE", "WorkLoad.State"},
		{"SCALE", "WorkLoad.Scale"},
		{"DETAIL", "WorkLoad.TransitioningMessage"},
	}, ctx)

	defer wlWriter.Close()

	for _, item := range workLoads.Data {
		wlWriter.Write(&WorkLoadPS{
			WorkLoad: item,
			Name:     fmt.Sprintf("%s/%s", item.NamespaceId, item.Name),
		})
	}

	// Add an empty line to the stack to separate the tables
	defer fmt.Println("")

	opts := defaultListOpts(ctx)
	opts.Filters["workloadId"] = ""

	orphanPods, err := c.ProjectClient.Pod.List(opts)
	if nil != err {
		return err
	}

	podWriter := NewTableWriter([][]string{
		{"NAME", "Name"},
		{"STATE", "Pod.State"},
		{"DETAIL", "Pod.TransitioningMessage"},
	}, ctx)

	defer podWriter.Close()

	for _, item := range orphanPods.Data {
		podWriter.Write(&PodPS{
			Pod:  item,
			Name: fmt.Sprintf("%s/%s", item.NamespaceId, item.Name),
		})
	}

	return nil
}

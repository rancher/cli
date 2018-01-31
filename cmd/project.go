package cmd

import (
	"errors"
	"strings"

	managementClient "github.com/rancher/types/client/management/v3"
	"github.com/urfave/cli"
)

type ProjectData struct {
	Project managementClient.Project
}

func ProjectCommand() cli.Command {
	return cli.Command{
		Name:    "projects",
		Aliases: []string{"project"},
		Usage:   "Operations on projects",
		Action:  defaultAction(projectLs),
		Subcommands: []cli.Command{
			{
				Name:        "ls",
				Usage:       "List projects",
				Description: "\nLists all projects in the current cluster.",
				ArgsUsage:   "None",
				Action:      projectLs,
			},
			{
				Name:        "create",
				Usage:       "Create a project",
				Description: "\nCreates a project in the current cluster.",
				ArgsUsage:   "[NEWPROJECTNAME...]",
				Action:      projectCreate,
			},
		},
	}
}

func projectLs(ctx *cli.Context) error {
	collection, err := getProjectList(ctx)
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "Project.ID"},
		{"NAME", "Project.Name"},
		{"STATE", "Project.State"},
	}, ctx)

	defer writer.Close()

	for _, item := range collection.Data {
		writer.Write(&ProjectData{
			Project: item,
		})
	}

	return writer.Err()
}

func projectCreate(ctx *cli.Context) error {
	config, err := lookupConfig(ctx)
	if nil != err {
		return err
	}
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	if ctx.NArg() == 0 {
		return errors.New("name is required")
	}

	name := ctx.Args().First()
	newProj := &managementClient.Project{
		Name:      name,
		ClusterId: strings.Split(config.Project, ":")[0],
	}

	c.ManagementClient.Project.Create(newProj)
	return nil
}

func getProjectList(ctx *cli.Context) (*managementClient.ProjectCollection, error) {
	c, err := GetClient(ctx)
	if err != nil {
		return nil, err
	}

	collection, err := c.ManagementClient.Project.List(defaultListOpts(ctx))
	if err != nil {
		return nil, err
	}
	return collection, nil
}

package cmd

import (
	"errors"
	"fmt"

	"github.com/rancher/cli/cliclient"
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
			{
				Name:      "delete",
				Aliases:   []string{"rm"},
				Usage:     "Delete a project by ID",
				ArgsUsage: "[PROJECTID]",
				Action:    deleteProject,
			},
		},
	}
}

func projectLs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	collection, err := getProjectList(ctx, c)
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "Project.ID"},
		{"NAME", "Project.Name"},
		{"STATE", "Project.State"},
		{"DESCRIPTION", "Project.Description"},
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
	if ctx.NArg() == 0 {
		return errors.New("project name is required")
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	newProj := &managementClient.Project{
		Name:      ctx.Args().First(),
		ClusterId: c.UserConfig.FocusedCluster(),
	}

	c.ManagementClient.Project.Create(newProj)
	return nil
}

func deleteProject(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return errors.New("project ID is required")
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	project, err := getProjectByID(ctx, c, ctx.Args().First())
	if nil != err {
		return err
	}

	err = c.ManagementClient.Project.Delete(&project)
	if nil != err {
		return err
	}

	return nil
}

func getProjectList(
	ctx *cli.Context,
	c *cliclient.MasterClient,
) (*managementClient.ProjectCollection, error) {
	filter := defaultListOpts(ctx)
	filter.Filters["clusterId"] = c.UserConfig.FocusedCluster()

	collection, err := c.ManagementClient.Project.List(filter)
	if err != nil {
		return nil, err
	}
	return collection, nil
}

func getProjectByID(
	ctx *cli.Context,
	c *cliclient.MasterClient,
	projectID string,
) (managementClient.Project, error) {
	projectCollection, err := getProjectList(ctx, c)
	if nil != err {
		return managementClient.Project{}, err
	}

	for _, project := range projectCollection.Data {
		if project.ID == projectID {
			return project, nil
		}
	}

	return managementClient.Project{}, fmt.Errorf("no project found with the ID [%s], run "+
		"`rancher projects` to see available projects", projectID)
}

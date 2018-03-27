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
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "format",
						Usage: "'json' or Custom format: '{{.Project.ID}} {{.Project.Name}}'",
					},
				},
			},
			{
				Name:        "create",
				Usage:       "Create a project",
				Description: "\nCreates a project in the current cluster.",
				ArgsUsage:   "[NEWPROJECTNAME...]",
				Action:      projectCreate,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "cluster",
						Usage: "Cluster ID to create the project in",
					},
				},
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

	clusterID := c.UserConfig.FocusedCluster()
	if ctx.String("cluster") != "" {
		cluster, err := getClusterByID(c, ctx.String("cluster"))
		if nil != err {
			return err
		}
		clusterID = cluster.ID
	}

	newProj := &managementClient.Project{
		Name:      ctx.Args().First(),
		ClusterId: clusterID,
	}

	_, err = c.ManagementClient.Project.Create(newProj)
	if nil != err {
		return err
	}
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

	project, err := getProjectByID(c, ctx.Args().First())
	if nil != err {
		return err
	}

	err = c.ManagementClient.Project.Delete(project)
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
	if nil != err {
		return nil, err
	}
	return collection, nil
}

func getProjectByID(
	c *cliclient.MasterClient,
	projectID string,
) (*managementClient.Project, error) {
	project, err := c.ManagementClient.Project.ByID(projectID)
	if nil != err {
		return nil, fmt.Errorf("no project found with the ID [%s], run "+
			"`rancher projects` to see available projects: %s", projectID, err)
	}
	return project, nil
}

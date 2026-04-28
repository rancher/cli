package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/rancher/cli/cliclient"
	"github.com/rancher/norman/types"
	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"github.com/urfave/cli/v3"
)

type ProjectData struct {
	ID      string
	Project managementClient.Project
}

func ProjectCommand() *cli.Command {
	return &cli.Command{
		Name:    "projects",
		Aliases: []string{"project"},
		Usage:   "Operations on projects",
		Action:  defaultAction(projectLs),
		Commands: []*cli.Command{
			{
				Name:        "ls",
				Usage:       "List projects",
				Description: "\nLists all projects in the current cluster.",
				ArgsUsage:   "None",
				Action:      projectLs,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "format",
						Usage: "'json', 'yaml' or Custom format: '{{.Project.ID}} {{.Project.Name}}'",
					},
					quietFlag,
				},
			},
			{
				Name:        "create",
				Usage:       "Create a project",
				Description: "\nCreates a project in the current cluster.",
				ArgsUsage:   "[NEWPROJECTNAME...]",
				Action:      projectCreate,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "cluster",
						Usage: "Cluster ID to create the project in",
					},
					&cli.StringFlag{
						Name:  "description",
						Usage: "Description to apply to the project",
					},
				},
			},
			{
				Name:      "delete",
				Aliases:   []string{"rm"},
				Usage:     "Delete a project by ID",
				ArgsUsage: "[PROJECTID PROJECTNAME]",
				Action:    projectDelete,
			},
			{
				Name:        "add-member-role",
				Usage:       "Add a member to the project",
				Action:      addProjectMemberRoles,
				Description: "Examples:\n #Create the roles of 'create-ns' and 'services-manage' for a user named 'user1'\n rancher project add-member-role user1 create-ns services-manage\n",
				ArgsUsage:   "[USERNAME, ROLE...]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "project-id",
						Usage: "Optional project ID to apply this change to, defaults to the current context",
					},
				},
			},
			{
				Name:        "delete-member-role",
				Usage:       "Delete a member from the project",
				Action:      deleteProjectMemberRoles,
				Description: "Examples:\n #Delete the roles of 'create-ns' and 'services-manage' for a user named 'user1'\n rancher project delete-member-role user1 create-ns services-manage\n",
				ArgsUsage:   "[USERNAME, ROLE...]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "project-id",
						Usage: "Optional project ID to apply this change to, defaults to the current context",
					},
				},
			},
			{
				Name:   "list-roles",
				Usage:  "List all available roles for a project",
				Action: listProjectRoles,
			},
			{
				Name:  "list-members",
				Usage: "List current members of the project",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					client, err := GetClient(cmd)
					if err != nil {
						return err
					}

					return listProjectMembers(
						cmd,
						cmd.Root().Writer,
						client.UserConfig,
						client.ManagementClient.ProjectRoleTemplateBinding,
						client.ManagementClient.Principal,
					)
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "project-id",
						Usage: "Optional project ID to list members for, defaults to the current context",
					},
					&cli.StringFlag{
						Name:  "format",
						Usage: "'json', 'yaml' or Custom format: '{{.ID }} {{.Member }}'",
					},
					quietFlag,
				},
			},
		},
	}
}

func projectLs(ctx context.Context, cmd *cli.Command) error {
	c, err := GetClient(cmd)
	if err != nil {
		return err
	}

	collection, err := getProjectList(cmd, c)
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "Project.Name"},
		{"STATE", "Project.State"},
		{"DESCRIPTION", "Project.Description"},
	}, cmd)

	defer writer.Close()

	for _, item := range collection.Data {
		writer.Write(&ProjectData{
			ID:      item.ID,
			Project: item,
		})
	}

	return writer.Err()
}

func projectCreate(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() == 0 {
		return cli.ShowSubcommandHelp(cmd)
	}

	c, err := GetClient(cmd)
	if err != nil {
		return err
	}

	clusterID := c.UserConfig.GetCurrentCluster()
	if cmd.String("cluster") != "" {
		resource, err := Lookup(c, cmd.String("cluster"), "cluster")
		if err != nil {
			return err
		}
		clusterID = resource.ID
	}

	newProj := &managementClient.Project{
		Name:        cmd.Args().First(),
		ClusterID:   clusterID,
		Description: cmd.String("description"),
	}

	_, err = c.ManagementClient.Project.Create(newProj)
	if err != nil {
		return err
	}
	return nil
}

func projectDelete(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() == 0 {
		return cli.ShowSubcommandHelp(cmd)
	}

	c, err := GetClient(cmd)
	if err != nil {
		return err
	}

	for _, arg := range cmd.Args().Slice() {
		resource, err := Lookup(c, arg, "project")
		if err != nil {
			return err
		}

		project, err := getProjectByID(c, resource.ID)
		if err != nil {
			return err
		}

		err = c.ManagementClient.Project.Delete(project)
		if err != nil {
			return err
		}
	}

	return nil
}

func addProjectMemberRoles(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() < 2 {
		return cli.ShowSubcommandHelp(cmd)
	}

	memberName := cmd.Args().First()

	roles := cmd.Args().Slice()[1:]

	c, err := GetClient(cmd)
	if err != nil {
		return err
	}

	member, err := searchForMember(cmd, c, memberName)
	if err != nil {
		return err
	}

	projectID := c.UserConfig.Project
	if cmd.String("project-id") != "" {
		projectID = cmd.String("project-id")
	}

	for _, role := range roles {
		rtb := managementClient.ProjectRoleTemplateBinding{
			ProjectID:      projectID,
			RoleTemplateID: role,
		}
		if member.PrincipalType == "user" {
			rtb.UserPrincipalID = member.ID
		} else {
			rtb.GroupPrincipalID = member.ID
		}
		_, err = c.ManagementClient.ProjectRoleTemplateBinding.Create(&rtb)
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteProjectMemberRoles(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() < 2 {
		return cli.ShowSubcommandHelp(cmd)
	}

	memberName := cmd.Args().First()

	roles := cmd.Args().Slice()[1:]

	c, err := GetClient(cmd)
	if err != nil {
		return err
	}

	member, err := searchForMember(cmd, c, memberName)
	if err != nil {
		return err
	}

	projectID := c.UserConfig.Project
	if cmd.String("project-id") != "" {
		projectID = cmd.String("project-id")
	}

	for _, role := range roles {
		filter := defaultListOpts(cmd)
		filter.Filters["projectId"] = projectID
		filter.Filters["roleTemplateId"] = role

		if member.PrincipalType == "user" {
			filter.Filters["userPrincipalId"] = member.ID
		} else {
			filter.Filters["groupPrincipalId"] = member.ID
		}

		bindings, err := c.ManagementClient.ProjectRoleTemplateBinding.List(filter)
		if err != nil {
			return err
		}

		for _, binding := range bindings.Data {
			err = c.ManagementClient.ProjectRoleTemplateBinding.Delete(&binding)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func listProjectRoles(ctx context.Context, cmd *cli.Command) error {
	return listRoles(cmd, "project")
}

type prtbLister interface {
	List(opts *types.ListOpts) (*managementClient.ProjectRoleTemplateBindingCollection, error)
}

func listProjectMembers(cmd *cli.Command, out io.Writer, config userConfig, prtbs prtbLister, principals principalGetter) error {
	projectID := config.GetCurrentProject()
	if cmd.String("project-id") != "" {
		projectID = cmd.String("project-id")
	}

	filter := defaultListOpts(cmd)
	filter.Filters["projectId"] = projectID

	bindings, err := prtbs.List(filter)
	if err != nil {
		return err
	}

	rtbs := make([]RoleTemplateBinding, 0, len(bindings.Data))

	for _, binding := range bindings.Data {
		parsedTime, err := createdTimeToHuman(binding.Created)
		if err != nil {
			return err
		}

		principalID := binding.UserPrincipalID
		if binding.GroupPrincipalID != "" {
			principalID = binding.GroupPrincipalID
		}

		rtbs = append(rtbs, RoleTemplateBinding{
			ID:      binding.ID,
			Member:  getMemberNameFromPrincipal(principals, principalID),
			Role:    binding.RoleTemplateID,
			Created: parsedTime,
		})
	}

	writerConfig := &TableWriterConfig{
		Format: cmd.String("format"),
		Quiet:  cmd.Bool("quiet"),
		Writer: out,
	}

	return listRoleTemplateBindings(writerConfig, rtbs)
}

func getProjectList(
	cmd *cli.Command,
	c *cliclient.MasterClient,
) (*managementClient.ProjectCollection, error) {
	filter := defaultListOpts(cmd)
	filter.Filters["clusterId"] = c.UserConfig.GetCurrentCluster()

	collection, err := c.ManagementClient.Project.List(filter)
	if err != nil {
		return nil, err
	}
	return collection, nil
}

func getProjectByID(
	c *cliclient.MasterClient,
	projectID string,
) (*managementClient.Project, error) {
	project, err := c.ManagementClient.Project.ByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("no project found with the ID [%s], run "+
			"`rancher projects` to see available projects: %s", projectID, err)
	}
	return project, nil
}

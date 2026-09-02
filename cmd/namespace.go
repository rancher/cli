package cmd

import (
	"context"
	"fmt"

	"github.com/rancher/cli/cliclient"
	clusterClient "github.com/rancher/rancher/pkg/client/generated/cluster/v3"
	"github.com/urfave/cli/v3"
)

type NamespaceData struct {
	ID        string
	Namespace clusterClient.Namespace
}

func NamespaceCommand() *cli.Command {
	return &cli.Command{
		Name:    "namespaces",
		Aliases: []string{"namespace"},
		Usage:   "Operations on namespaces",
		Action:  defaultAction(namespaceLs),
		Flags: []cli.Flag{
			quietFlag,
		},
		Commands: []*cli.Command{
			{
				Name:        "ls",
				Usage:       "List namespaces",
				Description: "\nLists all namespaces in the current project.",
				ArgsUsage:   "None",
				Action:      namespaceLs,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "all-namespaces",
						Usage: "List all namespaces in the current cluster",
					},
					&cli.StringFlag{
						Name:  "format",
						Usage: "'json', 'yaml' or Custom format: '{{.Namespace.ID}} {{.Namespace.Name}}'",
					},
					quietFlag,
				},
			},
			{
				Name:        "create",
				Usage:       "Create a namespace",
				Description: "\nCreates a namespace in the current cluster.",
				ArgsUsage:   "[NEWNAMESPACENAME...]",
				Action:      namespaceCreate,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "description",
						Usage: "Description to apply to the namespace",
					},
				},
			},
			{
				Name:      "delete",
				Aliases:   []string{"rm"},
				Usage:     "Delete a namespace by name or ID",
				ArgsUsage: "[NAMESPACEID NAMESPACENAME]",
				Action:    namespaceDelete,
			},
			{
				Name:      "move",
				Usage:     "Move a namespace to a different project",
				ArgsUsage: "[NAMESPACEID/NAMESPACENAME PROJECTID]",
				Action:    namespaceMove,
			},
		},
	}
}

func namespaceLs(ctx context.Context, cmd *cli.Command) error {
	c, err := GetClient(cmd)
	if err != nil {
		return err
	}

	collection, err := getNamespaceList(cmd, c)
	if err != nil {
		return err
	}

	if !cmd.Bool("all-namespaces") {
		var projectNamespaces []clusterClient.Namespace

		for _, namespace := range collection.Data {
			if namespace.ProjectID == c.UserConfig.Project {
				projectNamespaces = append(projectNamespaces, namespace)
			}

		}
		collection.Data = projectNamespaces
	}

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "Namespace.Name"},
		{"STATE", "Namespace.State"},
		{"PROJECT", "Namespace.ProjectID"},
		{"DESCRIPTION", "Namespace.Description"},
	}, cmd)

	defer writer.Close()

	for _, item := range collection.Data {
		writer.Write(&NamespaceData{
			ID:        item.ID,
			Namespace: item,
		})
	}

	return writer.Err()
}

func namespaceCreate(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() == 0 {
		return cli.ShowSubcommandHelp(cmd)
	}

	c, err := GetClient(cmd)
	if err != nil {
		return err
	}

	newNamespace := &clusterClient.Namespace{
		Name:        cmd.Args().First(),
		ProjectID:   c.UserConfig.Project,
		Description: cmd.String("description"),
	}

	_, err = c.ClusterClient.Namespace.Create(newNamespace)
	if err != nil {
		return err
	}

	return nil
}

func namespaceDelete(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() == 0 {
		return cli.ShowSubcommandHelp(cmd)
	}

	c, err := GetClient(cmd)
	if err != nil {
		return err
	}

	for _, arg := range cmd.Args().Slice() {
		resource, err := Lookup(c, arg, "namespace")
		if err != nil {
			return err
		}

		namespace, err := getNamespaceByID(c, resource.ID)
		if err != nil {
			return err
		}

		err = c.ClusterClient.Namespace.Delete(namespace)
		if err != nil {
			return err
		}
	}

	return nil
}

func namespaceMove(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() < 2 {
		return cli.ShowSubcommandHelp(cmd)
	}

	c, err := GetClient(cmd)
	if err != nil {
		return err
	}

	resource, err := Lookup(c, cmd.Args().First(), "namespace")
	if err != nil {
		return err
	}

	namespace, err := getNamespaceByID(c, resource.ID)
	if err != nil {
		return err
	}

	projResource, err := Lookup(c, cmd.Args().Get(1), "project")
	if err != nil {
		return err
	}

	proj, err := getProjectByID(c, projResource.ID)
	if err != nil {
		return err
	}

	if anno, ok := namespace.Annotations["cattle.io/appIds"]; ok && anno != "" {
		return fmt.Errorf("namespace %s cannot be moved", namespace.Name)
	}

	if _, ok := namespace.Actions["move"]; ok {
		move := &clusterClient.NamespaceMove{
			ProjectID: proj.ID,
		}
		return c.ClusterClient.Namespace.ActionMove(namespace, move)
	}

	update := make(map[string]string)
	update["projectId"] = proj.ID

	_, err = c.ClusterClient.Namespace.Update(namespace, update)
	if err != nil {
		return err
	}

	return nil
}

func getNamespaceList(
	cmd *cli.Command,
	c *cliclient.MasterClient,
) (*clusterClient.NamespaceCollection, error) {
	collection, err := c.ClusterClient.Namespace.List(defaultListOpts(cmd))
	if err != nil {
		return nil, err
	}
	return collection, nil
}

func getNamespaceByID(
	c *cliclient.MasterClient,
	namespaceID string,
) (*clusterClient.Namespace, error) {
	namespace, err := c.ClusterClient.Namespace.ByID(namespaceID)
	if err != nil {
		return nil, fmt.Errorf("no namespace found with the ID [%s], run "+
			"`rancher namespaces` to see available namespaces: %s", namespaceID, err)
	}
	return namespace, nil
}

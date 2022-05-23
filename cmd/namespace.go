package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/rancher/cli/cliclient"
	clusterClient "github.com/rancher/rancher/pkg/client/generated/cluster/v3"
	"github.com/urfave/cli"
)

type NamespaceData struct {
	ID        string
	Namespace clusterClient.Namespace
}

func NamespaceCommand() cli.Command {
	return cli.Command{
		Name:    "namespaces",
		Aliases: []string{"namespace"},
		Usage:   "Operations on namespaces",
		Action:  defaultAction(namespaceLs),
		Flags: []cli.Flag{
			quietFlag,
		},
		Subcommands: []cli.Command{
			{
				Name:        "ls",
				Usage:       "List namespaces",
				Description: "\nLists all namespaces in the current project.",
				ArgsUsage:   "None",
				Action:      namespaceLs,
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "all-namespaces",
						Usage: "List all namespaces in the current cluster",
					},
					cli.StringFlag{
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
				ArgsUsage:   "[NEWPROJECTNAME...]",
				Action:      namespaceCreate,
				Flags: []cli.Flag{
					cli.StringFlag{
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

func namespaceLs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	collection, err := getNamespaceList(ctx, c)
	if err != nil {
		return err
	}

	if !ctx.Bool("all-namespaces") {
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
	}, ctx)

	defer writer.Close()

	for _, item := range collection.Data {
		writer.Write(&NamespaceData{
			ID:        item.ID,
			Namespace: item,
		})
	}

	return writer.Err()
}

func namespaceCreate(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	newNamespace := &clusterClient.Namespace{
		Name:        ctx.Args().First(),
		ProjectID:   c.UserConfig.Project,
		Description: ctx.String("description"),
	}

	_, err = c.ClusterClient.Namespace.Create(newNamespace)
	if err != nil {
		return err
	}

	return nil
}

func namespaceDelete(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	for _, arg := range ctx.Args() {
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

func namespaceMove(ctx *cli.Context) error {
	if ctx.NArg() < 2 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	resource, err := Lookup(c, ctx.Args().First(), "namespace")
	if err != nil {
		return err
	}

	namespace, err := getNamespaceByID(c, resource.ID)
	if err != nil {
		return err
	}

	projResource, err := Lookup(c, ctx.Args().Get(1), "project")
	if err != nil {
		return err
	}

	proj, err := getProjectByID(c, projResource.ID)
	if err != nil {
		return err
	}

	if anno, ok := namespace.Annotations["cattle.io/appIds"]; ok && anno != "" {
		return errors.Errorf("Namespace %v cannot be moved", namespace.Name)
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
	ctx *cli.Context,
	c *cliclient.MasterClient,
) (*clusterClient.NamespaceCollection, error) {
	collection, err := c.ClusterClient.Namespace.List(defaultListOpts(ctx))
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

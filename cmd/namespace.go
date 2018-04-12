package cmd

import (
	"errors"
	"fmt"

	"github.com/rancher/cli/cliclient"
	clusterClient "github.com/rancher/types/client/cluster/v3"
	"github.com/urfave/cli"
)

type NamespaceData struct {
	Namespace clusterClient.Namespace
}

func NamespaceCommand() cli.Command {
	return cli.Command{
		Name:    "namespaces",
		Aliases: []string{"namespace"},
		Usage:   "Operations on namespaces",
		Action:  defaultAction(namespaceLs),
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
				Name:  "associate",
				Usage: "Associate a namespace with a project",
				Description: "\nAssociates a namespace with a project. If no " +
					"[PROJECTID] is provided the namespace will be unassociated from all projects",
				ArgsUsage: "[NAMESPACEID/NAMESPACENAME PROJECTID]",
				Action:    namespaceAssociate,
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
		{"ID", "Namespace.ID"},
		{"NAME", "Namespace.Name"},
		{"STATE", "Namespace.State"},
		{"PROJECT", "Namespace.ProjectID"},
		{"DESCRIPTION", "Namespace.Description"},
	}, ctx)

	defer writer.Close()

	for _, item := range collection.Data {
		writer.Write(&NamespaceData{
			Namespace: item,
		})
	}

	return writer.Err()
}

func namespaceCreate(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return errors.New("namespace name is required")
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
		return errors.New("namespace name or ID is required")
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	resource, err := Lookup(c, ctx.Args().First(), "namespace")
	if nil != err {
		return err
	}

	namespace, err := getNamespaceByID(c, resource.ID)
	if nil != err {
		return err
	}

	err = c.ClusterClient.Namespace.Delete(namespace)
	if nil != err {
		return err
	}

	return nil
}

func namespaceAssociate(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return errors.New("namespace is required")
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	resource, err := Lookup(c, ctx.Args().First(), "namespace")
	if nil != err {
		return err
	}

	namespace, err := getNamespaceByID(c, resource.ID)
	if err != nil {
		return err
	}

	update := make(map[string]string)
	update["projectId"] = ctx.Args().Get(1)

	_, err = c.ClusterClient.Namespace.Update(namespace, update)
	if nil != err {
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
	if nil != err {
		return nil, fmt.Errorf("no namespace found with the ID [%s], run "+
			"`rancher namespaces` to see available namespaces: %s", namespaceID, err)
	}
	return namespace, nil
}

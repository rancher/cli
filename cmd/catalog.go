package cmd

import (
	managementClient "github.com/rancher/types/client/management/v3"
	"github.com/urfave/cli"
)

const (
	addCatalogDescription = `
Add a new catalog to the Rancher server

Example:
	# Add a catalog
	$ rancher catalog add foo https://my.catalog

	# Add a catalog and specify the branch to use
	$ rancher catalog add --branch awesomebranch foo https://my.catalog
`
)

type CatalogData struct {
	ID      string
	Catalog managementClient.Catalog
}

func CatalogCommand() cli.Command {
	catalogLsFlags := []cli.Flag{
		formatFlag,
		quietFlag,
	}

	return cli.Command{
		Name:   "catalog",
		Usage:  "Operations with catalogs",
		Action: defaultAction(catalogLs),
		Flags:  catalogLsFlags,
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "ls",
				Usage:       "List catalogs",
				Description: "\nList all catalogs in the current Rancher server",
				ArgsUsage:   "None",
				Action:      catalogLs,
				Flags:       catalogLsFlags,
			},
			cli.Command{
				Name:        "add",
				Usage:       "Add a catalog",
				Description: addCatalogDescription,
				ArgsUsage:   "[NAME, URL]",
				Action:      catalogAdd,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "branch",
						Usage: "Branch from the url to use",
						Value: "master",
					},
				},
			},
			cli.Command{
				Name:        "delete",
				Usage:       "Delete a catalog",
				Description: "\nDelete a catalog from the Rancher server",
				ArgsUsage:   "[CATALOG_NAME/CATALOG_ID]",
				Action:      catalogDelete,
			},
		},
	}
}

func catalogLs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	collection, err := c.ManagementClient.Catalog.List(defaultListOpts(ctx))
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "Catalog.Name"},
		{"URL", "Catalog.URL"},
		{"BRANCH", "Catalog.Branch"},
		{"KIND", "Catalog.Kind"},
	}, ctx)

	defer writer.Close()

	for _, item := range collection.Data {
		writer.Write(&CatalogData{
			ID:      item.ID,
			Catalog: item,
		})
	}

	return writer.Err()

}

func catalogAdd(ctx *cli.Context) error {
	if len(ctx.Args()) < 2 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	catalog := &managementClient.Catalog{
		Branch: ctx.String("branch"),
		Name:   ctx.Args().First(),
		Kind:   "helm",
		URL:    ctx.Args().Get(1),
	}

	_, err = c.ManagementClient.Catalog.Create(catalog)
	if err != nil {
		return err
	}

	return nil
}

func catalogDelete(ctx *cli.Context) error {
	if len(ctx.Args()) < 1 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	for _, arg := range ctx.Args() {
		resource, err := Lookup(c, arg, "catalog")
		if err != nil {
			return err
		}

		catalog, err := c.ManagementClient.Catalog.ByID(resource.ID)
		if err != nil {
			return err
		}

		err = c.ManagementClient.Catalog.Delete(catalog)
		if err != nil {
			return err
		}
	}
	return nil
}

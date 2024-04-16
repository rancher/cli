package cmd

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"github.com/sirupsen/logrus"
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

	# Add a catalog and specify the helm version to use. Specify 'v2' for helm 2 and 'v3' for helm 3
	$ rancher catalog add --helm-version v3 foo https://my.catalog
`

	refreshCatalogDescription = `
Refresh a catalog on the Rancher server

Example:
	# Refresh a catalog
	$ rancher catalog refresh foo

	# Refresh multiple catalogs
	$ rancher catalog refresh foo bar baz

	# Refresh all catalogs
	$ rancher catalog refresh --all

	# Refresh is asynchronous unless you specify '--wait'
	$ rancher catalog refresh --all --wait --wait-timeout=60

	# Default wait timeout is 60 seconds, set to 0 to remove the timeout
	$ rancher catalog refresh --all --wait --wait-timeout=0
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
		cli.BoolFlag{
			Name:  "verbose,v",
			Usage: "Include the catalog's state",
		},
	}

	return cli.Command{
		Name:   "catalog",
		Usage:  "Operations with catalogs",
		Action: defaultAction(catalogLs),
		Flags:  catalogLsFlags,
		Subcommands: []cli.Command{
			{
				Name:        "ls",
				Usage:       "List catalogs",
				Description: "\nList all catalogs in the current Rancher server",
				ArgsUsage:   "None",
				Action:      catalogLs,
				Flags:       catalogLsFlags,
			},
			{
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
					cli.StringFlag{
						Name:  "helm-version",
						Usage: "Version of helm the app(s) in your catalog will use for deployment. Use 'v2' for helm 2 or 'v3' for helm 3",
						Value: "v2",
					},
				},
			},
			{
				Name:        "delete",
				Usage:       "Delete a catalog",
				Description: "\nDelete a catalog from the Rancher server",
				ArgsUsage:   "[CATALOG_NAME/CATALOG_ID]",
				Action:      catalogDelete,
			},
			{
				Name:        "refresh",
				Usage:       "Refresh catalog templates",
				Description: refreshCatalogDescription,
				ArgsUsage:   "[CATALOG_NAME/CATALOG_ID]...",
				Action:      catalogRefresh,
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "all",
						Usage: "Refresh all catalogs",
					},
					cli.BoolFlag{
						Name:  "wait,w",
						Usage: "Wait for catalog(s) to become active",
					},
					cli.IntFlag{
						Name:  "wait-timeout",
						Usage: "Wait timeout duration in seconds",
						Value: 60,
					},
				},
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

	fields := [][]string{
		{"ID", "ID"},
		{"NAME", "Catalog.Name"},
		{"URL", "Catalog.URL"},
		{"BRANCH", "Catalog.Branch"},
		{"KIND", "Catalog.Kind"},
		{"HELMVERSION", "Catalog.HelmVersion"},
	}

	if ctx.Bool("verbose") {
		fields = append(fields, []string{"STATE", "Catalog.State"})
	}

	writer := NewTableWriter(fields, ctx)

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
		Branch:      ctx.String("branch"),
		Name:        ctx.Args().First(),
		Kind:        "helm",
		URL:         ctx.Args().Get(1),
		HelmVersion: strings.ToLower(ctx.String("helm-version")),
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

func catalogRefresh(ctx *cli.Context) error {
	if len(ctx.Args()) < 1 && !ctx.Bool("all") {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	var catalogs []managementClient.Catalog

	if ctx.Bool("all") {
		opts := baseListOpts()

		collection, err := c.ManagementClient.Catalog.List(opts)
		if err != nil {
			return err
		}

		// save the catalogs in case we need to wait for them to become active
		catalogs = collection.Data

		_, err = c.ManagementClient.Catalog.CollectionActionRefresh(collection)
		if err != nil {
			return err
		}

	} else {
		for _, arg := range ctx.Args() {
			resource, err := Lookup(c, arg, "catalog")
			if err != nil {
				return err
			}

			catalog, err := c.ManagementClient.Catalog.ByID(resource.ID)
			if err != nil {
				return err
			}

			// collect the refreshing catalogs in case we need to wait for them later
			catalogs = append(catalogs, *catalog)

			_, err = c.ManagementClient.Catalog.ActionRefresh(catalog)
			if err != nil {
				return err
			}
		}
	}

	if ctx.Bool("wait") {
		timeout := time.Duration(ctx.Int("wait-timeout")) * time.Second
		start := time.Now()

		logrus.Debugf("catalog: waiting for catalogs to become active (timeout=%v)", timeout)

		for _, catalog := range catalogs {

			logrus.Debugf("catalog: waiting for %s to become active", catalog.Name)

			resource, err := Lookup(c, catalog.Name, "catalog")
			if err != nil {
				return err
			}

			catalog, err := c.ManagementClient.Catalog.ByID(resource.ID)
			if err != nil {
				return err
			}

			for catalog.State != "active" {
				time.Sleep(time.Second)
				catalog, err = c.ManagementClient.Catalog.ByID(resource.ID)
				if err != nil {
					return err
				}

				if timeout > 0 && time.Since(start) > timeout {
					return errors.New("catalog: timed out waiting for refresh")
				}
			}

		}
		logrus.Debugf("catalog: waited for %v", time.Since(start))
	}

	return nil
}

package cmd

import (
	"fmt"
	"github.com/rancher/go-rancher/v3"
	"github.com/urfave/cli"
)

func ClusterCommand() cli.Command {
	clusterLsFlags := []cli.Flag{
		listAllFlag(),
		cli.BoolFlag{
			Name:  "quiet,q",
			Usage: "Only display IDs",
		},
		cli.StringFlag{
			Name:  "format",
			Usage: "'json' or Custom format: '{{.ID}} {{.Environment.Name}}'",
		},
	}

	return cli.Command{
		Name:   "cluster",
		Usage:  "Interact with cluster",
		Action: defaultAction(clusterLs),
		Flags:  clusterLsFlags,
		Subcommands: []cli.Command{
			{
				Name:        "ls",
				Usage:       "List clusters",
				Description: "\nList all clusters in the current rancher setup\n",
				ArgsUsage:   "None",
				Action:      clusterLs,
				Flags:       clusterLsFlags,
			},
			{
				Name:        "create",
				Usage:       "create cluster",
				Description: "\nCreate cluster\n",
				ArgsUsage:   "None",
				Action:      clusterCreate,
			},
			{
				Name:        "rm",
				Usage:       "remove cluster",
				Description: "\nRemove cluster\n",
				ArgsUsage:   "None",
				Action:      clusterRemove,
			},
			{
				Name:        "export",
				Usage:       "export an external cluster",
				Description: "\nExport an external cluster inside the current cluster",
				ArgsUsage:   "None",
				Action:      clusterExport,
			},
		},
	}
}

func clusterLs(ctx *cli.Context) error {
	c, err := GetRawClient(ctx)
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "Cluster.Id"},
		{"NAME", "Cluster.Name"},
		{"STATE", "Cluster.State"},
		{"CREATED", "Cluster.Created"},
		{"EMBEDDED", "Cluster.Embedded"},
	}, ctx)
	defer writer.Close()

	clusters, err := c.Cluster.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"removed_null": "true",
		},
	})
	if err != nil {
		return err
	}
	for _, cluster := range clusters.Data {
		writer.Write(ClusterData{cluster})
	}
	return writer.Err()
}

func clusterCreate(ctx *cli.Context) error {
	c, err := GetRawClient(ctx)
	if err != nil {
		return err
	}
	name := RandomName()
	if ctx.NArg() > 0 {
		name = ctx.Args()[0]
	}
	cluster, err := c.Cluster.Create(&client.Cluster{
		Name: name,
	})
	if err != nil {
		return err
	}
	fmt.Println(cluster.Id)
	return nil
}

func clusterRemove(ctx *cli.Context) error {
	c, err := GetRawClient(ctx)
	if err != nil {
		return err
	}

	return forEachResourceWithClient(c, ctx, []string{"cluster"}, func(c *client.RancherClient, resource *client.Resource) (string, error) {
		return resource.Id, c.Delete(resource)
	})
}

func clusterExport(ctx *cli.Context) error {
	fmt.Println("Support coming soon")
	return nil
}

type ClusterData struct {
	Cluster client.Cluster
}

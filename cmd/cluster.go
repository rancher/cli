package cmd

import (
	"errors"
	"fmt"

	"github.com/rancher/cli/cliclient"
	managementClient "github.com/rancher/types/client/management/v3"
	"github.com/urfave/cli"
)

const (
	importDescription = `
Imports an existing cluster to be used in rancher by using a generated kubectl 
command to run in your existing Kubernetes cluster.
`
)

type ClusterData struct {
	Cluster managementClient.Cluster
}

func ClusterCommand() cli.Command {
	return cli.Command{
		Name:    "clusters",
		Aliases: []string{"cluster"},
		Usage:   "Operations on clusters",
		Action:  defaultAction(clusterLs),
		Subcommands: []cli.Command{
			{
				Name:        "ls",
				Usage:       "List clusters",
				Description: "Lists all clusters",
				ArgsUsage:   "None",
				Action:      clusterLs,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "format",
						Usage: "'json' or Custom format: '{{.Cluster.ID}} {{.Cluster.Name}}'",
					},
				},
			},
			{
				Name:        "create",
				Usage:       "Creates a new empty cluster",
				Description: "Creates a new empty cluster",
				ArgsUsage:   "[NEWCLUSTERNAME...]",
				Action:      clusterCreate,
			},
			{
				Name:        "import",
				Usage:       "Import an existing Kubernetes cluster into a Rancher cluster",
				Description: importDescription,
				ArgsUsage:   "[CLUSTERID]",
				Action:      clusterImport,
			},
			{
				Name:      "add-node",
				Usage:     "Returns the command needed to add a node to an existing Rancher cluster",
				ArgsUsage: "[CLUSTERID]",
				Action:    clusterAddNode,
				Flags: []cli.Flag{
					cli.StringSliceFlag{
						Name:  "label",
						Usage: "Label to apply to a node in the format [name]=[value]",
					},
					cli.BoolFlag{
						Name:  "etcd",
						Usage: "Use node for etcd",
					},
					cli.BoolFlag{
						Name:  "management",
						Usage: "Use node for management",
					},
					cli.BoolFlag{
						Name:  "worker",
						Usage: "Use node as a worker",
					},
				},
			},
			{
				Name:      "delete",
				Aliases:   []string{"rm"},
				Usage:     "Delete a cluster",
				ArgsUsage: "[CLUSTERID]",
				Action:    deleteCluster,
			},
			{
				Name:      "kubeconfig",
				Aliases:   []string{"kf"},
				Usage:     "Return the kube config used to access the cluster",
				ArgsUsage: "[CLUSTERID]",
				Action:    clusterKubeConfig,
			},
		},
	}
}

func clusterLs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	collection, err := c.ManagementClient.Cluster.List(defaultListOpts(ctx))
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "Cluster.ID"},
		{"NAME", "Cluster.Name"},
		{"STATE", "Cluster.State"},
		{"DESCRIPTION", "Cluster.Description"},
	}, ctx)

	defer writer.Close()

	for _, item := range collection.Data {
		writer.Write(&ClusterData{
			Cluster: item,
		})
	}

	return writer.Err()
}

func clusterCreate(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return errors.New("cluster name is required")
	}
	c, err := GetClient(ctx)
	if nil != err {
		return err
	}

	cluster, err := c.ManagementClient.Cluster.Create(&managementClient.Cluster{
		Name: ctx.Args().First(),
	})

	if nil != err {
		return err
	}

	fmt.Printf("Successfully created cluster %v\n", cluster.Name)
	return nil
}

func clusterImport(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return errors.New("cluster ID is required")
	}

	c, err := GetClient(ctx)
	if nil != err {
		return err
	}

	cluster, err := getClusterByID(c, ctx.Args().First())
	if nil != err {
		return err
	}

	clusterToken, err := getClusterRegToken(ctx, c, cluster.ID)
	if nil != err {
		return err
	}

	fmt.Printf("Run the following command in your cluster: %v", clusterToken.Command)

	return nil
}

// clusterAddNode prints the command needed to add a node to a cluster
func clusterAddNode(ctx *cli.Context) error {
	var clusterName string

	if ctx.NArg() == 0 {
		return errors.New("cluster ID is required")
	}

	clusterName = ctx.Args().First()

	c, err := GetClient(ctx)
	if nil != err {
		return err
	}

	cluster, err := getClusterByID(c, clusterName)
	if nil != err {
		return err
	}

	clusterToken, err := getClusterRegToken(ctx, c, cluster.ID)
	if nil != err {
		return err
	}

	var roleFlags string

	if ctx.Bool("etcd") {
		roleFlags = roleFlags + " --etcd"
	}

	if ctx.Bool("management") {
		roleFlags = roleFlags + " --controlplane"
	}

	if ctx.Bool("worker") {
		roleFlags = roleFlags + " --worker"
	}

	command := clusterToken.NodeCommand + roleFlags

	if labels := ctx.StringSlice("label"); labels != nil {
		for _, label := range labels {
			command = command + fmt.Sprintf(" --label %v", label)
		}
	}

	fmt.Printf("Run this command on an existing machine already running a "+
		"supported version of Docker:\n%v\n", command)

	return nil
}

func deleteCluster(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return errors.New("cluster ID is required")
	}

	c, err := GetClient(ctx)
	if nil != err {
		return err
	}

	cluster, err := getClusterByID(c, ctx.Args().First())
	if nil != err {
		return err
	}

	err = c.ManagementClient.Cluster.Delete(cluster)
	if nil != err {
		return err
	}

	return nil
}

func clusterKubeConfig(ctx *cli.Context) error {
	var clusterName string

	if ctx.NArg() == 0 {
		return errors.New("cluster ID is required")
	}

	clusterName = ctx.Args().First()

	c, err := GetClient(ctx)
	if nil != err {
		return err
	}

	cluster, err := getClusterByID(c, clusterName)
	if nil != err {
		return err
	}

	config, err := c.ManagementClient.Cluster.ActionGenerateKubeconfig(cluster)
	if nil != err {
		return err
	}
	fmt.Println(config.Config)
	return nil
}

// getClusterRegToken will return an existing token or create one if none exist
func getClusterRegToken(
	ctx *cli.Context,
	c *cliclient.MasterClient,
	clusterID string,
) (managementClient.ClusterRegistrationToken, error) {
	tokenOpts := defaultListOpts(ctx)
	tokenOpts.Filters["clusterId"] = clusterID

	clusterTokenCollection, err := c.ManagementClient.ClusterRegistrationToken.List(tokenOpts)
	if nil != err {
		return managementClient.ClusterRegistrationToken{}, err
	}

	if len(clusterTokenCollection.Data) == 0 {
		crt := &managementClient.ClusterRegistrationToken{
			ClusterId: clusterID,
		}
		clusterToken, err := c.ManagementClient.ClusterRegistrationToken.Create(crt)
		if nil != err {
			return managementClient.ClusterRegistrationToken{}, err
		}
		return *clusterToken, nil
	}
	return clusterTokenCollection.Data[0], nil
}

func getClusterByID(
	c *cliclient.MasterClient,
	clusterID string,
) (*managementClient.Cluster, error) {
	cluster, err := c.ManagementClient.Cluster.ByID(clusterID)
	if nil != err {
		return nil, fmt.Errorf("no cluster found with the ID [%s], run "+
			"`rancher clusters` to see available clusters: %s", clusterID, err)
	}
	return cluster, nil
}

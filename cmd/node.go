package cmd

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/rancher/cli/cliclient"
	managementClient "github.com/rancher/types/client/management/v3"
	"github.com/urfave/cli"
)

type NodeData struct {
	ID   string
	Node managementClient.Node
	Name string
	Pool string
}

func NodeCommand() cli.Command {
	return cli.Command{
		Name:    "nodes",
		Aliases: []string{"node"},
		Usage:   "Operations on nodes",
		Action:  defaultAction(nodeLs),
		Subcommands: []cli.Command{
			{
				Name:        "ls",
				Usage:       "List nodes",
				Description: "\nLists all nodes in the current cluster.",
				ArgsUsage:   "None",
				Action:      nodeLs,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "format",
						Usage: "'json', 'yaml' or Custom format: '{{.Node.ID}} {{.Node.Name}}'",
					},
					quietFlag,
				},
			},
			{
				Name:      "delete",
				Aliases:   []string{"rm"},
				Usage:     "Delete a node by ID",
				ArgsUsage: "[NODEID NODENAME]",
				Action:    nodeDelete,
			},
		},
	}
}

func nodeLs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	collection, err := getNodesList(ctx, c, c.UserConfig.FocusedCluster())
	if err != nil {
		return err
	}

	nodePools, err := getNodePools(ctx, c)
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "Name"},
		{"STATE", "Node.State"},
		{"POOL", "Pool"},
		{"DESCRIPTION", "Node.Description"},
	}, ctx)

	defer writer.Close()

	for _, item := range collection.Data {
		writer.Write(&NodeData{
			ID:   item.ID,
			Node: item,
			Name: getNodeName(item),
			Pool: getNodePoolName(item, nodePools),
		})
	}

	return writer.Err()
}

func nodeDelete(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	for _, arg := range ctx.Args() {
		resource, err := Lookup(c, arg, "node")
		if err != nil {
			return err
		}

		node, err := getNodeByID(ctx, c, resource.ID)
		if err != nil {
			return err
		}

		if _, ok := node.Links["remove"]; !ok {
			logrus.Warnf("node %v is externally managed and must be deleted "+
				"through the provider", getNodeName(node))
			continue
		}

		err = c.ManagementClient.Node.Delete(&node)
		if err != nil {
			return err
		}
	}
	return nil
}

func getNodesList(
	ctx *cli.Context,
	c *cliclient.MasterClient,
	clusterID string,
) (*managementClient.NodeCollection, error) {
	filter := defaultListOpts(ctx)
	filter.Filters["clusterId"] = clusterID

	collection, err := c.ManagementClient.Node.List(filter)
	if err != nil {
		return nil, err
	}
	return collection, nil
}

func getNodeByID(
	ctx *cli.Context,
	c *cliclient.MasterClient,
	nodeID string,
) (managementClient.Node, error) {
	nodeCollection, err := getNodesList(ctx, c, c.UserConfig.FocusedCluster())
	if err != nil {
		return managementClient.Node{}, err
	}

	for _, node := range nodeCollection.Data {
		if node.ID == nodeID {
			return node, nil
		}
	}

	return managementClient.Node{}, fmt.Errorf("no node found with the ID [%s], run "+
		"`rancher nodes` to see available nodes", nodeID)
}

func getNodeName(node managementClient.Node) string {
	if node.Name != "" {
		return node.Name
	} else if node.NodeName != "" {
		return node.NodeName
	} else if node.RequestedHostname != "" {
		return node.RequestedHostname
	}
	return node.ID
}

func getNodePools(
	ctx *cli.Context,
	c *cliclient.MasterClient,
) (*managementClient.NodePoolCollection, error) {
	filter := defaultListOpts(ctx)
	filter.Filters["clusterId"] = c.UserConfig.FocusedCluster()

	collection, err := c.ManagementClient.NodePool.List(filter)
	if err != nil {
		return nil, err
	}
	return collection, nil
}

func getNodePoolName(node managementClient.Node, pools *managementClient.NodePoolCollection) string {
	for _, pool := range pools.Data {
		if node.NodePoolID == pool.ID {
			return pool.HostnamePrefix
		}
	}
	return ""
}

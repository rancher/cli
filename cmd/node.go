package cmd

import (
	"errors"
	"fmt"

	"github.com/rancher/cli/cliclient"
	managementClient "github.com/rancher/types/client/management/v3"
	"github.com/urfave/cli"
)

type NodeData struct {
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
						Usage: "'json' or Custom format: '{{.Node.ID}} {{.Node.Name}}'",
					},
				},
			},
			{
				Name:      "delete",
				Aliases:   []string{"rm"},
				Usage:     "Delete a node by ID",
				ArgsUsage: "[NODEID]",
				Action:    deleteNode,
			},
		},
	}
}

func nodeLs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if nil != err {
		return err
	}

	collection, err := getNodesList(ctx, c, c.UserConfig.FocusedCluster())
	if nil != err {
		return err
	}

	nodePools, err := getNodePools(ctx, c)
	if nil != err {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "Node.ID"},
		{"NAME", "Name"},
		{"STATE", "Node.State"},
		{"POOL", "Pool"},
		{"DESCRIPTION", "Node.Description"},
	}, ctx)

	defer writer.Close()

	for _, item := range collection.Data {
		writer.Write(&NodeData{
			Node: item,
			Name: getNodeName(item),
			Pool: getNodePoolName(item, nodePools),
		})
	}

	return writer.Err()
}

func deleteNode(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return errors.New("node ID is required")
	}

	c, err := GetClient(ctx)
	if nil != err {
		return err
	}

	node, err := getNodeByID(ctx, c, ctx.Args().First())
	if nil != err {
		return err
	}

	if _, ok := node.Links["remove"]; !ok {
		return fmt.Errorf("node %v is externally managed and must be deleted "+
			"through the provider", getNodeName(node))
	}

	err = c.ManagementClient.Node.Delete(&node)
	if nil != err {
		return err
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
	if nil != err {
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
	if nil != err {
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
	if nil != err {
		return nil, err
	}
	return collection, nil
}

func getNodePoolName(node managementClient.Node, pools *managementClient.NodePoolCollection) string {
	for _, pool := range pools.Data {
		if node.NodePoolId == pool.ID {
			return pool.HostnamePrefix
		}
	}
	return ""
}

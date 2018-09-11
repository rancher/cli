package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/cli/cliclient"
	managementClient "github.com/rancher/types/client/management/v3"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	importDescription = `
Imports an existing cluster to be used in rancher by using a generated kubectl 
command to run in your existing Kubernetes cluster.
`
	importClusterNotice = "If you get an error about 'certificate signed by unknown authority' " +
		"because your Rancher installation is running with an untrusted/self-signed SSL " +
		"certificate, run the command below instead to bypass the certificate check:"
)

type ClusterData struct {
	ID       string
	Current  string
	Cluster  managementClient.Cluster
	Name     string
	Provider string
	Nodes    int
	CPU      string
	RAM      string
	Pods     string
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
						Usage: "'json', 'yaml' or Custom format: '{{.Cluster.ID}} {{.Cluster.Name}}'",
					},
					quietFlag,
				},
			},
			{
				Name:        "create",
				Usage:       "Creates a new empty cluster",
				Description: "Creates a new empty cluster",
				ArgsUsage:   "[NEWCLUSTERNAME...]",
				Action:      clusterCreate,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "description",
						Usage: "Description to apply to the cluster",
					},
					cli.BoolTFlag{
						Name:  "disable-docker-version",
						Usage: "Allow unsupported versions of docker on the nodes, [default=true]",
					},
					cli.BoolFlag{
						Name:  "import",
						Usage: "Mark the cluster for import, this is required if the cluster is going to be used to import an existing k8s cluster",
					},
					cli.StringFlag{
						Name:  "k8s-version",
						Usage: "Kubernetes version to use for the cluster, pass in 'list' to see available versions",
					},
					cli.StringFlag{
						Name:  "network-provider",
						Usage: "Network provider for the cluster (flannel, canal, calico)",
						Value: "canal",
					},
					cli.StringFlag{
						Name:  "psp-default-policy",
						Usage: "Default pod security policy to apply",
					},
					cli.StringFlag{
						Name:  "rke-config",
						Usage: "Location of an rke config file to import. Can be JSON or YAML format",
					},
				},
			},
			{
				Name:        "import",
				Usage:       "Import an existing Kubernetes cluster into a Rancher cluster",
				Description: importDescription,
				ArgsUsage:   "[CLUSTERID CLUSTERNAME]",
				Action:      clusterImport,
			},
			{
				Name:      "add-node",
				Usage:     "Outputs the docker command needed to add a node to an existing Rancher cluster",
				ArgsUsage: "[CLUSTERID CLUSTERNAME]",
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
				ArgsUsage: "[CLUSTERID/CLUSTERNAME...]",
				Action:    clusterDelete,
			},
			{
				Name:      "export",
				Usage:     "Export a cluster",
				ArgsUsage: "[CLUSTERID/CLUSTERNAME...]",
				Action:    clusterExport,
			},
			{
				Name:      "kubeconfig",
				Aliases:   []string{"kf"},
				Usage:     "Return the kube config used to access the cluster",
				ArgsUsage: "[CLUSTERID CLUSTERNAME]",
				Action:    clusterKubeConfig,
			},
			{
				Name:        "add-member-role",
				Usage:       "Add a member to the cluster",
				Action:      addClusterMemberRoles,
				Description: "Examples:\n #Create the roles of 'nodes-view' and 'projects-view' for a user named 'user1'\n rancher cluster add-member-role user1 nodes-view projects-view\n",
				ArgsUsage:   "[USERNAME, ROLE...]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "cluster-id",
						Usage: "Optional cluster ID to add member role to, defaults to the current context",
					},
				},
			},
			{
				Name:        "delete-member-role",
				Usage:       "Delete a member from the cluster",
				Action:      deleteClusterMemberRoles,
				Description: "Examples:\n #Delete the roles of 'nodes-view' and 'projects-view' for a user named 'user1'\n rancher cluster delete-member-role user1 nodes-view projects-view\n",
				ArgsUsage:   "[USERNAME, ROLE...]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "cluster-id",
						Usage: "Optional cluster ID to remove member role from, defaults to the current context",
					},
				},
			},
			{
				Name:   "list-roles",
				Usage:  "List all available roles for a cluster",
				Action: listClusterRoles,
			},
			{
				Name:   "list-members",
				Usage:  "List current members of the cluster",
				Action: listClusterMembers,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "cluster-id",
						Usage: "Optional cluster ID to list members for, defaults to the current context",
					},
				},
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
		{"CURRENT", "Current"},
		{"ID", "ID"},
		{"STATE", "Cluster.State"},
		{"NAME", "Name"},
		{"PROVIDER", "Provider"},
		{"NODES", "Nodes"},
		{"CPU", "CPU"},
		{"RAM", "RAM"},
		{"PODS", "Pods"},
	}, ctx)

	defer writer.Close()

	for _, item := range collection.Data {
		var current string
		if item.ID == c.UserConfig.FocusedCluster() {
			current = "*"
		}
		nodeCount, err := getClusterNodeCount(ctx, c, item.ID)
		if nil != err {
			logrus.Errorf("error getting cluster node count for cluster %s: %s", item.Name, err)
		}

		writer.Write(&ClusterData{
			ID:       item.ID,
			Current:  current,
			Cluster:  item,
			Name:     getClusterName(&item),
			Provider: getClusterProvider(item),
			Nodes:    nodeCount,
			CPU:      getClusterCPU(item),
			RAM:      getClusterRAM(item),
			Pods:     getClusterPods(item),
		})
	}

	return writer.Err()
}

func clusterCreate(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}
	c, err := GetClient(ctx)
	if nil != err {
		return err
	}

	if ctx.String("k8s-version") != "" {
		k8sVersions := getClusterK8sOptions(c)
		if ok := findStringInArray(ctx.String("k8s-version"), k8sVersions); !ok {
			fmt.Println("Available Kubernetes versions:")
			for _, val := range k8sVersions {
				fmt.Println(val)
			}
			return nil
		}
	}

	rkeConfig, err := getRKEConfig(ctx)
	if err != nil {
		return err
	}

	clusterConfig := &managementClient.Cluster{
		Name:                          ctx.Args().First(),
		Description:                   ctx.String("description"),
		RancherKubernetesEngineConfig: rkeConfig,
	}

	if ctx.String("psp-default-policy") != "" {
		clusterConfig.DefaultPodSecurityPolicyTemplateID = ctx.String("psp-default-policy")
	}

	createdCluster, err := c.ManagementClient.Cluster.Create(clusterConfig)

	if nil != err {
		return err
	}

	fmt.Printf("Successfully created cluster %v\n", createdCluster.Name)
	return nil
}

func clusterImport(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if nil != err {
		return err
	}

	resource, err := Lookup(c, ctx.Args().First(), "cluster")
	if nil != err {
		return err
	}

	cluster, err := getClusterByID(c, resource.ID)
	if nil != err {
		return err
	}

	if cluster.Driver != "" {
		return errors.New("existing k8s cluster can't be imported into this cluster")
	}

	clusterToken, err := getClusterRegToken(ctx, c, cluster.ID)
	if nil != err {
		return err
	}

	fmt.Printf("Run the following command in your cluster:\n%s\n\n%s\n%s\n", clusterToken.Command, importClusterNotice, clusterToken.InsecureCommand)

	return nil
}

// clusterAddNode prints the command needed to add a node to a cluster
func clusterAddNode(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if nil != err {
		return err
	}

	resource, err := Lookup(c, ctx.Args().First(), "cluster")
	if nil != err {
		return err
	}

	cluster, err := getClusterByID(c, resource.ID)
	if nil != err {
		return err
	}

	if cluster.Driver == "rancherKubernetesEngine" || cluster.Driver == "" {
		filter := defaultListOpts(ctx)
		filter.Filters["clusterId"] = cluster.ID
		nodePools, err := c.ManagementClient.NodePool.List(filter)
		if nil != err {
			return err
		}

		if len(nodePools.Data) > 0 {
			return errors.New("a node can't be added to the cluster this way")
		}
	} else {
		return errors.New("a node can't be added to the cluster this way")
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

func clusterDelete(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if nil != err {
		return err
	}

	for _, cluster := range ctx.Args() {

		resource, err := Lookup(c, cluster, "cluster")
		if nil != err {
			return err
		}

		cluster, err := getClusterByID(c, resource.ID)
		if nil != err {
			return err
		}

		err = c.ManagementClient.Cluster.Delete(cluster)
		if nil != err {
			return err
		}
	}

	return nil
}

func clusterExport(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if nil != err {
		return err
	}

	resource, err := Lookup(c, ctx.Args().First(), "cluster")
	if nil != err {
		return err
	}

	cluster, err := getClusterByID(c, resource.ID)
	if nil != err {
		return err
	}

	export, err := c.ManagementClient.Cluster.ActionExportYaml(cluster)
	if err != nil {
		return err
	}

	fmt.Println(export.YAMLOutput)
	return nil
}

func clusterKubeConfig(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if nil != err {
		return err
	}

	resource, err := Lookup(c, ctx.Args().First(), "cluster")
	if nil != err {
		return err
	}

	cluster, err := getClusterByID(c, resource.ID)
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

func addClusterMemberRoles(ctx *cli.Context) error {
	if len(ctx.Args()) < 2 {
		return cli.ShowSubcommandHelp(ctx)
	}

	memberName := ctx.Args().First()

	roles := ctx.Args()[1:]

	c, err := GetClient(ctx)
	if nil != err {
		return err
	}

	member, err := searchForMember(ctx, c, memberName)
	if nil != err {
		return err
	}

	clusterID := c.UserConfig.FocusedCluster()
	if ctx.String("cluster-id") != "" {
		clusterID = ctx.String("cluster-id")
	}

	for _, role := range roles {
		rtb := managementClient.ClusterRoleTemplateBinding{
			ClusterID:       clusterID,
			RoleTemplateID:  role,
			UserPrincipalID: member.ID,
		}
		if member.PrincipalType == "user" {
			rtb.UserPrincipalID = member.ID
		} else {
			rtb.GroupPrincipalID = member.ID
		}
		_, err = c.ManagementClient.ClusterRoleTemplateBinding.Create(&rtb)
		if nil != err {
			return err
		}
	}
	return nil
}

func deleteClusterMemberRoles(ctx *cli.Context) error {
	if len(ctx.Args()) < 2 {
		return cli.ShowSubcommandHelp(ctx)
	}

	memberName := ctx.Args().First()

	roles := ctx.Args()[1:]

	c, err := GetClient(ctx)
	if nil != err {
		return err
	}

	member, err := searchForMember(ctx, c, memberName)
	if nil != err {
		return err
	}

	clusterID := c.UserConfig.FocusedCluster()
	if ctx.String("cluster-id") != "" {
		clusterID = ctx.String("cluster-id")
	}

	for _, role := range roles {
		filter := defaultListOpts(ctx)
		filter.Filters["clusterId"] = clusterID
		filter.Filters["roleTemplateId"] = role

		if member.PrincipalType == "user" {
			filter.Filters["userPrincipalId"] = member.ID
		} else {
			filter.Filters["groupPrincipalId"] = member.ID
		}

		bindings, err := c.ManagementClient.ClusterRoleTemplateBinding.List(filter)
		if nil != err {
			return err
		}

		for _, binding := range bindings.Data {
			err = c.ManagementClient.ClusterRoleTemplateBinding.Delete(&binding)
			if nil != err {
				return err
			}
		}
	}
	return nil
}

func listClusterRoles(ctx *cli.Context) error {
	return listRoles(ctx, "cluster")
}

func listClusterMembers(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if nil != err {
		return err
	}

	clusterID := c.UserConfig.FocusedCluster()
	if ctx.String("cluster-id") != "" {
		clusterID = ctx.String("cluster-id")
	}

	filter := defaultListOpts(ctx)
	filter.Filters["clusterId"] = clusterID
	bindings, err := c.ManagementClient.ClusterRoleTemplateBinding.List(filter)
	if nil != err {
		return err
	}

	userFilter := defaultListOpts(ctx)
	users, err := c.ManagementClient.User.List(userFilter)
	if nil != err {
		return err
	}

	userMap := usersToNameMapping(users.Data)

	var b []RoleTemplateBinding

	for _, binding := range bindings.Data {
		parsedTime, err := createdTimetoHuman(binding.Created)
		if nil != err {
			return err
		}

		b = append(b, RoleTemplateBinding{
			ID:      binding.ID,
			User:    userMap[binding.UserID],
			Role:    binding.RoleTemplateID,
			Created: parsedTime,
		})
	}

	return listRoleTemplateBindings(ctx, b)
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
			ClusterID: clusterID,
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

func getClusterProvider(cluster managementClient.Cluster) string {
	switch cluster.Driver {
	case "imported":
		return "Imported"
	case "rancherKubernetesEngine":
		return "Rancher Kubernetes Engine"
	case "azureKubernetesService":
		return "Azure Container Service"
	case "googleKubernetesEngine":
		return "Google Kubernetes Engine"
	default:
		return "Unknown"
	}
}

func getClusterNodeCount(ctx *cli.Context, c *cliclient.MasterClient, clusterID string) (int, error) {
	nodes, err := getNodesList(ctx, c, clusterID)
	if err != nil {
		return 0, err
	}
	return len(nodes.Data), nil
}

func getClusterCPU(cluster managementClient.Cluster) string {
	cpu, err := strconv.ParseFloat(strings.Replace(cluster.Requested["cpu"], "m", "", -1), 64)
	if nil != err {
		fmt.Println(err)
	}
	cpu = cpu / 1000
	cpu2 := strconv.FormatFloat(cpu, 'f', 2, 32)
	return cpu2 + "/" + cluster.Allocatable["cpu"]
}

func getClusterRAM(cluster managementClient.Cluster) string {
	allo := strings.Replace(cluster.Allocatable["memory"], "Ki", "", -1)
	alloInt, err := strconv.ParseFloat(allo, 64)
	if nil != err {
		fmt.Println(err)
	}
	alloInt = alloInt / 1024 / 1024
	alloString := fmt.Sprintf("%.2f", alloInt)
	alloString = strings.TrimSuffix(alloString, ".0")

	requested := strings.Replace(cluster.Requested["memory"], "Mi", "", -1)
	requestedInt, err := strconv.ParseFloat(requested, 64)
	if nil != err {
		fmt.Println(err)
	}
	requestedInt = requestedInt / 1024
	requestedString := fmt.Sprintf("%.2f", requestedInt)
	requestedString = strings.TrimSuffix(requestedString, ".0")

	return requestedString + "/" + alloString + " GB"
}

func getClusterPods(cluster managementClient.Cluster) string {
	return cluster.Requested["pods"] + "/" + cluster.Allocatable["pods"]
}

func getClusterK8sOptions(c *cliclient.MasterClient) []string {
	var options []string
	setting, _ := c.ManagementClient.Setting.ByID("k8s-version-to-images")
	var objmap map[string]*json.RawMessage

	json.Unmarshal([]byte(setting.Value), &objmap)
	for key := range objmap {
		options = append(options, key)
	}
	return options
}

func getRKEConfig(ctx *cli.Context) (*managementClient.RancherKubernetesEngineConfig, error) {
	if ctx.Bool("import") {
		return nil, nil
	}

	rkeConfig := &managementClient.RancherKubernetesEngineConfig{}

	if ctx.String("rke-config") != "" {
		bytes, err := readFileReturnJSON(ctx.String("rke-config"))
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bytes, &rkeConfig)
		if err != nil {
			return nil, err
		}
	}

	rkeConfig.IgnoreDockerVersion = ctx.BoolT("disable-docker-version")

	if ctx.String("k8s-version") != "" {
		rkeConfig.Version = ctx.String("k8s-version")
	}

	if ctx.String("network-provider") != "" {
		rkeConfig.Network = &managementClient.NetworkConfig{
			Plugin: ctx.String("network-provider"),
		}
	}

	if ctx.String("psp-default-policy") != "" {
		rkeConfig.Services = &managementClient.RKEConfigServices{
			KubeAPI: &managementClient.KubeAPIService{
				PodSecurityPolicy: true,
			},
		}
	}

	return rkeConfig, nil
}

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/rancher/cli/cliclient"
	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
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
	Nodes    int64
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
				Description: "Create a new custom cluster with desired configuration",
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
				Flags: []cli.Flag{
					quietFlag,
				},
			},
			{
				Name:      "add-node",
				Usage:     "Outputs the docker command needed to add a node to an existing Rancher custom cluster",
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
						Usage: "Use node for management (DEPRECATED, use controlplane instead)",
					},
					cli.BoolFlag{
						Name:  "controlplane",
						Usage: "Use node for controlplane",
					},
					cli.BoolFlag{
						Name:  "worker",
						Usage: "Use node as a worker",
					},
					quietFlag,
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

		writer.Write(&ClusterData{
			ID:       item.ID,
			Current:  current,
			Cluster:  item,
			Name:     getClusterName(&item),
			Provider: getClusterProvider(item),
			Nodes:    item.NodeCount,
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
	if err != nil {
		return err
	}

	k8sVersion := ctx.String("k8s-version")
	if k8sVersion != "" {
		k8sVersions, err := getClusterK8sOptions(c)
		if err != nil {
			return err
		}

		if slices.Contains(k8sVersions, k8sVersion) {
			fmt.Println("Available Kubernetes versions:")
			for _, val := range k8sVersions {
				fmt.Println(val)
			}
			return nil
		}
	}

	config, err := getClusterConfig(ctx)
	if err != nil {
		return err
	}

	createdCluster, err := c.ManagementClient.Cluster.Create(config)
	if err != nil {
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
	if err != nil {
		return err
	}

	resource, err := Lookup(c, ctx.Args().First(), "cluster")
	if err != nil {
		return err
	}

	cluster, err := getClusterByID(c, resource.ID)
	if err != nil {
		return err
	}

	if cluster.Driver != "" {
		return errors.New("existing k8s cluster can't be imported into this cluster")
	}

	clusterToken, err := getClusterRegToken(ctx, c, cluster.ID)
	if err != nil {
		return err
	}

	if ctx.Bool("quiet") {
		fmt.Println(clusterToken.Command)
		fmt.Println(clusterToken.InsecureCommand)
		return nil
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
	if err != nil {
		return err
	}

	resource, err := Lookup(c, ctx.Args().First(), "cluster")
	if err != nil {
		return err
	}

	cluster, err := getClusterByID(c, resource.ID)
	if err != nil {
		return err
	}

	if cluster.Driver == "rancherKubernetesEngine" || cluster.Driver == "" {
		filter := defaultListOpts(ctx)
		filter.Filters["clusterId"] = cluster.ID
		nodePools, err := c.ManagementClient.NodePool.List(filter)
		if err != nil {
			return err
		}

		if len(nodePools.Data) > 0 {
			return errors.New("a node can't be manually registered to a cluster utilizing node-pools")
		}
	} else {
		return errors.New("a node can only be manually registered to a custom cluster")
	}

	clusterToken, err := getClusterRegToken(ctx, c, cluster.ID)
	if err != nil {
		return err
	}

	var roleFlags string

	if ctx.Bool("etcd") {
		roleFlags = roleFlags + " --etcd"
	}

	if ctx.Bool("management") || ctx.Bool("controlplane") {
		if ctx.Bool("management") && !ctx.Bool("quiet") {
			logrus.Info("The flag --management is deprecated and replaced by --controlplane")
		}
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

	if ctx.Bool("quiet") {
		fmt.Println(command)
		return nil
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
	if err != nil {
		return err
	}

	for _, cluster := range ctx.Args() {

		resource, err := Lookup(c, cluster, "cluster")
		if err != nil {
			return err
		}

		cluster, err := getClusterByID(c, resource.ID)
		if err != nil {
			return err
		}

		err = c.ManagementClient.Cluster.Delete(cluster)
		if err != nil {
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
	if err != nil {
		return err
	}

	resource, err := Lookup(c, ctx.Args().First(), "cluster")
	if err != nil {
		return err
	}

	cluster, err := getClusterByID(c, resource.ID)
	if err != nil {
		return err
	}

	if _, ok := cluster.Actions["exportYaml"]; !ok {
		return errors.New("cluster does not support being exported")
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
	if err != nil {
		return err
	}

	resource, err := Lookup(c, ctx.Args().First(), "cluster")
	if err != nil {
		return err
	}

	cluster, err := getClusterByID(c, resource.ID)
	if err != nil {
		return err
	}

	config, err := c.ManagementClient.Cluster.ActionGenerateKubeconfig(cluster)
	if err != nil {
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
	if err != nil {
		return err
	}

	member, err := searchForMember(ctx, c, memberName)
	if err != nil {
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
		if err != nil {
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
	if err != nil {
		return err
	}

	member, err := searchForMember(ctx, c, memberName)
	if err != nil {
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
		if err != nil {
			return err
		}

		for _, binding := range bindings.Data {
			err = c.ManagementClient.ClusterRoleTemplateBinding.Delete(&binding)
			if err != nil {
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
	if err != nil {
		return err
	}

	clusterID := c.UserConfig.FocusedCluster()
	if ctx.String("cluster-id") != "" {
		clusterID = ctx.String("cluster-id")
	}

	filter := defaultListOpts(ctx)
	filter.Filters["clusterId"] = clusterID
	bindings, err := c.ManagementClient.ClusterRoleTemplateBinding.List(filter)
	if err != nil {
		return err
	}

	userFilter := defaultListOpts(ctx)
	users, err := c.ManagementClient.User.List(userFilter)
	if err != nil {
		return err
	}

	userMap := usersToNameMapping(users.Data)

	var b []RoleTemplateBinding

	for _, binding := range bindings.Data {
		parsedTime, err := createdTimetoHuman(binding.Created)
		if err != nil {
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
	if err != nil {
		return managementClient.ClusterRegistrationToken{}, err
	}

	if len(clusterTokenCollection.Data) == 0 {
		crt := &managementClient.ClusterRegistrationToken{
			ClusterID: clusterID,
		}
		clusterToken, err := c.ManagementClient.ClusterRegistrationToken.Create(crt)
		if err != nil {
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
	if err != nil {
		return nil, fmt.Errorf("no cluster found with the ID [%s], run "+
			"`rancher clusters` to see available clusters: %s", clusterID, err)
	}
	return cluster, nil
}

func getClusterProvider(cluster managementClient.Cluster) string {
	switch cluster.Driver {
	case "imported":
		switch cluster.Provider {
		case "rke2":
			return "RKE2"
		case "k3s":
			return "K3S"
		default:
			return "Imported"
		}
	case "k3s":
		return "K3S"
	case "rke2":
		return "RKE2"
	case "rancherKubernetesEngine":
		return "Rancher Kubernetes Engine"
	case "azureKubernetesService", "AKS":
		return "Azure Kubernetes Service"
	case "googleKubernetesEngine", "GKE":
		return "Google Kubernetes Engine"
	case "EKS":
		return "Elastic Kubernetes Service"
	default:
		return "Unknown"
	}
}

func getClusterCPU(cluster managementClient.Cluster) string {
	req := parseResourceString(cluster.Requested["cpu"])
	alloc := parseResourceString(cluster.Allocatable["cpu"])
	return req + "/" + alloc
}

func getClusterRAM(cluster managementClient.Cluster) string {
	req := parseResourceString(cluster.Requested["memory"])
	alloc := parseResourceString(cluster.Allocatable["memory"])
	return req + "/" + alloc + " GB"
}

// parseResourceString returns GB for Ki and Mi and CPU cores from 'm'
func parseResourceString(mem string) string {
	if strings.HasSuffix(mem, "Ki") {
		num, err := strconv.ParseFloat(strings.Replace(mem, "Ki", "", -1), 64)
		if err != nil {
			return mem
		}
		num = num / 1024 / 1024
		return strings.TrimSuffix(fmt.Sprintf("%.2f", num), ".0")
	}
	if strings.HasSuffix(mem, "Mi") {
		num, err := strconv.ParseFloat(strings.Replace(mem, "Mi", "", -1), 64)
		if err != nil {
			return mem
		}
		num = num / 1024
		return strings.TrimSuffix(fmt.Sprintf("%.2f", num), ".0")
	}
	if strings.HasSuffix(mem, "m") {
		num, err := strconv.ParseFloat(strings.Replace(mem, "m", "", -1), 64)
		if err != nil {
			return mem
		}
		num = num / 1000
		return strconv.FormatFloat(num, 'f', 2, 32)
	}
	return mem
}

func getClusterPods(cluster managementClient.Cluster) string {
	return cluster.Requested["pods"] + "/" + cluster.Allocatable["pods"]
}

func getClusterK8sOptions(c *cliclient.MasterClient) ([]string, error) {
	var options []string

	setting, err := c.ManagementClient.Setting.ByID("k8s-version-to-images")
	if err != nil {
		return nil, err
	}

	var objmap map[string]*json.RawMessage
	err = json.Unmarshal([]byte(setting.Value), &objmap)
	if err != nil {
		return nil, err
	}

	for key := range objmap {
		options = append(options, key)
	}
	return options, nil
}

func getClusterConfig(ctx *cli.Context) (*managementClient.Cluster, error) {
	config := managementClient.Cluster{}
	config.Name = ctx.Args().First()
	config.Description = ctx.String("description")

	if !ctx.Bool("import") {
		config.RancherKubernetesEngineConfig = new(managementClient.RancherKubernetesEngineConfig)
		ignoreDockerVersion := ctx.BoolT("disable-docker-version")
		config.RancherKubernetesEngineConfig.IgnoreDockerVersion = &ignoreDockerVersion

		if ctx.String("k8s-version") != "" {
			config.RancherKubernetesEngineConfig.Version = ctx.String("k8s-version")
		}

		if ctx.String("network-provider") != "" {
			config.RancherKubernetesEngineConfig.Network = &managementClient.NetworkConfig{
				Plugin: ctx.String("network-provider"),
			}
		}

		if ctx.String("rke-config") != "" {
			bytes, err := readFileReturnJSON(ctx.String("rke-config"))
			if err != nil {
				return nil, err
			}

			var jsonObject map[string]interface{}
			if err = json.Unmarshal(bytes, &jsonObject); err != nil {
				return nil, err
			}

			// Most values in RancherKubernetesEngineConfig are defined with struct tags for both JSON and YAML in camelCase.
			// Changing the tags will be a breaking change. For proper deserialization, we must convert all keys to camelCase.
			// Note that we ignore kebab-case keys. Users themselves should ensure any relevant keys
			// (especially top-level keys in `services`, like `kube-api` or `kube-controller`) are camelCase or snake-case in cluster config.
			convertSnakeCaseKeysToCamelCase(jsonObject)

			marshalled, err := json.Marshal(jsonObject)
			if err != nil {
				return nil, err
			}
			if err = json.Unmarshal(marshalled, &config); err != nil {
				return nil, err
			}
		}
	}

	if ctx.String("psp-default-policy") != "" {
		config.DefaultPodSecurityPolicyTemplateID = ctx.String("psp-default-policy")

		if config.RancherKubernetesEngineConfig != nil {
			config.RancherKubernetesEngineConfig.Services = &managementClient.RKEConfigServices{
				KubeAPI: &managementClient.KubeAPIService{
					PodSecurityPolicy: true,
				},
			}
		}
	}

	return &config, nil
}

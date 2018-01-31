package cmd

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"text/template"

	"github.com/rancher/cli/cliclient"
	managementClient "github.com/rancher/types/client/management/v3"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	importDescription = `
Imports an existing cluster to be used in rancher either by providing the location
to .kube/config or using a generated kubectl command to run in your cluster.
`
	dockerCommandTemplate = "docker run -d --restart=unless-stopped " +
		"-v /var/run/docker.sock:/var/run/docker.sock --net=host " +
		"{{.Image}} {{range .RoleFlags}}{{.}}{{end}}--server {{.URL}} " +
		"--token {{.Token}} --ca-checksum {{.Checksum}}\n"
)

type dockerCommand struct {
	Checksum  string
	Image     string
	RoleFlags []string
	Token     string
	URL       string
}

type ClusterData struct {
	Cluster managementClient.Cluster
}

func ClusterCommand() cli.Command {
	clusterFileFlag := cli.StringFlag{
		Name:  "file, f",
		Usage: "Location of file to load",
	}

	return cli.Command{
		Name:    "clusters",
		Aliases: []string{"cluster"},
		Usage:   "Operations on clusters",
		Action:  defaultAction(clusterLs),
		Subcommands: []cli.Command{
			{
				Name:        "ls",
				Usage:       "List clusters",
				Description: "\nLists all clusters in the current cluster.",
				ArgsUsage:   "None",
				Action:      clusterLs,
			},
			// FIXME add this back in along with the required flags
			//{
			//	Name:        "create",
			//	Usage:       "create `NAME`",
			//	Description: "\nCreates a cluster on the server",
			//	ArgsUsage:   "[NEWCLUSTERNAME...]",
			//	Action:      clusterCreate,
			//	Flags: []cli.Flag{
			//		clusterFileFlag,
			//		cli.StringFlag{
			//			Name:  "type",
			//			Usage: "type of cluster to create",
			//		},
			//	},
			//},
			{
				Name:        "import",
				Usage:       "Import an existing cluster",
				Description: importDescription,
				ArgsUsage:   "[NEWCLUSTERNAME...]",
				Action:      clusterImport,
				Flags: []cli.Flag{
					clusterFileFlag,
				},
			},
			{
				Name:      "get-command",
				Usage:     "Returns the command needed to add a node to an existing cluster",
				ArgsUsage: "[CLUSTERNAME]",
				Action:    getDockerCommand,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "label",
						Usage: "Labels to apply to a node",
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
	c, err := GetClient(ctx)
	if nil != err {
		return err
	}

	var name string

	if ctx.NArg() > 0 {
		name = ctx.Args()[0]
	}

	_, err = c.ManagementClient.Cluster.Create(&managementClient.Cluster{
		Name: name,
	})

	if nil != err {
		return err
	}

	return nil
}

func clusterImport(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return errors.New("name is required")
	}

	if ctx.String("file") != "" {
		err := clusterFromKubeconfig(ctx)
		if nil != err {
			return err
		}
		return nil
	}

	err := clusterFromCommand(ctx)
	if nil != err {
		return err
	}

	return nil
}

// getDockerCommand prints the command needed to add a node to a cluster
func getDockerCommand(ctx *cli.Context) error {
	var clusterName string

	if ctx.NArg() == 0 {
		return errors.New("cluster name is required")
	}

	clusterName = ctx.Args().First()

	c, err := GetClient(ctx)
	if nil != err {
		return err
	}

	settingsMap, err := settingsToMap(c)
	if nil != err {
		return err
	}

	opts := defaultListOpts(ctx)
	opts.Filters["name"] = clusterName

	clusterCollection, err := c.ManagementClient.Cluster.List(opts)
	if nil != err {
		return err
	}

	if len(clusterCollection.Data) == 0 {
		return fmt.Errorf("no cluster found with the name [%s], run "+
			"`rancher clusters` to see available clusters", clusterName)
	}

	clusterToken, err := getClusterRegToken(ctx, c, clusterCollection.Data[0].ID)
	if nil != err {
		return err
	}

	var roleFlags []string

	if ctx.Bool("etcd") {
		roleFlags = append(roleFlags, "--etcd ")
	}

	if ctx.Bool("management") {
		roleFlags = append(roleFlags, "--controlplane ")
	}

	if ctx.Bool("worker") {
		roleFlags = append(roleFlags, "--worker ")
	}

	dockerString, err := template.New("docker").Parse(dockerCommandTemplate)
	if nil != err {
		return err
	}

	dockerInfo := dockerCommand{
		Checksum:  checkSum(settingsMap["cacerts"] + "\n"),
		Image:     settingsMap["agent-image"],
		RoleFlags: roleFlags,
		Token:     clusterToken.Token,
		URL:       c.UserConfig.URL,
	}

	fmt.Println("Run this command on an existing machine already running a " +
		"supported version of Docker:")

	dockerString.Execute(os.Stdout, dockerInfo)

	return nil
}

// clusterFromKubeconfig reads in a JSON or yaml of the kubeconfig and uses
// that to pull in the cluster
func clusterFromKubeconfig(ctx *cli.Context) error {
	blob, err := readFileReturnJSON(ctx.String("file"))
	if err != nil {
		return err
	}

	c, err := GetClient(ctx)
	if nil != err {
		return err
	}

	cluster, err := c.ManagementClient.Cluster.Create(&managementClient.Cluster{
		Name:           ctx.Args().First(),
		ImportedConfig: &managementClient.ImportedConfig{KubeConfig: string(blob)},
	})
	if nil != err {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"Name": cluster.Name,
		"ID":   cluster.ID,
	}).Info("Cluster created:")
	return nil
}

// clusterFromCommand creates a holder cluster and provides the command to run
// in the cluster to register with Rancher
func clusterFromCommand(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if nil != err {
		return err
	}

	// create a holder cluster so we can get the ClusterRegistrationToken
	cluster, err := c.ManagementClient.Cluster.Create(&managementClient.Cluster{
		Name: ctx.Args().First(),
		RancherKubernetesEngineConfig: &managementClient.RancherKubernetesEngineConfig{
			Nodes: make([]managementClient.RKEConfigNode, 1),
		},
	})
	if nil != err {
		return err
	}

	token, err := getClusterRegToken(ctx, c, cluster.ID)
	if nil != err {
		return err
	}

	//FIXME probably need more info here
	logrus.Printf("Run the following command in your cluster: %v", token.Command)
	return nil
}

func checkSum(s string) string {
	sum := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", sum)
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

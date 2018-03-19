package cmd

import (
	"io/ioutil"
	"strings"

	"github.com/rancher/types/client/management/v3"
	projectClient "github.com/rancher/types/client/project/v3"
	"github.com/urfave/cli"
)

func UpCommand() cli.Command {
	return cli.Command{
		Name:   "up",
		Usage:  "apply compose config",
		Action: defaultAction(apply),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "global,g",
				Usage: "Apply global-scoped config",
			},
			cli.BoolFlag{
				Name:  "cluster,c",
				Usage: "Apply cluster-scoped config. Will apply config into the current cluster",
			},
			cli.StringFlag{
				Name:  "namespace,n",
				Usage: "Apply namespace-scoped config to the specified namespace",
			},
			cli.StringFlag{
				Name:  "file,f",
				Usage: "The location of compose config file",
			},
		},
	}
}

func apply(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	clusterID := strings.Split(c.UserConfig.Project, ":")[0]

	filePath := ctx.String("file")
	compose, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	if ctx.Bool("global") {
		globalComposeConfig := &client.GlobalComposeConfig{
			RancherCompose: string(compose),
		}
		if _, err := c.ManagementClient.GlobalComposeConfig.Create(globalComposeConfig); err != nil {
			return err
		}
	}

	if ctx.Bool("cluster") {
		clusterComposeConfig := &client.ClusterComposeConfig{
			RancherCompose: string(compose),
			ClusterId:      clusterID,
		}
		if _, err := c.ManagementClient.ClusterComposeConfig.Create(clusterComposeConfig); err != nil {
			return err
		}
	}

	if namespace := ctx.String("namespace"); namespace != "" {
		namespaceComposeConfig := &projectClient.NamespaceComposeConfig{
			RancherCompose:   string(compose),
			InstallNamespace: namespace,
		}
		if _, err := c.ProjectClient.NamespaceComposeConfig.Create(namespaceComposeConfig); err != nil {
			return err
		}
	}
	return nil
}

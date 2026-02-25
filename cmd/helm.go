package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/urfave/cli"
	"k8s.io/client-go/tools/clientcmd"
)

func HelmCommand() cli.Command {
	return cli.Command{
		Name:            "helm",
		Usage:           "Run helm commands",
		Description:     "Use the current cluster context to run helm commands in the cluster",
		Action:          runHelm,
		SkipFlagParsing: true,
	}
}

func runHelm(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		return cli.ShowCommandHelp(ctx, "helm")
	}

	path, err := exec.LookPath("helm")
	if err != nil {
		return fmt.Errorf("helm is required to be set in your path to use this "+
			"command. See https://helm.sh/docs/intro/install/ "+
			"for more info. Error: %s", err.Error())
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	config, err := loadConfig(ctx)
	if err != nil {
		return err
	}

	currentRancherServer, err := config.FocusedServer()
	if err != nil {
		return err
	}

	currentToken := currentRancherServer.AccessKey
	t, err := c.ManagementClient.Token.ByID(currentToken)
	if err != nil {
		return err
	}

	currentUser := t.UserID
	kubeConfig, err := getKubeConfigForUser(ctx, currentUser)
	if err != nil {
		return err
	}

	var isTokenValid bool
	if kubeConfig != nil {
		tokenID, err := extractKubeconfigTokenID(*kubeConfig)
		if err != nil {
			return err
		}
		isTokenValid, err = validateToken(tokenID, c.ManagementClient.Token)
		if err != nil {
			return err
		}
	}

	if kubeConfig == nil || !isTokenValid {
		cluster, err := getClusterByID(c, c.UserConfig.FocusedCluster())
		if err != nil {
			return err
		}

		config, err := c.ManagementClient.Cluster.ActionGenerateKubeconfig(cluster)
		if err != nil {
			return err
		}

		kubeConfigBytes := []byte(config.Config)
		kubeConfig, err = clientcmd.Load(kubeConfigBytes)
		if err != nil {
			return err
		}

		if err := setKubeConfigForUser(ctx, currentUser, kubeConfig); err != nil {
			return err
		}
	}

	tmpfile, err := os.CreateTemp("", "rancher-")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())

	if err := clientcmd.WriteToFile(*kubeConfig, tmpfile.Name()); err != nil {
		return err
	}
	if err := tmpfile.Close(); err != nil {
		return err
	}

	cmd := exec.Command(path, ctx.Args()...)
	cmd.Env = append(os.Environ(), "KUBECONFIG="+tmpfile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

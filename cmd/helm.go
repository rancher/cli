package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/urfave/cli"
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
	if nil != err {
		return fmt.Errorf("helm is required to be set in your path to use this "+
			"command. See https://docs.helm.sh/using_helm/#installing-helm "+
			"for more info. Error: %s", err.Error())
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	cluster, err := getClusterByID(c, c.UserConfig.FocusedCluster())
	if nil != err {
		return err
	}

	config, err := c.ManagementClient.Cluster.ActionGenerateKubeconfig(cluster)
	if nil != err {
		return err
	}

	tmpfile, err := ioutil.TempFile("", "rancher-")
	if nil != err {
		return err
	}
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(config.Config))
	if nil != err {
		return err
	}

	err = tmpfile.Close()
	if nil != err {
		return err
	}

	cmd := exec.Command(path, ctx.Args()...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = append(os.Environ(),
		"KUBECONFIG="+tmpfile.Name(),
	)
	err = cmd.Run()
	if nil != err {
		return err
	}
	return nil
}

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/urfave/cli"
)

func KubectlCommand() cli.Command {
	return cli.Command{
		Name:            "kubectl",
		Usage:           "Run kubectl commands",
		Description:     "Use the current cluster context to run kubectl commands in the cluster",
		Action:          runKubectl,
		SkipFlagParsing: true,
	}
}

func runKubectl(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		return cli.ShowCommandHelp(ctx, "kubectl")
	}

	path, err := exec.LookPath("kubectl")
	if err != nil {
		return fmt.Errorf("kubectl is required to be set in your path to use this "+
			"command. See https://kubernetes.io/docs/tasks/tools/install-kubectl/ "+
			"for more info. Error: %s", err.Error())
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	cluster, err := getClusterByID(c, c.UserConfig.FocusedCluster())
	if err != nil {
		return err
	}

	config, err := c.ManagementClient.Cluster.ActionGenerateKubeconfig(cluster)
	if err != nil {
		return err
	}

	tmpfile, err := ioutil.TempFile("", "rancher-")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(config.Config))
	if err != nil {
		return err
	}

	err = tmpfile.Close()
	if err != nil {
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

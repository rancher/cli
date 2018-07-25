package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/urfave/cli"
)

func KomposeCommand() cli.Command {
	return cli.Command{
		Name:            "kompose",
		Usage:           "Run kompose commands",
		Description:     "Use the current cluster context to run kompose commands in the cluster",
		Action:          runKompose,
		SkipFlagParsing: true,
	}
}

func runKompose(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		return cli.ShowCommandHelp(ctx, "kompose")
	}

	path, err := exec.LookPath("kompose")
	if nil != err {
		return fmt.Errorf("kompose is required to be set in your path to use this "+
			"command. See https://github.com/kubernetes/kompose/blob/master/docs/installation.md "+
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

	kubeLocationEnvVar := "KUBECONFIG=" + tmpfile.Name()

	combinedArgs := ctx.Args()

	cmd := exec.Command(path, combinedArgs...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, kubeLocationEnvVar)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if nil != err {
		return err
	}
	return nil
}

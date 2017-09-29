package cmd

import (
	"fmt"

	"io/ioutil"
	"os"
	"os/exec"

	"github.com/rancher/go-rancher/v3"
	"github.com/urfave/cli"
)

func KubectlCommand() cli.Command {
	return cli.Command{
		Name:        "kubectl",
		Usage:       "Run Kubectl on a k8s cluster in rancher",
		Description: "\nRun Kubectl on rancher cluster. Example: 'rancher kubectl get pod'\nTo specify a cluster, run `rancher --cluster 1c1 kubectl get pod`\n",
		Action:      kubectl,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "help-kubectl",
				Usage: "Display the 'kubectl --help'",
			},
		},
		SkipFlagParsing: true,
	}
}

func kubectl(ctx *cli.Context) error {
	return processExitCode(doKubectl(ctx))
}

func doKubectl(ctx *cli.Context) error {
	c, err := GetRawClient(ctx)
	if err != nil {
		return err
	}
	args := ctx.Args()
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		return cli.ShowCommandHelp(ctx, "kubectl")
	}

	if len(args) > 0 && args[0] == "--help-kubectl" {
		return runKubectlHelp("")
	}
	filePath, err := generateKubeconfigFile(c, ctx)
	if err != nil {
		if filePath != "" {
			os.RemoveAll(filePath)
		}
		return err
	}
	defer os.RemoveAll(filePath)

	commandArgs := append([]string{"--kubeconfig", filePath}, ctx.Args()...)
	command := exec.Command("kubectl", commandArgs...)
	command.Env = os.Environ()
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}

func generateKubeconfigFile(c *client.RancherClient, ctx *cli.Context) (string, error) {
	clusterID := ""
	clusterName := ctx.GlobalString("cluster")
	if clusterName != "" {
		cluster, err := Lookup(c, clusterName, "cluster")
		if err != nil {
			return "", err
		}
		clusterID = cluster.Id
	}

	config, err := lookupConfig(ctx)
	if err != nil {
		return "", err
	}

	env, err := c.Project.ById(config.Environment)
	if err != nil {
		return "", err
	}

	if clusterID == "" {
		clusterID = env.ClusterId
	}

	baseURL, err := baseURL(config.URL)
	if err != nil {
		return "", err
	}

	serverAddress := fmt.Sprintf("%s/k8s/clusters/%s", baseURL, clusterID)

	configTemplate := `apiVersion: v1
kind: Config
clusters:
- cluster:
    api-version: v1
    insecure-skip-tls-verify: true
    server: "%s"
  name: "%s"
contexts:
- context:
    cluster: "%s"
    user: "%s"
  name: "%s"
current-context: "%s"
users:
- name: "%s"
  user:
    username: "%s"
    password: "%s"`

	kubeConfig := fmt.Sprintf(configTemplate, serverAddress, env.Name, env.Name, env.Name, env.Name, env.Name, env.Name, config.AccessKey, config.SecretKey)

	tempfile, err := ioutil.TempFile("", "kube-config")
	if err != nil {
		return "", err
	}

	_, err = tempfile.Write([]byte(kubeConfig))
	if err != nil {
		return tempfile.Name(), err
	}

	if err := tempfile.Close(); err != nil {
		return tempfile.Name(), err
	}

	return tempfile.Name(), nil
}

func runKubectlHelp(subcommand string) error {
	args := []string{"--help"}
	if subcommand != "" {
		args = []string{subcommand, "--help"}
	}
	cmd := exec.Command("kubectl", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

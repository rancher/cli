package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	extv1 "github.com/rancher/rancher/pkg/apis/ext.cattle.io/v1"
	"github.com/urfave/cli/v3"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func KubectlCommand() *cli.Command {
	return &cli.Command{
		Name:            "kubectl",
		Usage:           "Run kubectl commands",
		Description:     "Use the current cluster context to run kubectl commands in the cluster",
		Action:          runKubectl,
		SkipFlagParsing: true,
	}
}

func runKubectl(ctx context.Context, cmd *cli.Command) error {
	args := cmd.Args().Slice()
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		return cli.ShowCommandHelp(ctx, cmd, "kubectl")
	}

	path, err := exec.LookPath("kubectl")
	if err != nil {
		return fmt.Errorf("kubectl is required to be set in your path to use this "+
			"command. See https://kubernetes.io/docs/tasks/tools/install-kubectl/ "+
			"for more info. Error: %s", err.Error())
	}

	c, err := GetClient(cmd)
	if err != nil {
		return err
	}

	config, err := loadConfig(cmd)
	if err != nil {
		return err
	}

	currentRancherServer, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	currentToken := currentRancherServer.AccessKey
	// bearerToken intentionally keeps any "ext/" prefix on AccessKey because
	// the ext API authenticator parses the prefix back out of the Bearer header.
	bearerToken := currentRancherServer.AccessKey + ":" + currentRancherServer.SecretKey

	// Norman's MasterClient does not expose its underlying http.Client, so
	// build a parallel one here for the direct ext API call.
	tlsConf, err := getTLSConfig(false, currentRancherServer.CACerts)
	if err != nil {
		return fmt.Errorf("error creating TLS config: %w", err)
	}
	httpClient, err := newHTTPClient(currentRancherServer, tlsConf)
	if err != nil {
		return fmt.Errorf("error creating HTTP client: %w", err)
	}
	baseURL, err := currentRancherServer.EnvironmentURL()
	if err != nil {
		return fmt.Errorf("error resolving server base URL: %w", err)
	}
	extGetter := func(ctx context.Context, id string) (*extv1.Token, error) {
		return getExtToken(ctx, id, baseURL, bearerToken, httpClient)
	}
	v3ByID := c.ManagementClient.Token.ByID

	currentUser, err := getTokenUserID(ctx, currentToken, v3ByID, extGetter)
	if err != nil {
		return err
	}
	kubeConfig, err := getKubeConfigForUser(cmd, currentUser)
	if err != nil {
		return err
	}

	var isTokenValid bool
	if kubeConfig != nil {
		tokenID, err := extractKubeconfigTokenID(*kubeConfig)
		if err != nil {
			return err
		}
		isTokenValid, err = validateToken(ctx, tokenID, v3ByID, extGetter)
		if err != nil {
			return err
		}
	}

	if kubeConfig == nil || !isTokenValid {
		cluster, err := getClusterByID(c, c.UserConfig.GetCurrentCluster())
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

		if err := setKubeConfigForUser(cmd, currentUser, kubeConfig); err != nil {
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

	execCmd := exec.Command(path, cmd.Args().Slice()...)
	execCmd.Env = append(os.Environ(), "KUBECONFIG="+tmpfile.Name())
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	execCmd.Stdin = os.Stdin
	err = execCmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func extractKubeconfigTokenID(kubeconfig api.Config) (string, error) {
	if len(kubeconfig.AuthInfos) != 1 {
		return "", fmt.Errorf("invalid kubeconfig, expected to contain exactly 1 user")
	}
	var parts []string
	for _, val := range kubeconfig.AuthInfos {
		parts = strings.Split(val.Token, ":")
		if len(parts) != 2 {
			return "", fmt.Errorf("failed to parse kubeconfig token")
		}
	}

	return parts[0], nil
}


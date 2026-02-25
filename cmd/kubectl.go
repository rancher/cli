package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/rancher/norman/clientbase"
	client "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"github.com/urfave/cli"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// clusterIDRegexp matches valid cluster ID formats: c-xxxxx or c-m-xxxxxxxx
var clusterIDRegexp = regexp.MustCompile(`^(c-[[:alnum:]]{5})|(c-m-[[:alnum:]]{8})$`)

func KubectlCommand() cli.Command {
	return cli.Command{
		Name:  "kubectl",
		Usage: "Run kubectl commands",
		Description: `Use the current cluster context to run kubectl commands in the cluster.

Use --context to run commands in a different cluster without switching:
    rancher kubectl --context c-xxxxx get pods`,
		Action:          runKubectl,
		SkipFlagParsing: true,
	}
}

// extractContextFlag parses and removes the --context flag from args.
// Returns the context value (empty if not provided) and the remaining args for kubectl.
func extractContextFlag(args cli.Args) (contextOverride string, kubectlArgs []string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--context" && i+1 < len(args) {
			contextOverride = args[i+1]
			i++ // Skip the value
		} else if val, found := strings.CutPrefix(arg, "--context="); found {
			contextOverride = val
		} else {
			kubectlArgs = append(kubectlArgs, arg)
		}
	}
	return contextOverride, kubectlArgs
}

// parseContextOverride validates and returns a cluster ID.
// Accepts formats: "local", "c-xxxxx", or "c-m-xxxxxxxx".
func parseContextOverride(context string) (string, error) {
	if context == "" {
		return "", fmt.Errorf("context cannot be empty")
	}

	if context == "local" {
		return context, nil
	}
	if clusterIDRegexp.MatchString(context) {
		return context, nil
	}

	return "", fmt.Errorf("invalid cluster ID format %q: expected local, c-xxxxx, or c-m-xxxxxxxx", context)
}

func runKubectl(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		return cli.ShowCommandHelp(ctx, "kubectl")
	}

	// Extract --context flag from args before passing to kubectl
	contextOverride, kubectlArgs := extractContextFlag(args)

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

	// Determine target cluster ID
	var targetClusterID string
	if contextOverride != "" {
		targetClusterID, err = parseContextOverride(contextOverride)
		if err != nil {
			return err
		}
	} else {
		targetClusterID = c.UserConfig.FocusedCluster()
	}

	kubeConfig, err := getKubeConfigForUserAndCluster(ctx, currentUser, targetClusterID)
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
		cluster, err := getClusterByID(c, targetClusterID)
		if err != nil {
			if contextOverride != "" {
				return fmt.Errorf("invalid --context %q: %w", contextOverride, err)
			}
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

		if err := setKubeConfigForUserAndCluster(ctx, currentUser, targetClusterID, kubeConfig); err != nil {
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

	cmd := exec.Command(path, kubectlArgs...)
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

func validateToken(tokenID string, tokenClient client.TokenOperations) (bool, error) {
	token, err := tokenClient.ByID(tokenID)
	if err != nil {
		if !clientbase.IsNotFound(err) {
			return false, err
		}
		return false, nil
	}
	return !token.Expired, nil
}

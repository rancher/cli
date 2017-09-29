package cmd

import (
	"net/url"
	"os"

	dockerterm "github.com/docker/docker/pkg/term"
	"github.com/rancher/go-rancher/v3"
	"github.com/urfave/cli"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/pkg/kubectl/util/term"
)

func setupTTY(ctx *cli.Context) term.TTY {
	t := term.TTY{
		In:  os.Stdin,
		Out: os.Stdout,
	}

	// if we get to here, the user wants to attach stdin, wants a TTY, and o.In is a terminal, so we
	// can safely set t.Raw to true

	if !ctx.Bool("interactive") {
		t.In = nil
		t.Raw = false
		return t
	}
	isTerminalIn := func(tty term.TTY) bool {
		return tty.IsTerminalIn()
	}
	if !isTerminalIn(t) {
		return t
	}
	t.Raw = true
	// use dockerterm.StdStreams() to get the right I/O handles on Windows
	stdin, stdout, _ := dockerterm.StdStreams()
	t.In = stdin
	t.Out = stdout
	return t
}

func constructRestConfig(ctx *cli.Context, c *client.RancherClient) (*restclient.Config, error) {
	config, err := lookupConfig(ctx)
	if err != nil && err != errNoURL {
		return nil, err
	}

	project, err := c.Project.ById(config.Environment)
	if err != nil {
		return nil, err
	}

	cluster, err := c.Cluster.ById(project.ClusterId)
	if err != nil {
		return nil, err
	}

	u, _ := url.Parse(c.GetOpts().Url)
	u.Path = "/k8s/clusters/" + cluster.Id

	filePath, err := generateKubeconfigFile(c, ctx)
	if err != nil {
		if filePath != "" {
			os.RemoveAll(filePath)
		}
		return nil, err
	}
	defer os.RemoveAll(filePath)

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = filePath

	loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{},
	)

	conf, err := loader.ClientConfig()
	if err != nil {
		return nil, err
	}
	contentConfig := dynamic.ContentConfig()
	contentConfig.GroupVersion = &schema.GroupVersion{Group: "api", Version: "v1"}
	conf.ContentConfig = contentConfig
	return conf, nil
}

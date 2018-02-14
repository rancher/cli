package cliclient

import (
	"strings"

	"github.com/rancher/cli/config"

	"github.com/rancher/norman/clientbase"
	clusterClient "github.com/rancher/types/client/cluster/v3"
	managementClient "github.com/rancher/types/client/management/v3"
	projectClient "github.com/rancher/types/client/project/v3"
)

type MasterClient struct {
	ClusterClient    *clusterClient.Client
	ManagementClient *managementClient.Client
	ProjectClient    *projectClient.Client
	UserConfig       *config.ServerConfig
}

func NewMasterClient(config *config.ServerConfig) (*MasterClient, error) {
	mc := &MasterClient{
		UserConfig: config,
	}

	clustProj := SplitOnColon(config.Project)

	serverURL := config.URL

	if !strings.HasSuffix(serverURL, "/v3") {
		serverURL = config.URL + "/v3"
	}

	options := &clientbase.ClientOpts{
		URL:       serverURL,
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
		CACerts:   config.CACerts,
	}

	// Setup the management client
	mClient, err := managementClient.NewClient(options)
	if err != nil {
		return nil, err
	}
	mc.ManagementClient = mClient

	// Setup the cluster client
	if len(clustProj) == 2 {
		options.URL = serverURL + "/clusters/" + clustProj[0]
	}
	cClient, err := clusterClient.NewClient(options)
	if err != nil {
		return nil, err
	}
	mc.ClusterClient = cClient

	// Setup the project client
	if len(clustProj) == 2 {
		options.URL = serverURL + "/projects/" + config.Project
	}
	pClient, err := projectClient.NewClient(options)
	if err != nil {
		return nil, err
	}
	mc.ProjectClient = pClient

	return mc, nil
}

func SplitOnColon(s string) []string {
	return strings.Split(s, ":")
}

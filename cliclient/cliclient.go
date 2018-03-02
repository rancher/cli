package cliclient

import (
	"strings"

	"github.com/rancher/cli/config"
	"github.com/sirupsen/logrus"

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

// NewMasterClient returns a new MasterClient with Cluster Management and Cluster
// clients populated
func NewMasterClient(config *config.ServerConfig) (*MasterClient, error) {
	mc, err := NewManagementClient(config)
	if nil != err {
		return nil, err
	}

	clustProj := SplitOnColon(config.Project)

	options := createClientOpts(config)

	baseURL := options.URL

	// Setup the cluster client
	if len(clustProj) != 2 {
		logrus.Warn("No default project set, run `rancher login` again. " +
			"Some commands will not work until project is set")
	}
	options.URL = baseURL + "/clusters/" + clustProj[0]

	cClient, err := clusterClient.NewClient(options)
	if err != nil {
		return nil, err
	}
	mc.ClusterClient = cClient

	// Setup the project client
	pClient, err := NewProjectClient(config)
	if err != nil {
		return nil, err
	}
	mc.ProjectClient = pClient.ProjectClient

	return mc, nil
}

//NewManagementClient returns a new MasterClient with only the Management client
func NewManagementClient(config *config.ServerConfig) (*MasterClient, error) {
	mc := &MasterClient{
		UserConfig: config,
	}

	options := createClientOpts(config)

	// Setup the management client
	mClient, err := managementClient.NewClient(options)
	if err != nil {
		return nil, err
	}
	mc.ManagementClient = mClient
	return mc, nil
}

func NewProjectClient(config *config.ServerConfig) (*MasterClient, error) {
	mc := &MasterClient{
		UserConfig: config,
	}

	options := createClientOpts(config)
	options.URL = options.URL + "/projects/" + config.Project

	// Setup the project client
	pc, err := projectClient.NewClient(options)
	if err != nil {
		return nil, err
	}
	mc.ProjectClient = pc
	return mc, nil
}

func createClientOpts(config *config.ServerConfig) *clientbase.ClientOpts {
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
	return options
}

func SplitOnColon(s string) []string {
	return strings.Split(s, ":")
}

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

// NewMasterClient returns a new MasterClient with Management, Cluster and Project
// clients populated
func NewMasterClient(config *config.ServerConfig) (*MasterClient, error) {
	mc := &MasterClient{
		UserConfig: config,
	}

	// Setup the management client
	err := mc.NewManagementClient()
	if err != nil {
		return nil, err
	}

	// Setup the cluster client
	err = mc.NewClusterClient()
	if err != nil {
		return nil, err
	}

	// Setup the project client
	err = mc.NewProjectClient()
	if err != nil {
		return nil, err
	}

	return mc, nil
}

func (mc *MasterClient) NewManagementClient() error {
	options := createClientOpts(mc.UserConfig)

	// Setup the management client
	mClient, err := managementClient.NewClient(options)
	if err != nil {
		return err
	}
	mc.ManagementClient = mClient
	return nil
}

func (mc *MasterClient) NewClusterClient() error {
	clustProj := CheckProject(mc.UserConfig.Project)
	if clustProj == nil {
		return nil
	}

	options := createClientOpts(mc.UserConfig)
	baseURL := options.URL
	options.URL = baseURL + "/clusters/" + clustProj[0]

	// Setup the cluster client
	cClient, err := clusterClient.NewClient(options)
	if err != nil {
		return err
	}
	mc.ClusterClient = cClient
	return nil
}

func (mc *MasterClient) NewProjectClient() error {
	clustProj := CheckProject(mc.UserConfig.Project)
	if clustProj == nil {
		return nil
	}

	options := createClientOpts(mc.UserConfig)
	options.URL = options.URL + "/projects/" + mc.UserConfig.Project

	// Setup the project client
	pc, err := projectClient.NewClient(options)
	if err != nil {
		return err
	}
	mc.ProjectClient = pc
	return nil
}

//NewManagementClient returns a new MasterClient with only the Management client
func NewManagementClient(config *config.ServerConfig) (*MasterClient, error) {
	mc := &MasterClient{
		UserConfig: config,
	}

	err := mc.NewManagementClient()
	if err != nil {
		return nil, err
	}
	return mc, nil
}

//NewClusterClient returns a new MasterClient with only the Cluster client
func NewClusterClient(config *config.ServerConfig) (*MasterClient, error) {
	mc := &MasterClient{
		UserConfig: config,
	}

	err := mc.NewClusterClient()
	if err != nil {
		return nil, err
	}
	return mc, nil
}

//NewProjectClient returns a new MasterClient with only the Project client
func NewProjectClient(config *config.ServerConfig) (*MasterClient, error) {
	mc := &MasterClient{
		UserConfig: config,
	}

	err := mc.NewProjectClient()
	if err != nil {
		return nil, err
	}
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
		TokenKey:  config.TokenKey,
		CACerts:   config.CACerts,
	}
	return options
}

func SplitOnColon(s string) []string {
	return strings.Split(s, ":")
}

// CheckProject check project string format and returns it as []string or nil
func CheckProject(s string) []string {
	if len(s) == 0 {
		logrus.Warn("No default project set, run `rancher login` again. " +
			"Some commands will not work until project is set")
		return nil
	}

	clustProj := SplitOnColon(s)

	if len(clustProj) != 2 {
		logrus.Warn("No default project set, run `rancher login` again. " +
			"Some commands will not work until project is set")
		return nil
	}

	return clustProj
}

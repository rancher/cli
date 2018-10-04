package cliclient

import (
	"errors"
	"strings"

	"github.com/rancher/cli/config"
	"github.com/rancher/norman/clientbase"
	ntypes "github.com/rancher/norman/types"
	clusterClient "github.com/rancher/types/client/cluster/v3"
	managementClient "github.com/rancher/types/client/management/v3"
	projectClient "github.com/rancher/types/client/project/v3"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type MasterClient struct {
	ClusterClient    *clusterClient.Client
	ManagementClient *managementClient.Client
	ProjectClient    *projectClient.Client
	UserConfig       *config.ServerConfig
}

// NewMasterClient returns a new MasterClient with Cluster, Management and Project
// clients populated
func NewMasterClient(config *config.ServerConfig) (*MasterClient, error) {
	mc := &MasterClient{
		UserConfig: config,
	}

	clustProj := CheckProject(mc.UserConfig.Project)
	if clustProj == nil {
		logrus.Warn("No context set; some commands will not work. Run `rancher login` again.")
	}

	var g errgroup.Group

	g.Go(mc.newManagementClient)
	g.Go(mc.newClusterClient)
	g.Go(mc.newProjectClient)

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return mc, nil
}

// NewManagementClient returns a new MasterClient with only the Management client
func NewManagementClient(config *config.ServerConfig) (*MasterClient, error) {
	mc := &MasterClient{
		UserConfig: config,
	}

	err := mc.newManagementClient()
	if err != nil {
		return nil, err
	}
	return mc, nil
}

// NewClusterClient returns a new MasterClient with only the Cluster client
func NewClusterClient(config *config.ServerConfig) (*MasterClient, error) {
	clustProj := CheckProject(config.Project)
	if clustProj == nil {
		return nil, errors.New("no context set")
	}

	mc := &MasterClient{
		UserConfig: config,
	}

	err := mc.newClusterClient()
	if err != nil {
		return nil, err
	}
	return mc, nil
}

// NewProjectClient returns a new MasterClient with only the Project client
func NewProjectClient(config *config.ServerConfig) (*MasterClient, error) {
	clustProj := CheckProject(config.Project)
	if clustProj == nil {
		return nil, errors.New("no context set")
	}

	mc := &MasterClient{
		UserConfig: config,
	}

	err := mc.newProjectClient()
	if err != nil {
		return nil, err
	}
	return mc, nil
}

func (mc *MasterClient) newManagementClient() error {
	options := createClientOpts(mc.UserConfig)

	// Setup the management client
	mClient, err := managementClient.NewClient(options)
	if err != nil {
		return err
	}
	mc.ManagementClient = mClient

	return nil
}

func (mc *MasterClient) newClusterClient() error {
	options := createClientOpts(mc.UserConfig)
	options.URL = options.URL + "/clusters/" + mc.UserConfig.FocusedCluster()

	// Setup the project client
	cc, err := clusterClient.NewClient(options)
	if err != nil {
		return err
	}
	mc.ClusterClient = cc

	return nil
}

func (mc *MasterClient) newProjectClient() error {
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

func (mc *MasterClient) ByID(resource *ntypes.Resource, respObject interface{}) error {
	if _, ok := mc.ManagementClient.APIBaseClient.Types[resource.Type]; ok {
		err := mc.ManagementClient.ByID(resource.Type, resource.ID, &respObject)
		if err != nil {
			return err
		}
	} else if _, ok := mc.ProjectClient.APIBaseClient.Types[resource.Type]; ok {
		err := mc.ProjectClient.ByID(resource.Type, resource.ID, &respObject)
		if err != nil {
			return err
		}
	} else if _, ok := mc.ClusterClient.APIBaseClient.Types[resource.Type]; ok {
		err := mc.ClusterClient.ByID(resource.Type, resource.ID, &respObject)
		if err != nil {
			return err
		}
	}
	return errors.New("unknown resource type")
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

// CheckProject verifies s matches the valid project ID of <cluster>:<project>
func CheckProject(s string) []string {
	clustProj := SplitOnColon(s)

	if len(s) == 0 || len(clustProj) != 2 {
		return nil
	}

	return clustProj
}

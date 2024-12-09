package cliclient

import (
	"errors"
	"fmt"
	"strings"

	errorsPkg "github.com/pkg/errors"
	"github.com/rancher/cli/config"
	"github.com/rancher/norman/clientbase"
	ntypes "github.com/rancher/norman/types"
	capiClient "github.com/rancher/rancher/pkg/client/generated/cluster/v1beta1"
	clusterClient "github.com/rancher/rancher/pkg/client/generated/cluster/v3"
	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	projectClient "github.com/rancher/rancher/pkg/client/generated/project/v3"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type MasterClient struct {
	ClusterClient    *clusterClient.Client
	ManagementClient *managementClient.Client
	ProjectClient    *projectClient.Client
	UserConfig       *config.ServerConfig
	CAPIClient       *capiClient.Client
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
	g.Go(mc.newCAPIClient)

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
		if clientbase.IsNotFound(err) {
			err = errorsPkg.WithMessage(err, "Current cluster not available, try running `rancher context switch`. Error")
		}
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
		if clientbase.IsNotFound(err) {
			err = errorsPkg.WithMessage(err, "Current project not available, try running `rancher context switch`. Error")
		}
		return err
	}
	mc.ProjectClient = pc

	return nil
}

func (mc *MasterClient) newCAPIClient() error {
	options := createClientOpts(mc.UserConfig)
	options.URL = strings.TrimSuffix(options.URL, "/v3") + "/v1"

	// Setup the CAPI client
	cc, err := capiClient.NewClient(options)
	if err != nil {
		return err
	}
	mc.CAPIClient = cc

	return nil
}

func (mc *MasterClient) ByID(resource *ntypes.Resource, respObject interface{}) error {
	if strings.HasPrefix(resource.Type, "cluster.x-k8s.io") {
		return mc.CAPIClient.ByID(resource.Type, resource.ID, &respObject)
	} else if _, ok := mc.ManagementClient.APIBaseClient.Types[resource.Type]; ok {
		return mc.ManagementClient.ByID(resource.Type, resource.ID, &respObject)
	} else if _, ok := mc.ProjectClient.APIBaseClient.Types[resource.Type]; ok {
		return mc.ProjectClient.ByID(resource.Type, resource.ID, &respObject)
	} else if _, ok := mc.ClusterClient.APIBaseClient.Types[resource.Type]; ok {
		return mc.ClusterClient.ByID(resource.Type, resource.ID, &respObject)
	}
	return fmt.Errorf("MasterClient - unknown resource type %v", resource.Type)
}

func createClientOpts(config *config.ServerConfig) *clientbase.ClientOpts {
	serverURL := config.URL

	if !strings.HasSuffix(serverURL, "/v3") {
		serverURL = config.URL + "/v3"
	}

	return &clientbase.ClientOpts{
		URL:       serverURL,
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
		CACerts:   config.CACerts,
		ProxyURL:  config.ProxyURL,
		Timeout:   config.GetHTTPTimeout(),
	}
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

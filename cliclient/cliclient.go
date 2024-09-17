package cliclient

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

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

// HTTPClienter is a http.Client factory interface
type HTTPClienter interface {
	New() *http.Client
}

// DefaultHTTPClient stores the default http.Client factory
var DefaultHTTPClient HTTPClienter = &HTTPClient{}

/*
TestingReplaceDefaultHTTPClient replaces DefaultHTTPClient for unit tests.
Not thread-safe.
Call the returned function by defer keyword, for example:

	defer cliclient.TestingReplaceDefaultHTTPClient(mockClient)()
*/
func TestingReplaceDefaultHTTPClient(mockClient HTTPClienter) func() {
	origHttpClient := DefaultHTTPClient
	DefaultHTTPClient = mockClient
	return func() {
		DefaultHTTPClient = origHttpClient
	}
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

var testingForceClientInsecure = false

/*
TestingForceClientInsecure sets testForceClientInsecure to true for unit tests.
It's a workaround to github.com/rancher/norman/clientbase.NewAPIClient,
which replaces net/http.Client.Transport (including proxy and TLS config),
so the client TLS config of net/http/httptest.Server will be lost.
Not thread-safe.
Call the returned function by defer keyword, for example:

	defer cliclient.TestingForceClientInsecure()()
*/
func TestingForceClientInsecure() func() {
	origTestForceClientInsecure := testingForceClientInsecure
	testingForceClientInsecure = true
	return func() {
		testingForceClientInsecure = origTestForceClientInsecure
	}
}

func (mc *MasterClient) newManagementClient() error {
	options := createClientOpts(mc.UserConfig)
	if testingForceClientInsecure {
		options.Insecure = true
	}

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

	options := &clientbase.ClientOpts{
		HTTPClient: DefaultHTTPClient.New(),
		URL:        serverURL,
		AccessKey:  config.AccessKey,
		SecretKey:  config.SecretKey,
		CACerts:    config.CACerts,
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

type HTTPClient struct{}

/*
HTTPClient.New makes http.Client including http.Transport,
with default values (for example: proxy) and custom timeouts.
See: https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
*/
func (c *HTTPClient) New() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialer := &net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}
		return dialer.DialContext(ctx, network, addr)
	}
	transport.ResponseHeaderTimeout = 10 * time.Second

	return &http.Client{
		Transport: transport,
		Timeout:   time.Minute, // from github.com/rancher/norman/clientbase.NewAPIClient
	}
}

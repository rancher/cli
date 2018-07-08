package cmd

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/grantae/certinfo"
	"github.com/rancher/cli/cliclient"
	"github.com/rancher/cli/config"
	managementClient "github.com/rancher/types/client/management/v3"
	"github.com/urfave/cli"
)

type LoginData struct {
	Project     managementClient.Project
	Index       int
	ClusterName string
}

type CACertResponse struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func LoginCommand() cli.Command {
	return cli.Command{
		Name:      "login",
		Aliases:   []string{"l"},
		Usage:     "Login to a Rancher server",
		Action:    loginSetup,
		ArgsUsage: "[SERVERURL]",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "context",
				Usage: "Get the currently set context",
			},
			cli.StringFlag{
				Name:  "token,t",
				Usage: "Token from the Rancher UI",
			},
			cli.StringFlag{
				Name:  "cacert",
				Usage: "Location of the CACerts to use",
			},
			cli.StringFlag{
				Name:  "name",
				Usage: "Name of the Server",
			},
		},
	}
}

func loginSetup(ctx *cli.Context) error {
	if ctx.Bool("context") {
		err := loginContext(ctx)
		if nil != err {
			return err
		}
		return nil
	}

	if ctx.NArg() == 0 {
		return cli.ShowCommandHelp(ctx, "login")
	}

	path := ctx.GlobalString("cf")
	if path == "" {
		path = os.ExpandEnv("${HOME}/.rancher/cli2.json")
	}

	cf, err := loadConfig(path)
	if err != nil {
		return err
	}

	serverName := ctx.String("name")
	if serverName == "" {
		serverName = "rancherDefault"
	}

	serverConfig := &config.ServerConfig{}

	// Validate the url and drop the path
	u, err := url.ParseRequestURI(ctx.Args().First())
	if err != nil {
		return fmt.Errorf("Failed to parse SERVERURL (%s), make sure it is a valid HTTPS URL (e.g. https://rancher.yourdomain.com or https://1.1.1.1). Error: %s", ctx.Args().First(), err)
	}

	u.Path = ""
	serverConfig.URL = u.String()

	if ctx.String("token") != "" {
		auth := SplitOnColon(ctx.String("token"))
		if len(auth) != 2 {
			return errors.New("invalid token")
		}
		serverConfig.AccessKey = auth[0]
		serverConfig.SecretKey = auth[1]
		serverConfig.TokenKey = ctx.String("token")
	} else {
		// This can be removed once username and password is accepted
		return errors.New("token flag is required")
	}

	if ctx.String("cacert") != "" {
		cert, err := loadAndVerifyCert(ctx.String("cacert"))
		if nil != err {
			return err
		}
		serverConfig.CACerts = cert

	}

	c, err := cliclient.NewManagementClient(serverConfig)
	if nil != err {
		if _, ok := err.(*url.Error); ok && strings.Contains(err.Error(), "certificate signed by unknown authority") {
			// no cert was provided and it's most likely a self signed cert if
			// we get here so grab the cacert and see if the user accepts the server
			c, err = getCertFromServer(serverConfig)
			if nil != err {
				return err
			}
		} else {
			return err
		}
	}

	proj, err := getProjectContext(ctx, c)
	if nil != err {
		return err
	}

	// Set the default server and proj for the user
	serverConfig.Project = proj
	cf.CurrentServer = serverName
	cf.Servers[serverName] = serverConfig

	cf.Write()

	return nil
}

func getProjectContext(ctx *cli.Context, c *cliclient.MasterClient) (string, error) {
	projectCollection, err := c.ManagementClient.Project.List(defaultListOpts(ctx))
	if err != nil {
		return "", err
	}

	if len(projectCollection.Data) == 0 {
		logrus.Warn("There are no projects in the cluster, please create one and try again")
		return "", nil
	}

	if len(projectCollection.Data) == 1 {
		logrus.Infof("Only 1 project available: %s", projectCollection.Data[0].Name)
		return projectCollection.Data[0].ID, nil
	}

	clusterNames, err := getClusterNames(ctx, c)
	if err != nil {
		return "", err
	}

	writer := NewTableWriter([][]string{
		{"NUMBER", "Index"},
		{"CLUSTER NAME", "ClusterName"},
		{"PROJECT ID", "Project.ID"},
		{"PROJECT NAME", "Project.Name"},
		{"PROJECT DESCRIPTION", "Project.Description"},
	}, ctx)

	for i, item := range projectCollection.Data {
		writer.Write(&LoginData{
			Project:     item,
			Index:       i + 1,
			ClusterName: clusterNames[item.ClusterId],
		})
	}

	writer.Close()
	if nil != writer.Err() {
		return "", writer.Err()
	}

	fmt.Print("Select a Project:")

	reader := bufio.NewReader(os.Stdin)

	errMessage := fmt.Sprintf("Invalid input, enter a number between 1 and %v: ", len(projectCollection.Data))
	var selection int

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		input = strings.TrimSpace(input)

		if input != "" {
			i, err := strconv.Atoi(input)
			if nil != err {
				fmt.Print(errMessage)
				continue
			}
			if i <= len(projectCollection.Data) && i != 0 {
				selection = i - 1
				break
			}
			fmt.Print(errMessage)
			continue
		}
	}
	return projectCollection.Data[selection].ID, nil
}

func getCertFromServer(cf *config.ServerConfig) (*cliclient.MasterClient, error) {
	req, err := http.NewRequest("GET", cf.URL+"/v3/settings/cacerts", nil)
	if nil != err {
		return nil, err
	}

	req.SetBasicAuth(cf.AccessKey, cf.SecretKey)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	res, err := client.Do(req)
	if nil != err {
		return nil, err
	}

	defer res.Body.Close()

	content, err := ioutil.ReadAll(res.Body)
	if nil != err {
		return nil, err
	}

	var certReponse *CACertResponse
	err = json.Unmarshal(content, &certReponse)

	cert, err := verifyCert([]byte(certReponse.Value))
	if nil != err {
		return nil, err
	}

	// Get the server cert chain in a printable form
	serverCerts, err := processServerChain(res)
	if nil != err {
		return nil, err
	}

	if ok := verifyUserAcceptsCert(serverCerts, cf.URL); ok {
		cf.CACerts = cert
	} else {
		return nil, errors.New("CACert of server was not accepted, unable to login")
	}

	return cliclient.NewManagementClient(cf)
}

func verifyUserAcceptsCert(certs []string, url string) bool {
	fmt.Printf("The authenticity of server '%s' can't be established.\n", url)
	fmt.Printf("Cert chain is : %v \n", certs)
	fmt.Print("Do you want to continue connecting (yes/no)? ")

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		input := scanner.Text()
		input = strings.ToLower(strings.TrimSpace(input))

		if input == "yes" || input == "y" {
			return true
		} else if input == "no" || input == "n" {
			return false
		}
		fmt.Printf("Please type 'yes' or 'no': ")
	}
	return false
}

func processServerChain(res *http.Response) ([]string, error) {
	var allCerts []string

	for _, cert := range res.TLS.PeerCertificates {
		result, err := certinfo.CertificateText(cert)
		if err != nil {
			return allCerts, err
		}
		allCerts = append(allCerts, result)
	}
	return allCerts, nil
}

func loginContext(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	cluster, err := getClusterByID(c, c.UserConfig.FocusedCluster())
	if nil != err {
		return err
	}
	clusterName := getClusterName(cluster)

	project, err := getProjectByID(c, c.UserConfig.Project)
	if nil != err {
		return err
	}

	fmt.Printf("Cluster:%s Project:%s\n", clusterName, project.Name)
	return nil
}

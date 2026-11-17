package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/rancher/cli/cliclient"
	"github.com/rancher/cli/config"
	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
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

func LoginCommand() *cli.Command {
	return &cli.Command{
		Name:      "login",
		Aliases:   []string{"l"},
		Usage:     "Login to a Rancher server",
		Action:    loginSetup,
		ArgsUsage: "[SERVERURL]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "context",
				Usage: "Set the context during login",
			},
			&cli.StringFlag{
				Name:    "token",
				Aliases: []string{"t"},
				Usage:   "Token from the Rancher UI",
			},
			&cli.StringFlag{
				Name:  "cacert",
				Usage: "Location of the CACerts to use",
			},
			&cli.StringFlag{
				Name:  "name",
				Usage: "Name of the Server",
			},
			&cli.StringFlag{
				Name: "auth-provider",
				Usage: "Name or type of the auth provider to sign in with when no --token is given " +
					"(e.g. keyCloakProvider); prompts if the server has several",
			},
			&cli.StringFlag{
				Name:  "auth-flow",
				Usage: "Auth flow to use for OAuth providers: 'devicecode' (default) or 'authcode'",
			},
			&cli.BoolFlag{
				Name:  "skip-browser",
				Usage: "Print the login URL instead of opening it in the default browser",
			},
		},
	}
}

// loginOpenBrowser opens the SAML login page during "rancher login"; tests
// replace it.
var loginOpenBrowser openBrowserFunc = openBrowser

func loginSetup(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() == 0 {
		return cli.ShowCommandHelp(ctx, cmd, "login")
	}

	cf, err := loadConfig(cmd)
	if err != nil {
		return err
	}

	serverName := cmd.String("name")
	if serverName == "" {
		serverName = "rancherDefault"
	}

	serverConfig := &config.ServerConfig{}

	// Validate the url and drop the path
	u, err := url.ParseRequestURI(cmd.Args().First())
	if err != nil {
		return fmt.Errorf("failed to parse SERVERURL (%s), make sure it is a valid HTTPS URL (e.g. https://rancher.yourdomain.com or https://1.1.1.1). Error: %s", cmd.Args().First(), err)
	}

	u.Path = ""
	serverConfig.URL = u.String()

	if cmd.String("cacert") != "" {
		cert, err := loadAndVerifyCert(cmd.String("cacert"))
		if err != nil {
			return err
		}
		serverConfig.CACerts = cert
	}

	if cmd.String("token") != "" {
		if cmd.String("auth-provider") != "" || cmd.String("auth-flow") != "" || cmd.Bool("skip-browser") {
			return errors.New("--token cannot be combined with --auth-provider, --auth-flow or --skip-browser")
		}
		if err := setServerToken(serverConfig, cmd.String("token")); err != nil {
			return errors.New("invalid token")
		}
	} else if err := browserLogin(serverConfig, cmd.String("auth-provider"), cmd.String("auth-flow"), cmd.Bool("skip-browser")); err != nil {
		return err
	}

	c, err := cliclient.NewManagementClient(serverConfig)
	if err != nil {
		return err
	}

	proj, err := getProjectContext(cmd, c)
	if err != nil {
		return err
	}

	// Set the default server and proj for the user
	serverConfig.Project = proj
	cf.CurrentServer = serverName
	cf.Servers[serverName] = serverConfig

	err = cf.Write()
	if err != nil {
		return err
	}

	return nil
}

// setServerToken stores a token of the form <access key>:<secret key> in the
// server config.
func setServerToken(serverConfig *config.ServerConfig, token string) error {
	auth := SplitOnColon(token)
	if len(auth) != 2 || auth[0] == "" || auth[1] == "" {
		return errors.New("token is not of the form <access key>:<secret key>")
	}
	serverConfig.AccessKey = auth[0]
	serverConfig.SecretKey = auth[1]
	serverConfig.TokenKey = token
	return nil
}

// browserLogin signs the user in to serverConfig.URL with an auth provider,
// in the browser for SAML and OAuth providers or with a username and password
// otherwise, and stores the token Rancher issues in serverConfig.
func browserLogin(serverConfig *config.ServerConfig, authProvider, authFlow string, skipBrowser bool) error {
	// CACerts holds the PEM-encoded certificates read from --cacert.
	tlsConfig, err := tlsConfigFromPEM(false, serverConfig.CACerts)
	if err != nil {
		return err
	}
	client, err := newHTTPClient(serverConfig, tlsConfig)
	if err != nil {
		return err
	}

	input := &LoginInput{
		server:       serverConfig.URL,
		authProvider: authProvider,
		authFlow:     authFlow,
		openBrowser:  loginOpenBrowser,
		skipBrowser:  skipBrowser,
		responseType: sessionResponseType,
	}
	token, err := loginAndGetToken(client, input)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	if token.BearerToken == "" {
		return errors.New("login failed: no token received from the server")
	}
	if err := setServerToken(serverConfig, token.BearerToken); err != nil {
		return fmt.Errorf("login failed: unexpected token from the server: %w", err)
	}
	serverConfig.TokenExpiresAt = token.ExpiresAt

	if token.ExpiresAt != "" {
		customPrint(fmt.Sprintf("Logged in, token expires at %s", token.ExpiresAt))
	} else {
		customPrint("Logged in")
	}
	return nil
}

func getProjectContext(cmd *cli.Command, c *cliclient.MasterClient) (string, error) {
	// If context is given
	if cmd.String("context") != "" {
		context := cmd.String("context")
		// Check if given context is in valid format
		_, _, err := parseClusterAndProjectID(context)
		if err != nil {
			return "", fmt.Errorf("unable to parse context (%s). Please provide context as local:p-xxxxx, c-xxxxx:p-xxxxx, c-xxxxx:project-xxxxx, c-m-xxxxxxxx:p-xxxxx or c-m-xxxxxxxx:project-xxxxx", context)
		}
		// Check if context exists
		_, err = Lookup(c, context, "project")
		if err != nil {
			return "", fmt.Errorf("unable to find context (%s). Make sure the context exists and you have permissions to use it. Error: %s", context, err)
		}
		return context, nil
	}

	projectCollection, err := c.ManagementClient.Project.List(defaultListOpts(cmd))
	if err != nil {
		return "", err
	}

	projDataLen := len(projectCollection.Data)

	if projDataLen == 0 {
		logrus.Warn("No projects found, context could not be set. Please create a project and run `rancher login` again.")
		return "", nil
	}

	if projDataLen == 1 {
		logrus.Infof("Only 1 project available: %s", projectCollection.Data[0].Name)
		return projectCollection.Data[0].ID, nil
	}

	if projDataLen == 2 {
		var sysProj bool
		var defaultID string
		for _, proj := range projectCollection.Data {
			if proj.Name == "Default" {
				defaultID = proj.ID
			}
			if proj.Name == "System" {
				sysProj = true
			}
			if sysProj && defaultID != "" {
				return defaultID, nil
			}
		}
	}

	clusterNames, err := getClusterNames(cmd, c)
	if err != nil {
		return "", err
	}

	writer := NewTableWriter([][]string{
		{"NUMBER", "Index"},
		{"CLUSTER NAME", "ClusterName"},
		{"PROJECT ID", "Project.ID"},
		{"PROJECT NAME", "Project.Name"},
		{"PROJECT DESCRIPTION", "Project.Description"},
	}, cmd)

	for i, item := range projectCollection.Data {
		writer.Write(&LoginData{
			Project:     item,
			Index:       i + 1,
			ClusterName: clusterNames[item.ClusterID],
		})
	}

	writer.Close()
	if nil != writer.Err() {
		return "", writer.Err()
	}

	fmt.Print("Select a Project:")

	// Read from the reader the sign-in prompts share, which may have read
	// ahead of the project selection.
	reader := stdin

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
			if err != nil {
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

func loginContext(ctx context.Context, cmd *cli.Command) error {
	c, err := GetClient(cmd)
	if err != nil {
		return err
	}

	cluster, err := getClusterByID(c, c.UserConfig.GetCurrentCluster())
	if err != nil {
		return err
	}
	clusterName := getClusterName(cluster)

	project, err := getProjectByID(c, c.UserConfig.Project)
	if err != nil {
		return err
	}

	fmt.Printf("Cluster:%s Project:%s\n", clusterName, project.Name)
	return nil
}

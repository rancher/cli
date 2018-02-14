package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rancher/cli/cliclient"
	"github.com/rancher/cli/config"

	"github.com/urfave/cli"
)

func LoginCommand() cli.Command {
	return cli.Command{
		Name:      "login",
		Aliases:   []string{"l"},
		Usage:     "Login to a Rancher server",
		Action:    loginSetup,
		ArgsUsage: "None",
		Flags: []cli.Flag{
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
	path := ctx.GlobalString("cf")
	if path == "" {
		path = os.ExpandEnv("${HOME}/.rancher/cli.json")
	}

	cf, err := loadConfig(path)
	if err != nil {
		return err
	}

	serverName := ctx.String("name")
	if serverName == "" {
		serverName = RandomName()
	}

	serverConfig, ok := cf.Servers[serverName]
	if !ok {
		serverConfig = &config.ServerConfig{}
	}

	if ctx.NArg() == 0 || ctx.NArg() > 1 {
		return errors.New("one server is required")
	}
	serverConfig.URL = ctx.Args().First()

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

	proj, err := getDefaultProject(ctx, serverConfig)
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

func getDefaultProject(ctx *cli.Context, cf *config.ServerConfig) (string, error) {
	// Set this on global as it's an arg to login and will be needed for clients

	mc, err := cliclient.NewMasterClient(cf)
	if nil != err {
		return "", err
	}

	collection, err := mc.ManagementClient.Project.List(defaultListOpts(ctx))
	if err != nil {
		return "", err
	}

	if len(collection.Data) == 0 {
		fmt.Println("There are no projects in the cluster, please create one and login again")
		return "", nil
	}

	errMessage := fmt.Sprintf("invalid input, enter a number between 0 and %v", len(collection.Data)-1)

	fmt.Println("Select your default Project:")
	for i, project := range collection.Data {
		fmt.Printf("%v %v %v\n", i, project.ID, project.Name)
	}

	reader := bufio.NewReader(os.Stdin)
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
				fmt.Println(errMessage)
				continue
			}
			if i <= len(collection.Data)-1 {
				selection = i
				break
			}
			fmt.Println(errMessage)
			continue
		}
	}
	return collection.Data[selection].ID, nil
}

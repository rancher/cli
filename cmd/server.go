package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/cli/config"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type serverData struct {
	Index   int
	Current string
	Name    string
	URL     string
}

// ServerCommand defines the 'rancher server' sub-commands
func ServerCommand() cli.Command {
	return cli.Command{
		Name:  "server",
		Usage: "Operations for the server",
		Description: `Switch or view the server currently in focus.
`,
		Subcommands: []cli.Command{
			{
				Name:   "current",
				Usage:  "Display the current server",
				Action: serverCurrent,
			},
			{
				Name:      "delete",
				Usage:     "Delete a server from the local config",
				ArgsUsage: "[SERVER_NAME]",
				Description: `
The server arg is optional, if not passed in a list of available servers will
be displayed and one can be selected.
`,
				Action: serverDelete,
			},
			{
				Name:   "ls",
				Usage:  "List all servers",
				Action: serverLs,
			},
			{
				Name:      "switch",
				Usage:     "Switch to a new server",
				ArgsUsage: "[SERVER_NAME]",
				Description: `
The server arg is optional, if not passed in a list of available servers will
be displayed and one can be selected.
`,
				Action: serverSwitch,
			},
		},
	}
}

// serverCurrent command to display the name of the current server in the local config
func serverCurrent(ctx *cli.Context) error {
	cf, err := loadConfig(ctx)
	if err != nil {
		return err
	}

	serverName := cf.CurrentServer
	URL := cf.Servers[serverName].URL
	fmt.Printf("Name: %s URL: %s\n", serverName, URL)
	return nil
}

// serverDelete command to delete a server from the local config
func serverDelete(ctx *cli.Context) error {
	cf, err := loadConfig(ctx)
	if err != nil {
		return err
	}

	if err := validateServersConfig(cf); err != nil {
		return err
	}

	var serverName string
	if ctx.NArg() == 1 {
		serverName = ctx.Args().First()
	} else {
		serverName, err = serverFromInput(ctx, cf)
		if err != nil {
			return err
		}
	}

	_, ok := cf.Servers[serverName]
	if !ok {
		return errors.New("Server not found")
	}

	delete(cf.Servers, serverName)
	err = cf.Write()
	if err != nil {
		return err
	}
	logrus.Infof("Server %s deleted", serverName)
	return nil
}

// serverLs command to list rancher servers from the local config
func serverLs(ctx *cli.Context) error {
	cf, err := loadConfig(ctx)
	if err != nil {
		return err
	}

	if err := validateServersConfig(cf); err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"CURRENT", "Current"},
		{"NAME", "Name"},
		{"URL", "URL"},
	}, ctx)

	defer writer.Close()

	for name, server := range cf.Servers {
		var current string
		if name == cf.CurrentServer {
			current = "*"
		}
		writer.Write(&serverData{
			Current: current,
			Name:    name,
			URL:     server.URL,
		})
	}

	return writer.Err()
}

// serverSwitch command to switch rancher server.
func serverSwitch(ctx *cli.Context) error {
	cf, err := loadConfig(ctx)
	if err != nil {
		return err
	}

	if err := validateServersConfig(cf); err != nil {
		return err
	}

	var serverName string
	if ctx.NArg() == 1 {
		serverName = ctx.Args().First()
	} else {
		serverName, err = serverFromInput(ctx, cf)
		if err != nil {
			return err
		}
	}
	_, ok := cf.Servers[serverName]
	if !ok {
		return errors.New("Server not found")
	}

	if len(cf.Servers[serverName].Project) == 0 {
		logrus.Warn("No context set; some commands will not work. Run 'rancher context switch'")
	}

	cf.CurrentServer = serverName
	err = cf.Write()
	if err != nil {
		return err
	}

	return nil
}

// serverFromInput displays the list of servers from the local config and
// prompt the user to select one.
func serverFromInput(ctx *cli.Context, cf config.Config) (string, error) {
	serverNames := getServerNames(cf)

	displayListServers(ctx, cf)

	fmt.Print("Select a Server:")
	reader := bufio.NewReader(os.Stdin)

	errMessage := fmt.Sprintf("Invalid input, enter a number between 1 and %v: ", len(serverNames))
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
			if i <= len(serverNames) && i != 0 {
				selection = i - 1
				break
			}
			fmt.Print(errMessage)
			continue
		}
	}

	return serverNames[selection], nil
}

// displayListServers displays the list of rancher servers
func displayListServers(ctx *cli.Context, cf config.Config) error {
	writer := NewTableWriter([][]string{
		{"INDEX", "Index"},
		{"NAME", "Name"},
		{"URL", "URL"},
	}, ctx)

	defer writer.Close()

	for idx, server := range getServerNames(cf) {
		writer.Write(&serverData{
			Index: idx + 1,
			Name:  server,
			URL:   cf.Servers[server].URL,
		})
	}
	return writer.Err()
}

// getServerNames returns an order slice of existing server names
func getServerNames(cf config.Config) []string {
	var serverNames []string
	for server := range cf.Servers {
		serverNames = append(serverNames, server)
	}
	sort.Strings(serverNames)
	return serverNames
}

func validateServersConfig(cnf config.Config) error {
	if len(cnf.Servers) == 0 {
		return errors.New("no servers are currently configured")
	}
	return nil
}

package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/cli/config"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/exp/maps"
)

type serverData struct {
	Index   int
	Current string
	Name    string
	URL     string
}

// ServerCommand defines the 'rancher server' sub-commands
func ServerCommand() cli.Command {
	cfg := &config.Config{}

	return cli.Command{
		Name:  "server",
		Usage: "Operations for the server",
		Description: `Switch or view the server currently in focus.
`,
		Before: loadAndValidateConfig(cfg),
		Subcommands: []cli.Command{
			{
				Name:  "current",
				Usage: "Display the current server",
				Action: func(ctx *cli.Context) error {
					return serverCurrent(ctx.App.Writer, cfg)
				},
			},
			{
				Name:      "delete",
				Usage:     "Delete a server from the local config",
				ArgsUsage: "[SERVER_NAME]",
				Description: `
The server arg is optional, if not passed in a list of available servers will
be displayed and one can be selected.
`,
				Action: func(ctx *cli.Context) error {
					serverName, err := getSelectedServer(ctx, cfg)
					if err != nil {
						return err
					}
					return serverDelete(cfg, serverName)
				},
			},
			{
				Name:  "ls",
				Usage: "List all servers",
				Action: func(ctx *cli.Context) error {
					format := ctx.String("format")
					return serverLs(ctx.App.Writer, cfg, format)
				},
			},
			{
				Name:      "switch",
				Usage:     "Switch to a new server",
				ArgsUsage: "[SERVER_NAME]",
				Description: `
		The server arg is optional, if not passed in a list of available servers will
		be displayed and one can be selected.
		`,
				Action: func(ctx *cli.Context) error {
					serverName, err := getSelectedServer(ctx, cfg)
					if err != nil {
						return err
					}
					return serverSwitch(cfg, serverName)
				},
			},
		},
	}
}

// serverCurrent command to display the name of the current server in the local config
func serverCurrent(out io.Writer, cfg *config.Config) error {
	serverName := cfg.CurrentServer

	currentServer, found := cfg.Servers[serverName]
	if !found {
		return errors.New("Current server not set")
	}

	fmt.Fprintf(out, "Name: %s URL: %s\n", serverName, currentServer.URL)
	return nil
}

// serverDelete command to delete a server from the local config
func serverDelete(cfg *config.Config, serverName string) error {
	_, ok := cfg.Servers[serverName]
	if !ok {
		return errors.New("Server not found")
	}
	delete(cfg.Servers, serverName)

	if cfg.CurrentServer == serverName {
		cfg.CurrentServer = ""
	}

	err := cfg.Write()
	if err != nil {
		return err
	}
	logrus.Infof("Server %s deleted", serverName)
	return nil
}

// serverLs command to list rancher servers from the local config
func serverLs(out io.Writer, cfg *config.Config, format string) error {
	writerConfig := &TableWriterConfig{
		Writer: out,
		Format: format,
	}

	writer := NewTableWriterWithConfig([][]string{
		{"CURRENT", "Current"},
		{"NAME", "Name"},
		{"URL", "URL"},
	}, writerConfig)

	defer writer.Close()

	servers := getServers(cfg)
	for _, server := range servers {
		writer.Write(server)
	}

	return writer.Err()
}

// serverSwitch will alter and write the config to switch rancher server.
func serverSwitch(cf *config.Config, serverName string) error {
	_, ok := cf.Servers[serverName]
	if !ok {
		return errors.New("Server not found")
	}

	if len(cf.Servers[serverName].Project) == 0 {
		logrus.Warn("No context set; some commands will not work. Run 'rancher context switch'")
	}

	cf.CurrentServer = serverName

	err := cf.Write()
	if err != nil {
		return err
	}
	return nil
}

// getSelectedServer will get the selected server if provided as argument,
// or it will prompt the user to select one.
func getSelectedServer(ctx *cli.Context, cfg *config.Config) (string, error) {
	serverName := ctx.Args().First()
	if serverName != "" {
		return serverName, nil
	}
	return serverFromInput(ctx, cfg)
}

// serverFromInput displays the list of servers from the local config and
// prompt the user to select one.
func serverFromInput(ctx *cli.Context, cf *config.Config) (string, error) {
	servers := getServers(cf)

	if err := displayListServers(ctx, servers); err != nil {
		return "", err
	}

	fmt.Print("Select a Server:")
	reader := bufio.NewReader(os.Stdin)

	errMessage := fmt.Sprintf("Invalid input, enter a number between 1 and %v: ", len(servers))
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
			if i <= len(servers) && i != 0 {
				selection = i - 1
				break
			}
			fmt.Print(errMessage)
			continue
		}
	}

	return servers[selection].Name, nil
}

// displayListServers displays the list of rancher servers
func displayListServers(ctx *cli.Context, servers []*serverData) error {
	writer := NewTableWriter([][]string{
		{"INDEX", "Index"},
		{"NAME", "Name"},
		{"URL", "URL"},
	}, ctx)

	defer writer.Close()

	for _, server := range servers {
		writer.Write(server)
	}
	return writer.Err()
}

// getServers returns an ordered slice (by name) of serverData
func getServers(cfg *config.Config) []*serverData {
	serverNames := maps.Keys(cfg.Servers)
	sort.Strings(serverNames)

	servers := []*serverData{}

	for i, server := range serverNames {
		var current string
		if server == cfg.CurrentServer {
			current = "*"
		}

		servers = append(servers, &serverData{
			Index:   i + 1,
			Name:    server,
			Current: current,
			URL:     cfg.Servers[server].URL,
		})
	}

	return servers
}

func loadAndValidateConfig(cfg *config.Config) cli.BeforeFunc {
	return func(ctx *cli.Context) error {
		conf, err := loadConfig(ctx)
		if err != nil {
			return err
		}
		*cfg = conf

		if len(cfg.Servers) == 0 {
			return errors.New("no servers are currently configured")
		}
		return nil
	}
}

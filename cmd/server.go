package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type serverData struct {
	Name string
	URL  string
}

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
				Usage:     "Delete a server",
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

func serverCurrent(ctx *cli.Context) error {
	cf, err := loadConfig(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", cf.CurrentServer)
	return nil
}

func serverDelete(ctx *cli.Context) error {
	cf, err := loadConfig(ctx)
	if err != nil {
		return err
	}

	serverName, err := serverSelect(ctx)
	if err != nil {
		return err
	}

	_, ok := cf.Servers[serverName]
	if !ok {
		return errors.New("Server not found")
	}

	delete(cf.Servers, serverName)

	cf.Write()
	logrus.Infof("Server %s deleted", serverName)
	return nil
}

func serverLs(ctx *cli.Context) error {
	cf, err := loadConfig(ctx)
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"NAME", "Name"},
		{"URL", "URL"},
	}, ctx)

	defer writer.Close()

	for name, server := range cf.Servers {
		writer.Write(&serverData{
			Name: name,
			URL:  server.URL,
		})
	}

	return writer.Err()
}

func serverSwitch(ctx *cli.Context) error {
	cf, err := loadConfig(ctx)
	if err != nil {
		return err
	}

	serverName, err := serverSelect(ctx)
	if err != nil {
		return err
	}

	_, ok := cf.Servers[serverName]
	if !ok {
		return errors.New("Server not found")
	}

	cf.CurrentServer = serverName
	cf.Write()

	return nil
}

func serverSelect(ctx *cli.Context) (string, error) {
	serverName := ""
	if ctx.NArg() == 1 {
		serverName = ctx.Args().First()
	} else {
		serverLs(ctx)
		fmt.Print("Select a Server:")

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		serverName = strings.TrimSpace(input)
	}
	return serverName, nil
}

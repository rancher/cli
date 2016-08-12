package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/cli/cmd"
	"github.com/urfave/cli"
)

var VERSION = "dev"

var AppHelpTemplate = `{{.Usage}}

Usage: {{.Name}} {{if .Flags}}[OPTIONS] {{end}}COMMAND [arg...]

Version: {{.Version}}
{{if .Flags}}
Options:
  {{range .Flags}}{{.}}
  {{end}}{{end}}
Commands:
  {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
  {{end}}
Run '{{.Name}} COMMAND --help' for more information on a command.
`

var CommandHelpTemplate = `{{.Usage}}
{{if .Description}}{{.Description}}{{end}}
Usage: rancher [global options] {{.Name}} {{if .Flags}}[OPTIONS] {{end}}{{if ne "None" .ArgsUsage}}{{if ne "" .ArgsUsage}}{{.ArgsUsage}}{{else}}[arg...]{{end}}{{end}}

{{if .Flags}}Options:{{range .Flags}}
	 {{.}}{{end}}{{end}}
`

func main() {
	if err := mainErr(); err != nil {
		logrus.Fatal(err)
	}
}

func mainErr() error {
	cli.AppHelpTemplate = AppHelpTemplate
	cli.CommandHelpTemplate = CommandHelpTemplate

	app := cli.NewApp()
	app.Name = "rancher"
	app.Usage = "Rancher CLI, managing containers one UTF-8 character at a time"
	app.Before = func(ctx *cli.Context) error {
		if ctx.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return nil
	}
	app.Version = VERSION
	app.Author = "Rancher Labs, Inc."
	app.Email = ""
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Debug logging",
		},
		cli.StringFlag{
			Name:   "config,c",
			Usage:  "Client configuration file (default ${HOME}/.rancher/cli.json)",
			EnvVar: "RANCHER_CLIENT_CONFIG",
		},
		cli.StringFlag{
			Name:   "environment,env",
			Usage:  "Environment name or ID",
			EnvVar: "RANCHER_ENVIRONMENT",
		},
		cli.StringFlag{
			Name:   "url",
			Usage:  "Specify the Rancher API endpoint URL",
			EnvVar: "RANCHER_URL",
		},
		cli.StringFlag{
			Name:   "access-key",
			Usage:  "Specify Rancher API access key",
			EnvVar: "RANCHER_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "secret-key",
			Usage:  "Specify Rancher API secret key",
			EnvVar: "RANCHER_SECRET_KEY",
		},
		cli.StringFlag{
			Name:   "host",
			Usage:  "Host used for docker command",
			EnvVar: "RANCHER_DOCKER_HOST",
		},
		cli.StringFlag{
			Name:  "rancher-file,r",
			Usage: "Specify an alternate Rancher compose file (default: rancher-compose.yml)",
		},
		cli.StringFlag{
			Name:  "env-file,e",
			Usage: "Specify a file from which to read environment variables",
		},
		cli.StringSliceFlag{
			Name:   "file,f",
			Usage:  "Specify one or more alternate compose files (default: docker-compose.yml)",
			Value:  &cli.StringSlice{},
			EnvVar: "COMPOSE_FILE",
		},
		cli.StringFlag{
			Name:  "stack,s",
			Usage: "Specify an alternate project name (default: directory name)",
		},
		cli.BoolFlag{
			Name:  "wait,w",
			Usage: "Wait for resource to reach resting state",
		},
		cli.IntFlag{
			Name:  "wait-timeout",
			Usage: "Timeout in seconds to wait",
			Value: 600,
		},
		cli.StringFlag{
			Name:  "wait-state",
			Usage: "State to wait for (active, healthy, etc)",
		},
	}
	app.Commands = []cli.Command{
		cmd.CatalogCommand(),
		cmd.ConfigCommand(),
		cmd.DockerCommand(),
		cmd.EnvCommand(),
		cmd.EventsCommand(),
		cmd.ExecCommand(),
		cmd.ExportCommand(),
		cmd.HostCommand(),
		cmd.LogsCommand(),
		cmd.PsCommand(),
		cmd.RestartCommand(),
		cmd.RmCommand(),
		cmd.RunCommand(),
		cmd.ScaleCommand(),
		cmd.SSHCommand(),
		cmd.StackCommand(),
		cmd.StartCommand(),
		cmd.StopCommand(),
		cmd.UpCommand(),
		//cmd.VolumeCommand(),
		cmd.WaitCommand(),
	}

	return app.Run(os.Args)
}

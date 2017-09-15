package main

import (
	"os"

	"regexp"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/rancher/cli/cmd"
	"github.com/rancher/cli/rancher_prompt"
	"github.com/urfave/cli"
)

var VERSION = "dev"

var AppHelpTemplate = `{{.Usage}}

Usage: {{.Name}} {{if .Flags}}[OPTIONS] {{end}}COMMAND [arg...]

Version: {{.Version}}
{{if .Flags}}
Options:
  {{range .Flags}}{{if .Hidden}}{{else}}{{.}}
  {{end}}{{end}}{{end}}
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
		// Below four flags are for rancher-compose code capability.  The users doesn't use them directly
		cli.StringFlag{
			Name:   "rancher-file",
			Hidden: true,
		},
		cli.StringFlag{
			Name:   "env-file",
			Hidden: true,
		},
		cli.StringSliceFlag{
			Name:   "file,f",
			Hidden: true,
		},
		cli.StringFlag{
			Name:   "project-name",
			Hidden: true,
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
		cmd.PullCommand(),
		cmd.PromptCommand(),
		cmd.RestartCommand(),
		cmd.RmCommand(),
		cmd.RunCommand(),
		cmd.ScaleCommand(),
		cmd.SecretCommand(),
		cmd.SSHCommand(),
		cmd.StackCommand(),
		cmd.StartCommand(),
		cmd.StopCommand(),
		cmd.UpCommand(),
		cmd.VolumeCommand(),
		cmd.InspectCommand(),
		cmd.WaitCommand(),
	}
	for _, com := range app.Commands {
		rancherPrompt.Commands[com.Name] = com
		rancherPrompt.Commands[com.ShortName] = com
	}
	rancherPrompt.Flags = app.Flags
	parsed, err := parseArgs(os.Args)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	return app.Run(parsed)
}

var singleAlphaLetterRegxp = regexp.MustCompile("[a-zA-Z]")

func parseArgs(args []string) ([]string, error) {
	result := []string{}
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") && len(arg) > 1 {
			for i, c := range arg[1:] {
				if string(c) == "=" {
					if i < 1 {
						return nil, errors.New("invalid input with '-' and '=' flag")
					}
					result[len(result)-1] = result[len(result)-1] + arg[i+1:]
					break
				} else if singleAlphaLetterRegxp.MatchString(string(c)) {
					result = append(result, "-"+string(c))
				} else {
					return nil, errors.Errorf("invalid input %v in flag", string(c))
				}
			}
		} else {
			result = append(result, arg)
		}
	}
	return result, nil
}

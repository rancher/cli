package main

import (
	"os"

	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/cli/cmd"
	"github.com/rancher/cli/rancher_prompt"
	"github.com/sirupsen/logrus"
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
  {{range .Commands}}{{.Name}}{{with .Aliases}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
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
	}
	app.Commands = []cli.Command{
		cmd.ClusterCommand(),
		cmd.KubectlCommand(),
		cmd.LoginCommand(),
		cmd.ProjectCommand(),
		cmd.PsCommand(),
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

package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/rancher/cli/cmd"
)

func main() {
	if err := mainErr(); err != nil {
		logrus.Fatal(err)
	}
}

func mainErr() error {
	app := cli.NewApp()
	app.Name = "rancher"
	app.Usage = "Rancher CLI"
	app.Before = func(ctx *cli.Context) error {
		if ctx.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return nil
	}
	//app.Version = version.VERSION
	app.Author = "Rancher Labs, Inc."
	app.Email = ""
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Debug logging",
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
			EnvVar: "RANCHER_HOSTNAME",
		},
		//cli.StringFlag{
		//	Name:  "rancher-file,r",
		//	Usage: "Specify an alternate Rancher compose file (default: rancher-compose.yml)",
		//},
		//cli.StringFlag{
		//	Name:  "env-file,e",
		//	Usage: "Specify a file from which to read environment variables",
		//},
	}
	app.Commands = []cli.Command{
		cmd.PsCommand(),
		cmd.RunCommand(),
		cmd.RmCommand(),
		cmd.HostCommand(),
		cmd.SSHCommand(),
		cmd.DockerCommand(),
		cmd.ScaleCommand(),
		cmd.ExecCommand(),
		cmd.ExportCommand(),
		cmd.StopCommand(),
		cmd.StartCommand(),
		cmd.RestartCommand(),
		cmd.EventsCommand(),
	}

	return app.Run(os.Args)
}

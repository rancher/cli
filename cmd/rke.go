package cmd

import (
	rke "github.com/rancher/rke/cmd"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func RKECommand() cli.Command {
	return cli.Command{
		Name:            "rke",
		Usage:           "Rancher Kubernetes Engine, Running k8s cluster everywhere",
		SkipFlagParsing: true,
		Subcommands: []cli.Command{
			rke.UpCommand(),
			rke.RemoveCommand(),
			rke.VersionCommand(),
			rke.ConfigCommand(),
			rke.EtcdCommand(),
		},
	}
}

func RKERun(args []string) error {
	app := cli.NewApp()
	app.Name = "rancher"
	app.Usage = "Rancher Kubernetes Engine, Running kubernetes cluster in the cloud"
	app.Before = func(ctx *cli.Context) error {
		if ctx.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return nil
	}
	app.Author = "Rancher Labs, Inc."
	app.Email = ""
	app.Commands = []cli.Command{
		RKECommand(),
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug,d",
			Usage: "Debug logging",
		},
	}

	a := append([]string{"rancher", "rke"}, args...)
	return app.Run(a)
}

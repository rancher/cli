package cmd

import (
	"io/ioutil"

	"github.com/rancher/cli/cliclient"
	"github.com/rancher/types/client/management/v3"
	"github.com/urfave/cli"
)

func UpCommand() cli.Command {
	return cli.Command{
		Name:   "up",
		Usage:  "apply compose config",
		Action: defaultAction(apply),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "file,f",
				Usage: "The location of compose config file",
			},
		},
	}
}

func apply(ctx *cli.Context) error {
	cf, err := lookupConfig(ctx)
	if err != nil {
		return err
	}
	c, err := cliclient.NewManagementClient(cf)
	if err != nil {
		return err
	}

	filePath := ctx.String("file")
	compose, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	globalComposeConfig := &client.ComposeConfig{
		RancherCompose: string(compose),
	}
	if _, err := c.ManagementClient.ComposeConfig.Create(globalComposeConfig); err != nil {
		return err
	}
	return nil
}

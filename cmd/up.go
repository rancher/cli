package cmd

import (
	"github.com/docker/libcompose/project"
	rancherApp "github.com/rancher/rancher-compose/app"
	"github.com/urfave/cli"
)

func UpCommand() cli.Command {
	factory := &projectFactory{}
	return rancherApp.UpCommand(factory)
}

type projectFactory struct {
}

func (p *projectFactory) Create(c *cli.Context) (project.APIProject, error) {
	factory := &rancherApp.ProjectFactory{}

	config, err := lookupConfig(c)
	if err != nil {
		return nil, err
	}

	url, err := config.EnvironmentURL()
	if err != nil {
		return nil, err
	}

	c.GlobalSet("url", url)
	c.GlobalSet("access-key", config.AccessKey)
	c.GlobalSet("secret-key", config.SecretKey)
	c.GlobalSet("project-name", c.GlobalString("stack"))

	return factory.Create(c)
}

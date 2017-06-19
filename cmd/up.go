package cmd

import (
	"github.com/pkg/errors"
	"github.com/rancher/go-rancher/v2"
	"github.com/rancher/rancher-compose-executor/app"
	"github.com/rancher/rancher-compose-executor/project"
	"github.com/urfave/cli"
)

func UpCommand() cli.Command {
	factory := &projectFactory{}
	cmd := app.UpCommand(factory)
	cmd.Flags = append(cmd.Flags, []cli.Flag{
		cli.StringFlag{
			Name:  "rancher-file",
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
	}...)
	return cmd
}

type projectFactory struct {
}

func (p *projectFactory) Create(c *cli.Context) (*project.Project, error) {
	config, err := lookupConfig(c)
	if err != nil {
		return nil, err
	}

	url, err := config.EnvironmentURL()
	if err != nil {
		return nil, err
	}

	rc, err := GetClient(c)
	if err != nil {
		return nil, err
	}

	w, err := NewWaiter(c)
	if err != nil {
		return nil, err
	}

	// from config
	c.GlobalSet("url", url)
	c.GlobalSet("access-key", config.AccessKey)
	c.GlobalSet("secret-key", config.SecretKey)

	// copy from flags
	c.GlobalSet("rancher-file", c.String("rancher-file"))
	c.GlobalSet("env-file", c.String("env-file"))
	c.GlobalSet("project-name", c.String("stack"))
	for _, f := range c.StringSlice("file") {
		c.GlobalSet("file", f)
	}

	factory := &app.RancherProjectFactory{}
	proj, err := factory.Create(c)
	if err != nil {
		return nil, err
	}
	stacks, err := rc.Stack.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"name": proj.Name,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(stacks.Data) == 0 {
		return nil, errors.Errorf("can't find the stack with name %s", proj.Name)
	}
	w.Add(stacks.Data[0].Id)
	if err := w.Wait(); err != nil {
		return nil, err
	}
	return proj, nil
}

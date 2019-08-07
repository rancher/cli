package cmd

import (
	"github.com/rancher/cli/cliclient"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func ContextCommand() cli.Command {
	return cli.Command{
		Name:  "context",
		Usage: "Operations for the context",
		Description: `Switch or view context. A context is the server->cluster->project currently in focus.
`,
		Subcommands: []cli.Command{
			{
				Name:  "switch",
				Usage: "Switch to a new context",
				Description: `
The project arg is optional, if not passed in a list of available projects will 
be displayed and one can be selected. If only one project is available it will 
be automatically selected.
`,
				ArgsUsage: "[PROJECT_ID/PROJECT_NAME]",
				Action:    contextSwitch,
			},
			{
				Name:   "current",
				Usage:  "Display the current context",
				Action: loginContext,
			},
		},
	}
}

func contextSwitch(ctx *cli.Context) error {
	cf, err := loadConfig(ctx)
	if err != nil {
		return err
	}

	server := cf.FocusedServer()
	c, err := cliclient.NewManagementClient(server)
	if err != nil {
		return err
	}

	var projectID string

	if ctx.NArg() == 0 {
		projectID, err = getProjectContext(ctx, c)
		if err != nil {
			return nil
		}
	} else {
		resource, err := Lookup(c, ctx.Args().First(), "project")
		if err != nil {
			return err
		}
		projectID = resource.ID
	}

	project, err := c.ManagementClient.Project.ByID(projectID)
	if err != nil {
		return nil
	}

	logrus.Infof("Setting new context to project %s", project.Name)

	server.Project = project.ID

	err = cf.Write()
	if err != nil {
		return err
	}

	return nil
}

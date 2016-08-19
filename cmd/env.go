package cmd

import (
	"fmt"
	"strings"

	"github.com/rancher/go-rancher/client"
	"github.com/urfave/cli"
)

func EnvCommand() cli.Command {
	return cli.Command{
		Name:      "environment",
		ShortName: "env",
		Usage:     "Interact with environments",
		Action:    defaultAction(envLs),
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "ls",
				Usage:       "List environments",
				Description: "\nWith an account API key, all environments in Rancher will be listed. If you are using an environment API key, it will only list the environment of the API key. \n\nExample:\n\t$ rancher env ls\n",
				ArgsUsage:   "None",
				Action:      envLs,
				Flags: []cli.Flag{
					listAllFlag(),
					cli.BoolFlag{
						Name:  "quiet,q",
						Usage: "Only display IDs",
					},
					cli.StringFlag{
						Name:  "format",
						Usage: "'json' or Custom format: {{.Id}} {{.Name}}",
					},
				},
			},
			cli.Command{
				Name:        "create",
				Usage:       "Create an environment",
				Description: "\nBy default, an environment with cattle orchestration framework will be created. This command only works for Account API keys.\n\nExample:\n\t$ rancher env create newEnv\n\t$ rancher env create -o kubernetes newK8sEnv\n\t$ rancher env create -o mesos newMesosEnv\n\t$ rancher env create -o swarm newSwarmEnv\n",
				ArgsUsage:   "[NEWENVNAME...]",
				Action:      envCreate,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "orchestration,o",
						Usage: "Orchestration framework",
					},
				},
			},
			cli.Command{
				Name:        "rm",
				Usage:       "Remove environment(s)",
				Description: "\nExample:\n\t$ rancher env rm 1a5\n\t$ rancher env rm newEnv\n",
				ArgsUsage:   "[ENVID ENVNAME...]",
				Action:      envRm,
				Flags:       []cli.Flag{},
			},
			cli.Command{
				Name:        "update",
				Usage:       "Update environment",
				Description: "\nChange the orchestration framework of the environment. This command only works for Account API keys.\n\nExample:\n\t$ rancher env update -o kubernetes 1a5\n\t$ rancher env update -o cattle Default\n\t$ rancher env update -o swarm 1a5\n\t$ rancher env update -o mesos 1a5\n",
				ArgsUsage:   "[ENVID ENVNAME...]",
				Action:      envUpdate,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "orchestration,o",
						Usage: "Orchestration framework",
					},
				},
			},
			cli.Command{
				Name:  "deactivate",
				Usage: "Deactivate environment(s)",
				Description: `
Deactivate an environment by ID or name

Example:
	$ rancher env deactivate 1a5
	$ rancher env deactivate Default
`,
				ArgsUsage: "[ID NAME...]",
				Action:    envDeactivate,
				Flags:     []cli.Flag{},
			},
			cli.Command{
				Name:  "activate",
				Usage: "Activate environment(s)",
				Description: `
Activate an environment by ID or name

Example:
	$ rancher env activate 1a5
	$ rancher env activate Default
`,
				ArgsUsage: "[ID NAME...]",
				Action:    envActivate,
				Flags:     []cli.Flag{},
			},
		},
	}
}

type EnvData struct {
	ID            string
	Environment   *client.Project
	Orchestration string
}

func NewEnvData(project client.Project) *EnvData {
	orch := "Cattle"

	switch {
	case project.Swarm:
		orch = "Swarm"
	case project.Mesos:
		orch = "Mesos"
	case project.Kubernetes:
		orch = "Kubernetes"
	}

	return &EnvData{
		ID:            project.Id,
		Environment:   &project,
		Orchestration: orch,
	}
}

func envRm(ctx *cli.Context) error {
	c, err := GetRawClient(ctx)
	if err != nil {
		return err
	}

	return forEachResourceWithClient(c, ctx, []string{"project"}, func(c *client.RancherClient, resource *client.Resource) (string, error) {
		return resource.Id, c.Delete(resource)
	})
}

func envUpdate(ctx *cli.Context) error {
	c, err := GetRawClient(ctx)
	if err != nil {
		return err
	}

	if ctx.NArg() < 1 {
		return cli.NewExitError("Environment name/id is required as the first argument", 1)
	}

	orch := ctx.String("orchestration")
	if orch == "" {
		return nil
	}

	env, err := LookupEnvironment(c, ctx.Args()[0])
	if err != nil {
		return err
	}

	data := map[string]interface{}{}
	setFields(ctx, data)

	var newEnv client.Project

	err = c.Update("project", &env.Resource, data, &newEnv)
	if err != nil {
		return err
	}

	fmt.Println(env.Id)
	return nil
}

func envCreate(ctx *cli.Context) error {
	c, err := GetRawClient(ctx)
	if err != nil {
		return err
	}

	name := RandomName()
	if ctx.NArg() > 0 {
		name = ctx.Args()[0]
	}

	data := map[string]interface{}{
		"name": name,
	}

	setFields(ctx, data)

	var newEnv client.Project
	if err := c.Create("project", data, &newEnv); err != nil {
		return err
	}

	fmt.Println(newEnv.Id)
	return nil
}

func setFields(ctx *cli.Context, data map[string]interface{}) {
	orch := strings.ToLower(ctx.String("orchestration"))

	data["swarm"] = false
	data["kubernetes"] = false
	data["mesos"] = false

	if orch == "k8s" {
		orch = "kubernetes"
	}

	data[orch] = true
}

func envLs(ctx *cli.Context) error {
	c, err := GetRawClient(ctx)
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "Environment.Name"},
		{"ORCHESTRATION", "Orchestration"},
		{"STATE", "Environment.State"},
		{"CREATED", "Environment.Created"},
	}, ctx)
	defer writer.Close()

	collection := client.ProjectCollection{}
	listOpts := defaultListOpts(ctx)
	if err = c.List("project", listOpts, &collection); err != nil {
		return err
	}

	for _, item := range collection.Data {
		writer.Write(NewEnvData(item))
	}

	return writer.Err()
}

func envDeactivate(ctx *cli.Context) error {
	c, err := GetRawClient(ctx)
	if err != nil {
		return err
	}

	return forEachResourceWithClient(c, ctx, []string{"project"}, func(c *client.RancherClient, resource *client.Resource) (string, error) {
		action, err := pickAction(resource, "deactivate")
		if err != nil {
			return "", err
		}
		return resource.Id, c.Action(resource.Type, action, resource, nil, resource)
	})
}

func envActivate(ctx *cli.Context) error {
	c, err := GetRawClient(ctx)
	if err != nil {
		return err
	}

	return forEachResourceWithClient(c, ctx, []string{"project"}, func(c *client.RancherClient, resource *client.Resource) (string, error) {
		action, err := pickAction(resource, "activate")
		if err != nil {
			return "", err
		}
		return resource.Id, c.Action(resource.Type, action, resource, nil, resource)
	})
}

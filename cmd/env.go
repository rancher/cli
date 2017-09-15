package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/rancher/go-rancher/v3"
	"github.com/urfave/cli"
)

func EnvCommand() cli.Command {
	envLsFlags := []cli.Flag{
		listAllFlag(),
		cli.BoolFlag{
			Name:  "quiet,q",
			Usage: "Only display IDs",
		},
		cli.StringFlag{
			Name:  "format",
			Usage: "'json' or Custom format: '{{.ID}} {{.Environment.Name}}'",
		},
	}

	return cli.Command{
		Name:      "environment",
		ShortName: "env",
		Usage:     "Interact with environments",
		Action:    defaultAction(envLs),
		Flags:     envLsFlags,
		Subcommands: []cli.Command{
			{
				Name:        "ls",
				Usage:       "List environments",
				Description: "\nWith an account API key, all environments in Rancher will be listed. If you are using an environment API key, it will only list the environment of the API key. \n\nExample:\n\t$ rancher env ls\n",
				ArgsUsage:   "None",
				Action:      envLs,
				Flags:       envLsFlags,
			},
			{
				Name:  "create",
				Usage: "Create an environment",
				Description: `
By default, an environment with cattle orchestration framework will be created. This command only works with Account API keys.

Example:

	$ rancher env create newEnv

To add an orchestration framework do TODO
	$ rancher env create -t kubernetes newK8sEnv
	$ rancher env create -t mesos newMesosEnv
	$ rancher env create -t swarm newSwarmEnv
`,
				ArgsUsage: "[NEWENVNAME...]",
				Action:    envCreate,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "cluster,c",
						Usage: "Cluster name to create the environment",
						Value: "Default",
					},
				},
			},
			{
				Name:        "rm",
				Usage:       "Remove environment(s)",
				Description: "\nExample:\n\t$ rancher env rm 1a5\n\t$ rancher env rm newEnv\n",
				ArgsUsage:   "[ENVID ENVNAME...]",
				Action:      envRm,
				Flags:       []cli.Flag{},
			},
			{
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
			{
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
			{
				Name:  "switch",
				Usage: "Switch environment(s)",
				Description: `
Switch current environment to others,

Example:
	$ rancher env switch 1a5
	$ rancher env switch Default
`,
				ArgsUsage: "[ID NAME...]",
				Action:    envSwitch,
				Flags:     []cli.Flag{},
			},
		},
	}
}

type EnvData struct {
	ID          string
	Environment *client.Project
	Current     string
	Name        string
}

func NewEnvData(project client.Project, current bool, name string) *EnvData {
	marked := ""
	if current {
		marked = "   *"
	}
	return &EnvData{
		ID:          project.Id,
		Environment: &project,
		Current:     marked,
		Name:        name,
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

func envCreate(ctx *cli.Context) error {
	c, err := GetRawClient(ctx)
	if err != nil {
		return err
	}

	name := RandomName()
	if ctx.NArg() > 0 {
		name = ctx.Args()[0]
	}
	clusters, err := c.Cluster.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"name":         ctx.String("cluster"),
			"removed_null": true,
		},
	})
	if err != nil {
		return err
	}

	if len(clusters.Data) == 0 {
		return errors.Errorf("can't find cluster with name %v", ctx.String("cluster"))
	}
	data := map[string]interface{}{
		"name":      name,
		"clusterId": clusters.Data[0].Id,
	}

	var newEnv client.Project
	if err := c.Create("project", data, &newEnv); err != nil {
		return err
	}

	fmt.Println(newEnv.Id)
	return nil
}

func envLs(ctx *cli.Context) error {
	c, err := GetRawClient(ctx)
	if err != nil {
		return err
	}
	config, err := lookupConfig(ctx)
	if err != nil {
		return err
	}
	currentEnvID := config.Environment

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"CLUSTER/NAME", "Name"},
		{"STATE", "Environment.State"},
		{"CREATED", "Environment.Created"},
		{"CURRENT", "Current"},
	}, ctx)
	defer writer.Close()

	listOpts := defaultListOpts(ctx)
	listOpts.Filters["all"] = true
	collection, err := c.Project.List(listOpts)
	if err != nil {
		return err
	}

	for _, item := range collection.Data {
		current := false
		if item.Id == currentEnvID {
			current = true
		}
		clusterName := ""
		if item.ClusterId != "" {
			cluster, err := c.Cluster.ById(item.ClusterId)
			if err != nil {
				return err
			}
			clusterName = cluster.Name
		}
		name := item.Name
		if clusterName != "" {
			name = fmt.Sprintf("%s/%s", clusterName, name)
		}
		writer.Write(NewEnvData(item, current, name))
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

func envSwitch(ctx *cli.Context) error {
	c, err := GetRawClient(ctx)
	if err != nil {
		return err
	}

	if ctx.NArg() == 0 {
		return cli.ShowCommandHelp(ctx, "env")
	}
	envID := ""
	name := ctx.Args()[0]
	if env, err := c.Project.ById(name); err == nil && env != nil && env.Id == name {
		envID = name
	} else {
		if envs, err := c.Project.List(&client.ListOpts{
			Filters: map[string]interface{}{
				"name": name,
			},
		}); err == nil {
			if len(envs.Data) == 1 {
				envID = envs.Data[0].Id
			} else if len(envs.Data) > 1 {
				names := []string{}
				for _, item := range envs.Data {
					names = append(names, fmt.Sprintf("%s(%s/%s)", item.Name, item.ClusterId, item.Id))
				}
				idx := selectFromList("Found multiple environments in different clusters:", names)
				envID = envs.Data[idx].Id
			}
		}
	}
	if envID == "" {
		return cli.NewExitError("Error: can't find associated environment", 1)
	}
	config, err := lookupConfig(ctx)
	if err != nil {
		return err
	}
	config.Environment = envID
	err = config.Write()
	if err != nil {
		return err
	}
	return envLs(ctx)
}

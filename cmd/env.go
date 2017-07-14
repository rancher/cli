package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/rancher/go-rancher/v2"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
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
			Usage: "'json' or Custom format: {{.Id}} {{.Name}}",
		},
	}

	return cli.Command{
		Name:      "environment",
		ShortName: "env",
		Usage:     "Interact with environments",
		Action:    defaultAction(envLs),
		Flags:     envLsFlags,
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "ls",
				Usage:       "List environments",
				Description: "\nWith an account API key, all environments in Rancher will be listed. If you are using an environment API key, it will only list the environment of the API key. \n\nExample:\n\t$ rancher env ls\n",
				ArgsUsage:   "None",
				Action:      envLs,
				Flags:       envLsFlags,
			},
			cli.Command{
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
						Name:  "template,t",
						Usage: "Environment template to create from",
						Value: "Cattle",
					},
				},
			},
			cli.Command{
				Name:      "templates",
				ShortName: "template",
				Usage:     "Interact with environment templates",
				Action:    defaultAction(envTemplateLs),
				Flags:     envLsFlags,
				Subcommands: []cli.Command{
					cli.Command{
						Name:      "export",
						Usage:     "Export an environment template to STDOUT",
						ArgsUsage: "[TEMPLATEID TEMPLATENAME...]",
						Action:    envTemplateExport,
						Flags:     []cli.Flag{},
					},
					cli.Command{
						Name:      "import",
						Usage:     "Import an environment template to from file",
						ArgsUsage: "[FILE FILE...]",
						Action:    envTemplateImport,
						Flags: []cli.Flag{
							cli.BoolFlag{
								Name:  "public",
								Usage: "Make template public",
							},
						},
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
			cli.Command{
				Name:  "select",
				Usage: "Select environment",
				Description: `
Interactively select an environment

Example:
	$ rancher env select
`,
				ArgsUsage: "None",
				Action:    envSelect,
				Flags:     []cli.Flag{},
			},
		},
	}
}

type EnvData struct {
	ID          string
	Environment *client.Project
}

type TemplateData struct {
	ID              string
	ProjectTemplate *client.ProjectTemplate
}

func NewEnvData(project client.Project) *EnvData {
	return &EnvData{
		ID:          project.Id,
		Environment: &project,
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

	data := map[string]interface{}{
		"name": name,
	}

	template := ctx.String("template")
	if template != "" {
		template, err := Lookup(c, template, "projectTemplate")
		if err != nil {
			return err
		}
		data["projectTemplateId"] = template.Id
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

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "Environment.Name"},
		{"ORCHESTRATION", "Environment.Orchestration"},
		{"STATE", "Environment.State"},
		{"CREATED", "Environment.Created"},
	}, ctx)
	defer writer.Close()

	collection, err := c.Project.List(defaultListOpts(ctx))
	if err != nil {
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

func envTemplateLs(ctx *cli.Context) error {
	c, err := GetRawClient(ctx)
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "ProjectTemplate.Name"},
		{"DESC", "ProjectTemplate.Description"},
	}, ctx)
	defer writer.Close()

	collection, err := c.ProjectTemplate.List(defaultListOpts(ctx))
	if err != nil {
		return err
	}

	for _, item := range collection.Data {
		writer.Write(TemplateData{
			ID:              item.Id,
			ProjectTemplate: &item,
		})
	}

	return writer.Err()
}

func envTemplateImport(ctx *cli.Context) error {
	c, err := GetRawClient(ctx)
	if err != nil {
		return err
	}

	w, err := NewWaiter(ctx)
	if err != nil {
		return err
	}

	for _, file := range ctx.Args() {
		template := client.ProjectTemplate{
			IsPublic: ctx.Bool("public"),
		}
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		if err := yaml.Unmarshal(content, &template); err != nil {
			return err
		}

		created, err := c.ProjectTemplate.Create(&template)
		if err != nil {
			return err
		}

		w.Add(created.Id)
	}

	return w.Wait()
}

func envTemplateExport(ctx *cli.Context) error {
	c, err := GetRawClient(ctx)
	if err != nil {
		return err
	}

	for _, name := range ctx.Args() {
		r, err := Lookup(c, name, "projectTemplate")
		if err != nil {
			return err
		}

		template, err := c.ProjectTemplate.ById(r.Id)
		if err != nil {
			return err
		}

		stacks := []map[string]interface{}{}
		for _, s := range template.Stacks {
			data := map[string]interface{}{
				"name": s.Name,
			}
			if s.TemplateId != "" {
				data["template_id"] = s.TemplateId
			}
			if s.TemplateVersionId != "" {
				data["template_version_id"] = s.TemplateVersionId
			}
			if len(s.Answers) > 0 {
				data["answers"] = s.Answers
			}
			stacks = append(stacks, data)
		}

		data := map[string]interface{}{
			"name":        template.Name,
			"description": template.Description,
			"stacks":      stacks,
		}

		content, err := yaml.Marshal(&data)
		if err != nil {
			return err
		}

		_, err = os.Stdout.Write(content)
		if err != nil {
			return err
		}
	}

	return nil
}

func envSelect(ctx *cli.Context) error {
	config, err := lookupConfig(ctx)
	if err != nil && err != errNoURL {
		return err
	}

	c, err := client.NewRancherClient(&client.ClientOpts{
		Url:       config.URL,
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
	})
	if err != nil {
		return err
	}

	if schema, ok := c.GetSchemas().CheckSchema("schema"); ok {
		// Normalize URL
		config.URL = schema.Links["collection"]
	} else {
		return fmt.Errorf("Failed to find schema URL")
	}

	c, err = client.NewRancherClient(&client.ClientOpts{
		Url:       config.URL,
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
	})
	if err != nil {
		return err
	}

	project, err := GetEnvironment("", c)
	if err != errNoEnv {
		if err != nil {
			return err
		}
		config.Environment = project.Id
	}

	return config.Write()
}

package cmd

import (
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher/client"
	"github.com/urfave/cli"
)

func EnvCommand() cli.Command {
	return cli.Command{
		Name:      "environment",
		ShortName: "env",
		Usage:     "Interact with environments",
		Action:    errorWrapper(envLs),
		Subcommands: []cli.Command{
			cli.Command{
				Name:   "ls",
				Usage:  "list environments",
				Action: errorWrapper(envLs),
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "quiet,q",
						Usage: "Only display IDs",
					},
					cli.StringFlag{
						Name:  "format",
						Usage: "'json' or Custom format: {{.Id}} {{.Name}",
					},
				},
			},
			cli.Command{
				Name:   "create",
				Usage:  "create environment",
				Action: errorWrapper(envCreate),
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "orchestration,o",
						Usage: "Name",
					},
				},
			},
			cli.Command{
				Name:   "rm",
				Usage:  "Remove environment(s) by ID",
				Action: errorWrapper(envRm),
			},
			cli.Command{
				Name:   "update",
				Usage:  "Update environment",
				Action: errorWrapper(envUpdate),
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "orchestration,o",
						Usage: "Orchestration framework",
					},
				},
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

	var lastErr error
	for _, id := range ctx.Args() {
		env, err := Lookup(c, id, "account")
		if err != nil {
			logrus.Errorf("Failed to delete %s: %v", id, err)
			lastErr = err
			continue
		}
		if err := c.Delete(env); err != nil {
			logrus.Errorf("Failed to delete %s: %v", id, err)
			lastErr = err
			continue
		}
		fmt.Println(env.Id)
	}

	return lastErr
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
	err = c.List("account", &client.ListOpts{
		Filters: map[string]interface{}{
			"kind": "project",
		},
	}, &collection)
	if err != nil {
		return err
	}

	for _, item := range collection.Data {
		writer.Write(NewEnvData(item))
	}

	return writer.Err()
}

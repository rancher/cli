package cmd

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/go-rancher/v2"
	"github.com/urfave/cli"
)

func StackCommand() cli.Command {
	stackLsFlags := []cli.Flag{
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
		Name:      "stacks",
		ShortName: "stack",
		Usage:     "Operations on stacks",
		Action:    defaultAction(stackLs),
		Flags:     stackLsFlags,
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "ls",
				Usage:       "List stacks",
				Description: "\nLists all stacks in the current $RANCHER_ENVIRONMENT. Use `--env <envID>` or `--env <envName>` to select a different environment.\n\nExample:\n\t$ rancher stacks ls\n\t$ rancher --env 1a5 stacks ls\n",
				ArgsUsage:   "None",
				Action:      stackLs,
				Flags:       stackLsFlags,
			},
			cli.Command{
				Name:        "create",
				Usage:       "Create a stacks",
				Description: "\nCreate all stack in the current $RANCHER_ENVIRONMENT. Use `--env <envID>` or `--env <envName>` to select a different environment.\n\nExample:\n\t$ rancher stacks create\n\t$ rancher --env 1a5 stacks ls\n",
				ArgsUsage:   "None",
				Action:      stackCreate,
				Flags: []cli.Flag{
					cli.BoolTFlag{
						Name:  "start",
						Usage: "Start stack on create",
					},
					cli.BoolFlag{
						Name:  "system",
						Usage: "Create a system stack",
					},
					cli.BoolFlag{
						Name:  "quiet,q",
						Usage: "Only display IDs",
					},
					cli.StringFlag{
						Name:  "docker-compose,f",
						Usage: "Docker Compose file",
						Value: "docker-compose.yml",
					},
					cli.StringFlag{
						Name:  "rancher-compose,r",
						Usage: "Rancher Compose file",
						Value: "rancher-compose.yml",
					},
					cli.StringFlag{
						Name:  "answers,a",
						Usage: "Answers files",
						Value: "answers",
					},
				},
			},
		},
	}
}

type StackData struct {
	ID      string
	Catalog string
	Stack   client.Stack
	State   string
	System  bool
}

func stackLs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	collection, err := c.Stack.List(defaultListOpts(nil))
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "Stack.Name"},
		{"STATE", "State"},
		{"CATALOG", "Catalog"},
		{"SYSTEM", "System"},
		{"DETAIL", "Stack.TransitioningMessage"},
	}, ctx)

	defer writer.Close()

	for _, item := range collection.Data {
		system := strings.HasPrefix(item.ExternalId, "system://")
		if !system {
			system = strings.HasPrefix(item.ExternalId, "system-catalog://")
		}
		if !system {
			system = strings.HasPrefix(item.ExternalId, "kubernetes")
		}
		combined := item.HealthState
		if item.State != "active" || combined == "" {
			combined = item.State
		}
		writer.Write(&StackData{
			ID:      item.Id,
			Stack:   item,
			State:   combined,
			System:  system,
			Catalog: item.ExternalId,
		})
	}

	return writer.Err()
}

func getFile(name string) (string, error) {
	if name == "" {
		return "", nil
	}
	bytes, err := ioutil.ReadFile(name)
	if err == os.ErrNotExist {
		return "", nil

	}
	return string(bytes), err
}

func parseAnswers(ctx *cli.Context) (map[string]interface{}, error) {
	answers := map[string]interface{}{}
	answersFile, err := getFile(ctx.String("answers"))
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewBuffer([]byte(answersFile)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 1 {
			answers[parts[0]] = ""
		} else {
			answers[parts[0]] = parts[1]
		}
	}

	return answers, scanner.Err()
}

func stackCreate(ctx *cli.Context) error {
	c, err := GetClient(ctx)

	dockerCompose, err := getFile(ctx.String("docker-compose"))
	if err != nil {
		return err
	}
	if dockerCompose == "" {
		return errors.New("docker-compose.yml files is required")
	}

	rancherCompose, err := getFile(ctx.String("rancher-compose"))
	if err != nil {
		return err
	}

	answers, err := parseAnswers(ctx)
	if err != nil {
		return errors.Wrap(err, "reading answers")
	}

	name := RandomName()
	if len(ctx.Args()) > 0 {
		name = ctx.Args()[0]
	}

	stack, err := c.Stack.Create(&client.Stack{
		Name:           name,
		DockerCompose:  dockerCompose,
		RancherCompose: rancherCompose,
		Environment:    answers,
		System:         ctx.Bool("system"),
		StartOnCreate:  ctx.Bool("startOnCreate"),
	})
	if err != nil {
		return err
	}

	return WaitFor(ctx, stack.Id)
}

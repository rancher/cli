package cmd

import (
	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/rancher-compose-executor/lookup"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
)

func StackCommand() cli.Command {
	stackLsFlags := []cli.Flag{
		listSystemFlag(),
		cli.BoolFlag{
			Name:  "quiet,q",
			Usage: "Only display IDs",
		},
		cli.StringFlag{
			Name:  "format",
			Usage: "'json' or Custom format: '{{.ID}} {{.Stack.Name}}'",
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
						Name:  "system,s",
						Usage: "Create a system stack",
					},
					cli.BoolFlag{
						Name:  "empty,e",
						Usage: "Create an empty stack",
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
					//cli.StringFlag{
					//	Name:  "answers,a",
					//	Usage: "Answers files",
					//	Value: "answers",
					//},
				},
			},
		},
	}
}

type StackData struct {
	ID           string
	Catalog      string
	Stack        client.Stack
	State        string
	ServiceCount int
}

func stackLs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	collection, err := c.Stack.List(defaultListOpts(ctx))
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "Stack.Name"},
		{"STATE", "State"},
		{"CATALOG", "Catalog"},
		{"SERVICES", "ServiceCount"},
		{"DETAIL", "Stack.TransitioningMessage"},
	}, ctx)

	defer writer.Close()

	for _, item := range collection.Data {
		combined := item.HealthState
		if item.State != "active" || combined == "" {
			combined = item.State
		}
		writer.Write(&StackData{
			ID:           item.Id,
			Stack:        item,
			State:        combined,
			Catalog:      item.ExternalId,
			ServiceCount: len(item.ServiceIds),
		})
	}

	return writer.Err()
}

func getFile(name string) (string, error) {
	if name == "" {
		return "", nil
	}
	bytes, err := ioutil.ReadFile(name)
	if os.IsNotExist(err) {
		return "", nil

	}
	return string(bytes), err
}

func parseAnswers(ctx *cli.Context) (map[string]interface{}, error) {
	file := ctx.String("answers")

	return lookup.ParseEnvFile(file)
}

func stackCreate(ctx *cli.Context) error {
	c, err := GetClient(ctx)

	names := []string{RandomName()}
	if len(ctx.Args()) > 0 {
		names = ctx.Args()
	}

	w, err := NewWaiter(ctx)
	if err != nil {
		return err
	}

	var lastErr error
	for _, name := range names {
		stack := &client.Stack{
			Name: name,
		}

		if !ctx.Bool("empty") {
			//var err error
			// todo: revisit
			//stack.Answers, err = parseAnswers(ctx)
			//if err != nil {
			//return errors.Wrap(err, "reading answers")
			//}
		}

		stack, err = c.Stack.Create(stack)
		if err != nil {
			lastErr = err
		}

		w.Add(stack.Id)
	}

	if lastErr != nil {
		return lastErr
	}

	return w.Wait()
}

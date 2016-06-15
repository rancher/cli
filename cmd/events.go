package cmd

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/rancher/cli/monitor"
)

func EventsCommand() cli.Command {
	return cli.Command{
		Name:   "events",
		Usage:  "Show services/containers",
		Action: events,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "format",
				Usage: "'json' or Custom format: {{.Id}} {{.Name}",
			},
		},
	}
}

func events(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	m := monitor.New(c)
	sub := m.Subscribe()
	go m.Start()

	for event := range sub.C {
		resource, _ := event.Data["resource"].(map[string]interface{})
		state, _ := resource["state"].(string)
		name, _ := resource["name"].(string)
		if name == "" {
			name = fmt.Sprintf("(%s)", event.ResourceID)
		}

		if len(state) > 0 {
			message := resource["transitioningMessage"]
			if message == nil {
				message = ""
			}
			fmt.Printf("%s %s %s %v %v\n", event.Name, event.ResourceType, name, state, message)
		}
	}

	return nil
}

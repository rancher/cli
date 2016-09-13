package cmd

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/cli/monitor"
	"github.com/rancher/go-rancher/v2"
	"github.com/urfave/cli"
)

func EventsCommand() cli.Command {
	return cli.Command{
		Name:        "events",
		Usage:       "Displays resource change events",
		Description: "\nOnly events that are actively occuring in Rancher are listed.\n",
		ArgsUsage:   "None",
		Action:      events,
		Flags: []cli.Flag{
			//cli.StringFlag{
			//	Name:  "format",
			//	Usage: "'json' or Custom format: {{.Id}} {{.Name}}",
			//},
			cli.BoolFlag{
				Name:  "reconnect,r",
				Usage: "Reconnect on error",
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
	go func() {
		if ctx.Bool("reconnect") {
			for {
				if err := m.Start(); err != nil {
					logrus.Error(err)
				}
				time.Sleep(time.Second)
			}
		} else {
			logrus.Fatal(m.Start())
		}
	}()

	for event := range sub.C {
		resource, _ := event.Data["resource"].(map[string]interface{})
		state, _ := resource["state"].(string)
		name, _ := resource["name"].(string)

		if len(state) > 0 {
			message := resource["transitioningMessage"]
			if message == nil {
				message = ""
			}
			fmt.Printf("%s %s %s [%s] %v\n", event.ResourceType, event.ResourceID, state, name, message)
		}
	}

	return nil
}

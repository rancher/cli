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
			cli.StringFlag{
				Name:  "format",
				Usage: "'json' or Custom format: {{.Id}} {{.Name}}",
			},
			cli.BoolFlag{
				Name:  "reconnect,r",
				Usage: "Reconnect on error",
			},
		},
	}
}

func getClientForSubscribe(ctx *cli.Context) (*client.RancherClient, error) {
	if ctx.Bool("all") {
		return GetRawClient(ctx)
	}
	return GetClient(ctx)
}

func events(ctx *cli.Context) error {
	reconnect := ctx.Bool("reconnect")

	for {
		c, err := getClientForSubscribe(ctx)
		if err != nil {
			if reconnect {
				logrus.Error(err)
				time.Sleep(time.Second)
				continue
			} else {
				return err
			}
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

		format := ctx.String("format")
		for event := range sub.C {
			if format == "" {
				resource, _ := event.Data["resource"].(map[string]interface{})
				name, _ := resource["name"].(string)

				if name == "ping" {
					continue
				}

				healthState, _ := resource["healthState"].(string)
				state, _ := resource["state"].(string)

				combined := healthState
				if state != "active" || combined == "" {
					combined = state
				}

				message, _ := resource["transitioningMessage"].(string)
				fmt.Printf("%s %s %s [%s] %v\n", event.ResourceType, event.ResourceID, combined, name, message)
			} else {
				writer := NewTableWriter(nil, ctx)
				writer.Write(event)
				if err := writer.Err(); err != nil {
					logrus.Error(err)
				}
			}
		}
	}
}

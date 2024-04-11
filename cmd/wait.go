package cmd

import (
	"fmt"
	"strings"
	"time"

	ntypes "github.com/rancher/norman/types"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	waitTypes = []string{"cluster", "app", "project", "multiClusterApp"}
)

func WaitCommand() cli.Command {
	return cli.Command{
		Name:      "wait",
		Usage:     "Wait for resources " + strings.Join(waitTypes, ", "),
		ArgsUsage: "[ID/NAME]",
		Action:    defaultAction(wait),
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "timeout",
				Usage: "Time in seconds to wait for a resource",
				Value: 120,
			},
		},
	}
}

func wait(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowCommandHelp(ctx, "wait")
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	resource, err := Lookup(c, ctx.Args().First(), waitTypes...)
	if err != nil {
		return err
	}

	mapResource := map[string]interface{}{}

	// Initial check shortcut
	err = c.ByID(resource, &mapResource)
	if err != nil {
		return err
	}

	ok, err := checkDone(resource, mapResource)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	timeout := time.After(time.Duration(ctx.Int("timeout")) * time.Second)
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("Timeout reached %v:%v transitioningMessage: %v", resource.Type, resource.ID, mapResource["transitioningMessage"])
		case <-ticker.C:
			err = c.ByID(resource, &mapResource)
			if err != nil {
				return err
			}

			ok, err := checkDone(resource, mapResource)
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
		}
	}
}

func checkDone(resource *ntypes.Resource, data map[string]interface{}) (bool, error) {
	transitioning := fmt.Sprint(data["transitioning"])
	logrus.Debugf("%s:%s transitioning=%s state=%v", resource.Type, resource.ID, transitioning,
		data["state"])

	switch transitioning {
	case "yes":
		return false, nil
	case "error":
		if data["state"] == "provisioning" {
			break
		}
		return false, fmt.Errorf("%v:%v failed, transitioningMessage: %v", resource.Type, resource.ID, data["transitioningMessage"])
	}

	return data["state"] == "active", nil
}

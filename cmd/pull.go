package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/rancher/go-rancher/v3"
	"github.com/urfave/cli"
	"time"
)

func PullCommand() cli.Command {
	return cli.Command{
		Name:        "pull",
		Usage:       "Pull images on hosts that are in the current environment. Examples: rancher pull ubuntu",
		Action:      pullImages,
		Subcommands: []cli.Command{},
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "hosts",
				Usage: "Specify which host should pull images. By default it will pull images on all the hosts in the current environment. Examples: rancher pull --hosts 1h1 --hosts 1h2 ubuntu",
			},
		},
	}
}

func pullImages(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}
	if ctx.NArg() == 0 {
		return cli.ShowCommandHelp(ctx, "")
	}
	image := ctx.Args()[0]

	hosts := ctx.StringSlice("hosts")
	if len(hosts) == 0 {
		hts, err := c.Host.List(defaultListOpts(ctx))
		if err != nil {
			return err
		}
		for _, ht := range hts.Data {
			hosts = append(hosts, ht.Id)
		}
	}
	pullTask := client.PullTask{
		Mode:    "all",
		Image:   image,
		HostIds: hosts,
	}
	task, err := c.PullTask.Create(&pullTask)
	if err != nil {
		return err
	}
	cl := getRandomColor()
	lastMsg := ""
	for {
		if task.Transitioning != "yes" {
			fmt.Printf("Finished pulling image %s\n", image)
			return nil
		}
		time.Sleep(150 * time.Millisecond)
		if task.TransitioningMessage != lastMsg {
			color.New(cl).Printf("Pulling image. Status: %s\n", task.TransitioningMessage)
			lastMsg = task.TransitioningMessage
		}
		task, err = c.PullTask.ById(task.Id)
		if err != nil {
			return err
		}
	}
}

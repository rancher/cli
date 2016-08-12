package cmd

import (
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher/client"
	"github.com/urfave/cli"
)

func VolumeCommand() cli.Command {
	return cli.Command{
		Name:      "volumes",
		ShortName: "volume",
		Usage:     "Operations on volumes",
		Action:    defaultAction(volumeLs),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "quiet,q",
				Usage: "Only display IDs",
			},
			cli.StringFlag{
				Name:  "format",
				Usage: "'json' or Custom format: {{.Id}} {{.Name}}",
			},
		},
		Subcommands: []cli.Command{
			cli.Command{
				Name:   "ls",
				Usage:  "List volumes",
				Action: volumeLs,
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "quiet,q",
						Usage: "Only display IDs",
					},
					cli.StringFlag{
						Name:  "format",
						Usage: "'json' or Custom format: {{.Id}} {{.Name}}",
					},
				},
			},
			cli.Command{
				Name:   "rm",
				Usage:  "Delete volume",
				Action: volumeRm,
			},
			cli.Command{
				Name:   "create",
				Usage:  "Create volume",
				Action: volumeCreate,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "driver",
						Usage: "Specify volume driver name",
					},
					cli.StringSliceFlag{
						Name:  "opt",
						Usage: "Set driver specific key/value options",
					},
				},
			},
		},
	}
}

type VolumeData struct {
	ID     string
	Volume client.Volume
}

func volumeLs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	collection, err := c.Volume.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"limit": -1,
		},
	})
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "Volume.Name"},
		{"STATE", "Volume.State"},
		{"DRIVER", "Volume.Driver"},
		{"DETAIL", "Volume.TransitioningMessage"},
	}, ctx)

	defer writer.Close()

	for _, item := range collection.Data {
		writer.Write(&VolumeData{
			ID:     item.Id,
			Volume: item,
		})
	}

	return writer.Err()
}

func volumeRm(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	var lastErr error
	for _, id := range ctx.Args() {
		volume, err := Lookup(c, id, "volume")
		if err != nil {
			lastErr = err
			logrus.Errorf("Failed to delete %s: %v", id, err)
			continue
		}

		if err := c.Delete(volume); err != nil {
			lastErr = err
			logrus.Errorf("Failed to delete %s: %v", id, err)
			continue
		}

		fmt.Println(volume.Id)
	}

	return lastErr
}

func volumeCreate(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		return cli.NewExitError("Volume name is required as the first argument", 1)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	newVol := &client.Volume{
		Name:       ctx.Args()[0],
		Driver:     ctx.String("driver"),
		DriverOpts: map[string]interface{}{},
	}

	for _, arg := range ctx.StringSlice("opt") {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 1 {
			newVol.DriverOpts[parts[0]] = ""
		} else {
			newVol.DriverOpts[parts[0]] = parts[1]
		}
	}

	newVol, err = c.Volume.Create(newVol)
	if err != nil {
		return err
	}

	fmt.Println(newVol.Id)
	return nil
}

package cmd

import (
	"archive/tar"
	"fmt"
	"io"
	"os"

	"github.com/codegangsta/cli"
)

func ExportCommand() cli.Command {
	return cli.Command{
		Name:   "export",
		Usage:  "Export configuration yml for a service",
		Action: exportService,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "output,o",
				Usage: "Write to a file, instead of STDOUT",
			},
		},
	}
}

func getOutput(ctx *cli.Context) (io.WriteCloser, error) {
	output := ctx.String("output")
	if output == "" {
		return os.Stdout, nil
	} else {
		return os.Create(output)
	}
}

func exportService(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	if len(ctx.Args()) != 1 {
		return fmt.Errorf("One stack name is required as an argument")
	}

	resource, err := Lookup(c, ctx.Args()[0], "environment")
	if err != nil {
		return err
	}

	env, err := c.Environment.ById(resource.Id)
	if err != nil {
		return err
	}

	config, err := c.Environment.ActionExportconfig(env, nil)
	if err != nil {
		return err
	}

	output, err := getOutput(ctx)
	if err != nil {
		return err
	}
	defer output.Close()

	archive := tar.NewWriter(output)
	defer archive.Close()

	if err := addToTar(archive, "docker-compose.yml", config.DockerComposeConfig); err != nil {
		return err
	}
	return addToTar(archive, "rancher-compose.yml", config.RancherComposeConfig)
}

func addToTar(archive *tar.Writer, name string, stringContent string) error {
	if len(stringContent) == 0 {
		return nil
	}

	content := []byte(stringContent)
	err := archive.WriteHeader(&tar.Header{
		Name:  name,
		Size:  int64(len(content)),
		Mode:  0644,
		Uname: "root",
		Gname: "root",
	})
	if err != nil {
		return err
	}

	_, err = archive.Write(content)
	return err
}

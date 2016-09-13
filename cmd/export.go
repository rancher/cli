package cmd

import (
	"archive/tar"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli"
)

func ExportCommand() cli.Command {
	return cli.Command{
		Name:        "export",
		Usage:       "Export configuration yml for a stack as a tar archive",
		Description: "\nExports the docker-compose.yml and rancher-compose.yml for the specified stack as a tar archive.\n\nExample:\n\t$ rancher export mystack > files.tar\n\t$ rancher export -o files.tar mystack\n",
		ArgsUsage:   "[STACKNAME STACKID...]",
		Action:      exportService,
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
	}
	return os.Create(output)
}

func exportService(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	if len(ctx.Args()) != 1 {
		return fmt.Errorf("One stack name or ID is required as an argument")
	}

	resource, err := Lookup(c, ctx.Args()[0], "environment")
	if err != nil {
		return err
	}

	env, err := c.Stack.ById(resource.Id)
	if err != nil {
		return err
	}

	config, err := c.Stack.ActionExportconfig(env, nil)
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

package cmd

import "github.com/urfave/cli"

func ContainerCommand() cli.Command {
	return cli.Command{
		Name:   "container",
		Usage:  "Interact with containers",
		Action: errorWrapper(containerLs),
		Subcommands: []cli.Command{
			cli.Command{
				Name:   "ls",
				Usage:  "list containers",
				Action: errorWrapper(containerLs),
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "quiet,q",
						Usage: "Only display IDs",
					},
				},
			},
		},
	}
}

func containerLs(ctx *cli.Context) error {
	client, err := GetClient(ctx)
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "Id"},
		{"NAME", "Name"},
		{"STATE", "State"},
		{"CREATED", "Created"},
		{"START COUNT", "StartCount"},
		{"CREATE INDEX", "CreateIndex"},
	}, ctx)
	defer writer.Close()

	collection, err := client.Container.List(nil)
	if err != nil {
		return err
	}

	for _, item := range collection.Data {
		writer.Write(item)
	}

	return writer.Err()
}

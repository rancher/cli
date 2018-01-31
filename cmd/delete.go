package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func DeleteCommand() cli.Command {
	return cli.Command{
		Name:    "delete",
		Aliases: []string{"rm"},
		Usage:   "Delete resources by ID",
		Action:  deleteResource,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "type",
				Usage: "type of resource to delete",
			},
		},
	}
}

func deleteResource(ctx *cli.Context) error {
	if ctx.String("type") == "" {
		return errors.New("type is required for deletes")
	}
	//c, err := GetClient(ctx)
	//if err != nil {
	//	return err
	//}
	fmt.Println("This isn't implemented yet")

	return nil
}

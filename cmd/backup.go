package cmd

import (
	"context"
	"fmt"
	"github.com/urfave/cli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BackupCommand() cli.Command {
	return cli.Command{
		Name:   "backup",
		Usage:  "Operations with catalogs",
		Action: defaultAction(backupCreate),
		//Flags:  catalogLsFlags,
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "create",
				Usage:       "Perform backup/create snapshot",
				Description: "\nCreate a backup of Rancher MCM",
				ArgsUsage:   "None",
				Action:      backupCreate,
				//Flags:       catalogLsFlags,
			},
		},
	}
}

func backupCreate(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	CRDs, err := c.CRDClient.ApiextensionsV1beta1().CustomResourceDefinitions().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("\nCRDs to backup for Rancher MCM: %v\n", CRDs.Items)
	return nil
}

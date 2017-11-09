package cmd

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher/v2"
	"github.com/urfave/cli"
)

func SecretCommand() cli.Command {
	secretLsFlags := []cli.Flag{
		cli.BoolFlag{
			Name:  "quiet,q",
			Usage: "Only display IDs",
    },
    cli.BoolFlag{
			Name:  "json,j",
			Usage: "Use json format as context",
		},
		cli.StringFlag{
			Name:  "format",
			Usage: "'json' or Custom format: {{.Id}} {{.Name}}",
		},
	}

	return cli.Command{
		Name:      "secrets",
		ShortName: "secret",
		Usage:     "Operations on secrets",
		Action:    defaultAction(secretLs),
		Flags:     secretLsFlags,
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "ls",
				Usage:       "List secrets",
				Description: "\nLists all secrets in the current $RANCHER_ENVIRONMENT. Use `--env <envID>` or `--env <envName>` to select a different environment.\n\nExample:\n\t$ rancher secrets ls\n\t$ rancher --env 1a5 secrets ls\n",
				ArgsUsage:   "None",
				Action:      secretLs,
				Flags:       secretLsFlags,
			},
			cli.Command{
				Name:        "create",
				Usage:       "Create a secret",
				Description: "\nCreate all secret in the current $RANCHER_ENVIRONMENT. Use `--env <envID>` or `--env <envName>` to select a different environment.\n\nExample:\n\t$ rancher secret create my-name file-with-secret\n",
				ArgsUsage:   "NAME [FILE|-]",
				Action:      secretCreate,
				Flags:       []cli.Flag{},
			},
		},
	}
}

type SecretData struct {
	ID     string
	Secret client.Secret
}

func secretLs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	collection, err := c.Secret.List(defaultListOpts(ctx))
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "Secret.Name"},
		{"CREATED", "Secret.Created"},
	}, ctx)

	defer writer.Close()

	for _, item := range collection.Data {
		writer.Write(&SecretData{
			ID:     item.Id,
			Secret: item,
		})
	}

	return writer.Err()
}

func secretCreate(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	w, err := NewWaiter(ctx)
	if err != nil {
		return err
	}

	if ctx.NArg() != 2 {
		return fmt.Errorf("both NAME and FILE|- are required")
	}

	name, file := ctx.Args()[0], ctx.Args()[1]
	var input io.Reader

	if file == "-" {
		input = os.Stdin
	} else {
		input, err = os.Open(file)
		if os.IsNotExist(err) {
			logrus.Errorf("Failed to find %s, argument must be a file or -", file)
		}
	}

	content, err := ioutil.ReadAll(input)
	if err != nil {
		return err
	}

	secret, err := c.Secret.Create(&client.Secret{
		Name:  name,
		Value: base64.StdEncoding.EncodeToString(content),
	})
	if err != nil {
		return err
	}

	w.Add(secret.Id)
	return w.Wait()
}

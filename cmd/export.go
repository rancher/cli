package cmd

import (
	"archive/tar"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher/v3"
	"github.com/urfave/cli"
)

func ExportCommand() cli.Command {
	return cli.Command{
		Name:  "export",
		Usage: "Export configuration yml for a stack as a tar archive or to local files",
		Description: `
Exports the docker-compose.yml and rancher-compose.yml for the specified stack as a tar archive.

Example:
    $ rancher export mystack
	$ rancher export -f files.tar mystack
	# Export the entire environment, including system stacks
    $ rancher export --system mystack
`,
		ArgsUsage: "[STACKNAME STACKID...]",
		Action:    exportService,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "file,f",
				Usage: "Write to a file, instead of local files, use - to write to STDOUT",
			},
			cli.BoolFlag{
				Name:  "system,s",
				Usage: "If exporting the entire environment, include system",
			},
		},
	}
}

func getOutput(ctx *cli.Context) (io.WriteCloser, error) {
	output := ctx.String("file")
	if output == "" {
		return nil, nil
	} else if output == "-" {
		return os.Stdout, nil
	}
	return os.Create(output)
}

func getStackNames(ctx *cli.Context, c *client.RancherClient) ([]string, error) {
	stacks, err := c.Stack.List(defaultListOpts(ctx))
	if err != nil {
		return nil, err
	}

	result := []string{}
	for _, stack := range stacks.Data {
		result = append(result, stack.Name)
	}

	return result, nil
}

func exportService(ctx *cli.Context) error {
	var err error
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	names := ctx.Args()
	if len(names) == 0 {
		names, err = getStackNames(ctx, c)
		if err != nil {
			return err
		}
	}

	var archive *tar.Writer
	output, err := getOutput(ctx)
	if err != nil {
		return err
	}
	if output != nil {
		defer output.Close()
		archive = tar.NewWriter(output)
		defer archive.Close()
	}

	for _, name := range names {
		resource, err := Lookup(c, name, "stack")
		if err != nil {
			return err
		}

		stack, err := c.Stack.ById(resource.Id)
		if err != nil {
			return err
		}

		if _, ok := stack.Actions["exportconfig"]; !ok {
			continue
		}

		config, err := c.Stack.ActionExportconfig(stack, nil)
		if err != nil {
			return err
		}

		if err := addToTar(archive, stack.Name, "compose.yml", config.Templates["compose.yml"]); err != nil {
			return err
		}
		if len(config.Actions) > 0 {
			if err := addToTar(archive, stack.Name, "answers", marshalAnswers(config.Actions)); err != nil {
				return err
			}
		}
	}

	return nil
}

func marshalAnswers(answers map[string]string) string {
	buf := &bytes.Buffer{}
	for k, v := range answers {
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(v)
		buf.WriteString("\n")
	}
	return buf.String()
}

func addToTar(archive *tar.Writer, stackName, name string, stringContent string) error {
	if len(stringContent) == 0 {
		return nil
	}

	f := filepath.Join(stackName, name)
	if archive == nil {
		err := os.MkdirAll(stackName, 0755)
		if err != nil {
			return err
		}
		logrus.Infof("Creating %s", f)
		return ioutil.WriteFile(f, []byte(stringContent), 0600)
	}

	content := []byte(stringContent)
	err := archive.WriteHeader(&tar.Header{
		Name:  f,
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

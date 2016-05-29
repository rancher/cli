package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"text/template"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/rancher/go-rancher/client"
)

func GetClient(ctx *cli.Context) (*client.RancherClient, error) {
	url := ctx.GlobalString("url")
	accessKey := ctx.GlobalString("access-key")
	secretKey := ctx.GlobalString("secret-key")

	if url == "" {
		return nil, fmt.Errorf("RANCHER_URL environment or --url is not set")
	}

	return client.NewRancherClient(&client.ClientOpts{
		Url:       url,
		AccessKey: accessKey,
		SecretKey: secretKey,
	})
}

func GetEnvironment(c *client.RancherClient) (*client.Project, error) {
	resp, err := c.Project.List(nil)
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("Failed to find the current environment")
	}

	return &resp.Data[0], nil
}

func GetOrCreateDefaultStack(c *client.RancherClient, name string) (*client.Environment, error) {
	required := false
	stackName := "Default"

	if name != "" {
		required = true
		stackName = name
	}

	resp, err := c.Environment.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"name": stackName,
		},
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Data) > 0 {
		return &resp.Data[0], nil
	}

	if required {
		return nil, fmt.Errorf("Failed to find stack: %s", name)
	}

	return c.Environment.Create(&client.Environment{
		Name: "Default",
	})
}

func getHostByHostname(c *client.RancherClient, name string) (client.ResourceCollection, error) {
	var result client.ResourceCollection
	allHosts, err := c.Host.List(nil)
	if err != nil {
		return result, err
	}

	for _, host := range allHosts.Data {
		if host.Hostname == name {
			result.Data = append(result.Data, host.Resource)
		}
	}

	return result, nil
}

func Lookup(c *client.RancherClient, name string, types ...string) (*client.Resource, error) {
	var byName *client.Resource

	for _, schemaType := range types {
		var resource client.Resource
		if err := c.ById(schemaType, name, &resource); !client.IsNotFound(err) && err != nil {
			return nil, err
		} else if err == nil {
			return &resource, nil
		}

		var collection client.ResourceCollection
		if err := c.List(schemaType, &client.ListOpts{
			Filters: map[string]interface{}{
				"name": name,
			},
		}, &collection); err != nil {
			return nil, err
		}

		if len(collection.Data) > 1 {
			ids := []string{}
			for _, data := range collection.Data {
				ids = append(ids, data.Id)
			}
			return nil, fmt.Errorf("Multiple reosurces of type %s found for name %s: %v", schemaType, name, ids)
		}

		if len(collection.Data) == 0 {
			var err error
			// Per type specific logic
			switch schemaType {
			case "host":
				collection, err = getHostByHostname(c, name)
			}
			if err != nil {
				return nil, err
			}
		}

		if len(collection.Data) == 0 {
			continue
		}

		if byName != nil {
			return nil, fmt.Errorf("Multiple resources named %s: %s:%s, %s:%s", name, collection.Data[0].Type,
				collection.Data[0].Id, byName.Type, byName.Id)
		}

		byName = &collection.Data[0]
	}

	if byName == nil {
		return nil, fmt.Errorf("Not found: %s", name)
	}

	return byName, nil
}

func appendTabDelim(buf *bytes.Buffer, value string) {
	if buf.Len() == 0 {
		buf.WriteString(value)
	} else {
		buf.WriteString("\t")
		buf.WriteString(value)
	}
}

func SimpleFormat(values [][]string) (string, string) {
	headerBuffer := bytes.Buffer{}
	valueBuffer := bytes.Buffer{}
	for _, v := range values {
		appendTabDelim(&headerBuffer, v[0])
		if strings.Contains(v[1], "{{") {
			appendTabDelim(&valueBuffer, v[1])
		} else {
			appendTabDelim(&valueBuffer, "{{."+v[1]+"}}")
		}
	}

	headerBuffer.WriteString("\n")
	valueBuffer.WriteString("\n")

	return headerBuffer.String(), valueBuffer.String()
}

func errorWrapper(f func(*cli.Context) error) func(*cli.Context) {
	return func(ctx *cli.Context) {
		if err := f(ctx); err != nil {
			logrus.Fatal(err)
		}
	}
}

func printTemplate(out io.Writer, templateContent string, obj interface{}) error {
	funcMap := map[string]interface{}{
		"endpoint": FormatEndpoint,
		"ips":      FormatIPAddresses,
	}
	tmpl, err := template.New("").Funcs(funcMap).Parse(templateContent)
	if err != nil {
		return err
	}

	return tmpl.Execute(out, obj)
}

func processExitCode(err error) error {
	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			os.Exit(status.ExitStatus())
		}
	}

	return err
}

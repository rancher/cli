package cmd

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/fatih/color"
	"github.com/rancher/go-rancher/v3"

	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var (
	errNoEnv         = errors.New("Failed to find the current environment")
	errNoURL         = errors.New("RANCHER_URL environment or --url is not set, run `config`")
	namespaceLabel   = "io.kubernetes.pod.namespace"
	podNameLabel     = "io.kubernetes.pod.name"
	podContainerName = "io.kubernetes.container.name"
)

func GetRawClient(ctx *cli.Context) (*client.RancherClient, error) {
	config, err := lookupConfig(ctx)
	if err != nil {
		return nil, err
	}
	url, err := baseURL(config.URL)
	if err != nil {
		return nil, err
	}
	return client.NewRancherClient(&client.ClientOpts{
		Url:       url + "/v3",
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
	})
}

func lookupConfig(ctx *cli.Context) (Config, error) {
	path := ctx.GlobalString("config")
	if path == "" {
		path = os.ExpandEnv("${HOME}/.rancher/cli.json")
	}

	config, err := LoadConfig(path)
	if err != nil {
		return config, err
	}

	url := ctx.GlobalString("url")
	accessKey := ctx.GlobalString("access-key")
	secretKey := ctx.GlobalString("secret-key")
	envName := ctx.GlobalString("environment")

	if url != "" {
		config.URL = url
	}
	if accessKey != "" {
		config.AccessKey = accessKey
	}
	if secretKey != "" {
		config.SecretKey = secretKey
	}
	if envName != "" {
		config.Environment = envName
	}

	if config.URL == "" {
		return config, errNoURL
	}

	return config, nil
}

func GetClient(ctx *cli.Context) (*client.RancherClient, error) {
	config, err := lookupConfig(ctx)
	if err != nil {
		return nil, err
	}

	url, err := config.EnvironmentURL()
	if err != nil {
		return nil, err
	}

	return client.NewRancherClient(&client.ClientOpts{
		Url:       url + "/schemas",
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
	})
}

func GetEnvironment(def string, c *client.RancherClient) (*client.Project, error) {
	resp, err := c.Project.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"all": true,
		},
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, errNoEnv
	}

	if len(resp.Data) == 1 {
		return &resp.Data[0], nil
	}

	if def == "" {
		names := []string{}
		for _, p := range resp.Data {
			cluster, err := c.Cluster.ById(p.ClusterId)
			if err != nil {
				return nil, err
			}
			names = append(names, fmt.Sprintf("%s(%s), cluster Name: %s (%s)", p.Name, p.Id, cluster.Name, cluster.Id))
		}

		idx := selectFromList("Environments:", names)
		return &resp.Data[idx], nil
	}

	return LookupEnvironment(c, def)
}

func LookupEnvironment(c *client.RancherClient, name string) (*client.Project, error) {
	env, err := Lookup(c, name, "account")
	if err != nil {
		return nil, err
	}
	if env.Type != "project" {
		return nil, fmt.Errorf("Failed to find environment: %s", name)
	}
	return c.Project.ById(env.Id)
}

func GetOrCreateDefaultStack(c *client.RancherClient, name string) (*client.Stack, error) {
	if name == "" {
		name = "Default"
	}

	resp, err := c.Stack.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"name":         name,
			"removed_null": 1,
		},
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Data) > 0 {
		return &resp.Data[0], nil
	}

	return c.Stack.Create(&client.Stack{
		Name: name,
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

func RandomName() string {
	return strings.Replace(namesgenerator.GetRandomName(0), "_", "-", -1)
}

func getContainerByName(c *client.RancherClient, name string) (client.ResourceCollection, error) {
	var result client.ResourceCollection
	stack, containerName, err := ParseName(c, name)
	containers, err := c.Container.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"stackId": stack.Id,
			"name":    containerName,
		},
	})
	if err != nil {
		return result, err
	}
	for _, container := range containers.Data {
		result.Data = append(result.Data, container.Resource)
	}
	return result, nil
}

func getProjectByname(c *client.RancherClient, name string) (client.ResourceCollection, error) {
	var result client.ResourceCollection
	clusterName, projectName := parseClusterAndProject(name)
	if clusterName != "" {
		clusters, err := c.Cluster.List(&client.ListOpts{
			Filters: map[string]interface{}{
				"name":         clusterName,
				"removed_null": "true",
			},
		})
		if err != nil {
			return result, err
		}
		if len(clusters.Data) == 0 {
			return result, errors.Errorf("failed to find cluster with name %s", clusterName)
		}
		projects, err := c.Project.List(&client.ListOpts{
			Filters: map[string]interface{}{
				"clusterId":    clusters.Data[0].Id,
				"name":         projectName,
				"removed_null": "true",
			},
		})
		if err != nil {
			return result, err
		}
		for _, project := range projects.Data {
			result.Data = append(result.Data, project.Resource)
		}
		return result, nil
	}
	projects, err := c.Project.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"name":         projectName,
			"all":          "true",
			"removed_null": "true",
		},
	})
	if err != nil {
		return result, err
	}
	for _, project := range projects.Data {
		result.Data = append(result.Data, project.Resource)
	}
	return result, nil
}

func getServiceByName(c *client.RancherClient, name string) (client.ResourceCollection, error) {
	var result client.ResourceCollection
	stack, serviceName, err := ParseName(c, name)

	services, err := c.Service.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"stackId": stack.Id,
			"name":    serviceName,
		},
	})
	if err != nil {
		return result, err
	}

	for _, service := range services.Data {
		result.Data = append(result.Data, service.Resource)
	}

	return result, nil
}

func Lookup(c *client.RancherClient, name string, types ...string) (*client.Resource, error) {
	var byName *client.Resource

	for _, schemaType := range types {
		var resource client.Resource
		// this is a hack for projects, it returns 403
		if !strings.Contains(name, "/") {
			if err := c.ById(schemaType, name, &resource); !client.IsNotFound(err) && err != nil {
				return nil, err
			} else if err == nil && resource.Id == name { // The ID check is because of an oddity in the id obfuscation
				return &resource, nil
			}
		}

		var collection client.ResourceCollection
		if err := c.List(schemaType, &client.ListOpts{
			Filters: map[string]interface{}{
				"name":         name,
				"all":          "true",
				"removed_null": "1",
			},
		}, &collection); err != nil {
			return nil, err
		}

		if len(collection.Data) > 1 {
			ids := []string{}
			for _, data := range collection.Data {
				switch schemaType {
				case "project":
					project, err := c.Project.ById(data.Id)
					if err != nil {
						return nil, err
					}
					cluster, err := c.Cluster.ById(project.ClusterId)
					if err != nil {
						return nil, err
					}
					ids = append(ids, fmt.Sprintf("cluster %s, %s (%s)", cluster.Name, data.Id, name))
				case "container":
					container, err := c.Container.ById(data.Id)
					if err != nil {
						return nil, err
					}
					host, err := c.Host.ById(container.HostId)
					if err != nil {
						return nil, err
					}
					ids = append(ids, fmt.Sprintf("host %s, %s (%s)", host.Hostname, data.Id, name))
				default:
					ids = append(ids, fmt.Sprintf("%s (%s)", data.Id, name))
				}

			}
			index := selectFromList("Resources: ", ids)
			return &collection.Data[index], nil
		}

		if len(collection.Data) == 0 {
			var err error
			// Per type specific logic
			switch schemaType {
			case "host":
				collection, err = getHostByHostname(c, name)
			case "service":
				collection, err = getServiceByName(c, name)
			case "container":
				collection, err = getContainerByName(c, name)
			case "project":
				collection, err = getProjectByname(c, name)
			}
			if err != nil {
				return nil, err
			}
		}

		if len(collection.Data) == 0 {
			continue
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

func defaultAction(fn func(ctx *cli.Context) error) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		if ctx.Bool("help") {
			cli.ShowAppHelp(ctx)
			return nil
		}
		return fn(ctx)
	}
}

func printTemplate(out io.Writer, templateContent string, obj interface{}) error {
	funcMap := map[string]interface{}{
		"endpoint": FormatEndpoint,
		"ips":      FormatIPAddresses,
		"json":     FormatJSON,
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

func getRandomColor() color.Attribute {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	index := r1.Intn(8)
	return colors[index]
}

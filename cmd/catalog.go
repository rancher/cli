package cmd

import (
	"fmt"
	"strings"

	"github.com/rancher/go-rancher/catalog"
	"github.com/urfave/cli"
)

func CatalogCommand() cli.Command {
	return cli.Command{
		Name:   "catalog",
		Usage:  "Operations with catalogs",
		Action: catalogLs,
		Flags:  []cli.Flag{},
	}
}

type CatalogData struct {
	ID       string
	Template catalog.Template
}

func catalogLs(ctx *cli.Context) error {
	config, err := lookupConfig(ctx)
	if err != nil {
		return err
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	proj, err := GetEnvironment(config.Environment, c)
	if err != nil {
		return err
	}

	cc, err := GetCatalogClient(ctx)
	if err != nil {
		return err
	}

	envData := NewEnvData(*proj)
	envFilter := ""
	switch envData.Orchestration {
	case "Kubernetes":
		envFilter = "kubernetes"
	case "Swarm":
		envFilter = "swarm"
	case "Mesos":
		envFilter = "mesos"
	}

	collection, err := cc.Template.List(nil)
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"NAME", "Template.Name"},
		{"CATEGORY", "Template.Category"},
		{"ID", "ID"},
	}, ctx)
	defer writer.Close()

	for _, item := range collection.Data {
		if item.TemplateBase != envFilter {
			continue
		}
		if item.Category == "System" {
			continue
		}
		writer.Write(CatalogData{
			ID:       templateID(item),
			Template: item,
		})
	}

	return writer.Err()
}

func templateID(template catalog.Template) string {
	parts := strings.SplitN(template.Path, "/", 2)
	if len(parts) != 2 {
		return template.Name
	}

	first := parts[0]
	second := parts[1]
	version := template.DefaultVersion

	parts = strings.SplitN(parts[1], "*", 2)
	if len(parts) == 2 {
		second = parts[1]
	}

	if version == "" {
		return fmt.Sprintf("%s/%s", first, second)
	}
	return fmt.Sprintf("%s/%s:%s", first, second, version)
}

func GetCatalogClient(ctx *cli.Context) (*catalog.RancherClient, error) {
	config, err := lookupConfig(ctx)
	if err != nil {
		return nil, err
	}

	idx := strings.LastIndex(config.URL, "/v1")
	if idx == -1 {
		return nil, fmt.Errorf("Invalid URL %s, must contain /v1", config.URL)
	}

	url := config.URL[:idx] + "/v1-catalog/schemas"
	return catalog.NewRancherClient(&catalog.ClientOpts{
		Url: url,
	})
}

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/rancher/go-rancher/catalog"
	"github.com/rancher/go-rancher/v2"
	"github.com/urfave/cli"
)

const (
	orchestrationSupported = "io.rancher.orchestration.supported"
)

func CatalogCommand() cli.Command {
	catalogLsFlags := []cli.Flag{
		cli.BoolFlag{
			Name:  "quiet,q",
			Usage: "Only display IDs",
		},
		cli.StringFlag{
			Name:  "format",
			Usage: "'json' or Custom format: {{.Id}} {{.Name}}",
		},
		cli.BoolFlag{
			Name:  "system,s",
			Usage: "Show system templates, not user",
		},
	}

	return cli.Command{
		Name:   "catalog",
		Usage:  "Operations with catalogs",
		Action: defaultAction(catalogLs),
		Flags:  catalogLsFlags,
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "ls",
				Usage:       "List catalog templates",
				Description: "\nList all catalog templates in the current $RANCHER_ENVIRONMENT. Use `--env <envID>` or `--env <envName>` to select a different environment.\n\nExample:\n\t$ rancher --env k8slab catalog ls\n",
				ArgsUsage:   "None",
				Action:      catalogLs,
				Flags:       catalogLsFlags,
			},
			cli.Command{
				Name:      "install",
				Usage:     "Install catalog template",
				Action:    catalogInstall,
				ArgsUsage: "[ID or NAME]...",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "answers,a",
						Usage: "Answer file",
					},
					cli.StringFlag{
						Name:  "name",
						Usage: "Name of stack to create",
					},
					cli.BoolFlag{
						Name:  "system,s",
						Usage: "Install a system template",
					},
				},
			},
			/*
				cli.Command{
					Name:   "upgrade",
					Usage:  "Upgrade catalog template",
					Action: errorWrapper(envUpdate),
					ArgsUsage: "[ID or NAME]"
					Flags:  []cli.Flag{},
				},
			*/
		},
	}
}

type CatalogData struct {
	ID       string
	Template catalog.Template
}

func getEnvFilter(proj *client.Project, ctx *cli.Context) string {
	envFilter := proj.Orchestration
	if envFilter == "cattle" {
		envFilter = ""
	}
	if ctx.Bool("system") {
		envFilter = "system"
	}
	return envFilter
}

func isInCSV(value, csv string) bool {
	for _, part := range strings.Split(csv, ",") {
		if value == part {
			return true
		}
	}
	return false
}

func isOrchestrationSupported(ctx *cli.Context, proj *client.Project, labels map[string]interface{}) bool {
	// Only check for system templates
	if !ctx.Bool("system") {
		return true
	}

	if supported, ok := labels[orchestrationSupported]; ok {
		supportedString := fmt.Sprint(supported)
		if supportedString != "" && !isInCSV(proj.Orchestration, supportedString) {
			return false
		}
	}

	return true
}

func isSupported(ctx *cli.Context, proj *client.Project, item catalog.Template) bool {
	envFilter := getEnvFilter(proj, ctx)
	if item.TemplateBase != envFilter {
		return false
	}
	return isOrchestrationSupported(ctx, proj, item.Labels)
}

func catalogLs(ctx *cli.Context) error {
	writer := NewTableWriter([][]string{
		{"NAME", "Template.Name"},
		{"CATEGORY", "Template.Category"},
		{"ID", "ID"},
	}, ctx)
	defer writer.Close()

	err := forEachTemplate(ctx, func(item catalog.Template) error {
		writer.Write(CatalogData{
			ID:       templateID(item),
			Template: item,
		})
		return nil
	})
	if err != nil {
		return err
	}
	return writer.Err()
}

func forEachTemplate(ctx *cli.Context, f func(item catalog.Template) error) error {
	_, c, proj, cc, err := setupCatalogContext(ctx)
	if err != nil {
		return err
	}

	opts, err := getListTemplatesOpts(ctx, c)
	if err != nil {
		return err
	}

	collection, err := cc.Template.List(opts)
	if err != nil {
		return err
	}

	for _, item := range collection.Data {
		if !isSupported(ctx, proj, item) {
			continue
		}
		if err := f(item); err != nil {
			return err
		}
	}

	return err
}

func getListTemplatesOpts(ctx *cli.Context, c *client.RancherClient) (*catalog.ListOpts, error) {
	opts := &catalog.ListOpts{
		Filters: map[string]interface{}{},
	}
	setting, err := c.Setting.ById("rancher.server.version")
	if err != nil {
		return nil, err
	}

	if setting != nil && setting.Value != "" {
		opts.Filters["minimumRancherVersion_lte"] = setting.Value
	}

	opts.Filters["category_ne"] = "system"
	return opts, nil
}

func setupCatalogContext(ctx *cli.Context) (Config, *client.RancherClient, *client.Project, *catalog.RancherClient, error) {
	config, err := lookupConfig(ctx)
	if err != nil {
		return config, nil, nil, nil, err
	}

	c, err := GetClient(ctx)
	if err != nil {
		return config, nil, nil, nil, err
	}

	proj, err := GetEnvironment(config.Environment, c)
	if err != nil {
		return config, nil, nil, nil, err
	}

	cc, err := GetCatalogClient(ctx)
	if err != nil {
		return config, nil, nil, nil, err
	}

	return config, c, proj, cc, nil
}

func templateNameAndVersion(name string) (string, string) {
	parts := strings.Split(name, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}

func catalogInstall(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return errors.New("Exactly one arguement is required")
	}

	_, c, proj, cc, err := setupCatalogContext(ctx)
	if err != nil {
		return err
	}

	templateReference := ctx.Args()[0]
	name, version := templateNameAndVersion(templateReference)

	template, err := getTemplate(ctx, name)
	if err != nil {
		return err
	}

	templateVersion, err := getTemplateVersion(ctx, cc, template, name, version)
	if err != nil {
		return err
	}

	answers, err := parseAnswers(ctx)
	if err != nil {
		return err
	}

	answers, err = askQuestions(answers, templateVersion)
	if err != nil {
		return err
	}

	stackName := ctx.String("name")
	if stackName == "" {
		stackName = strings.Title(strings.Split(name, "/")[1])
	}

	externalID := fmt.Sprintf("catalog://%s", templateVersion.Id)
	id := ""
	switch proj.Orchestration {
	case "cattle":
		stack, err := c.Stack.Create(&client.Stack{
			Name:           stackName,
			DockerCompose:  toString(templateVersion.Files["docker-compose.yml"]),
			RancherCompose: toString(templateVersion.Files["rancher-compose.yml"]),
			Environment:    answers,
			ExternalId:     externalID,
			System:         ctx.Bool("system"),
			StartOnCreate:  true,
		})
		if err != nil {
			return err
		}
		id = stack.Id
	case "kubernetes":
		stack, err := c.KubernetesStack.Create(&client.KubernetesStack{
			Name:        stackName,
			Templates:   templateVersion.Files,
			ExternalId:  externalID,
			Environment: answers,
			System:      ctx.Bool("system"),
		})
		if err != nil {
			return err
		}
		id = stack.Id
	}

	return WaitFor(ctx, id)
}

func toString(s interface{}) string {
	if s == nil {
		return ""
	}
	return fmt.Sprint(s)
}

func getTemplateVersion(ctx *cli.Context, cc *catalog.RancherClient, template catalog.Template, name, version string) (catalog.TemplateVersion, error) {
	templateVersion := catalog.TemplateVersion{}
	config, err := lookupConfig(ctx)
	if err != nil {
		return templateVersion, err
	}	
	
	if version == "" {
		version = template.DefaultVersion
	}

	link, ok := template.VersionLinks[version]
	if !ok {
		fmt.Printf("%#v\n", template)
		return templateVersion, fmt.Errorf("Failed to find the version %s for template %s", version, name)
	}

	client := &http.Client{}
    req, err := http.NewRequest("GET", fmt.Sprint(link), nil)
    req.SetBasicAuth(config.AccessKey, config.SecretKey)
    resp, err := client.Do(req)
	if err != nil {
		return templateVersion, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return templateVersion, fmt.Errorf("Bad response %d looking up %s", resp.StatusCode, link)

	}

	err = json.NewDecoder(resp.Body).Decode(&templateVersion)
	return templateVersion, err
}

func getTemplate(ctx *cli.Context, name string) (catalog.Template, error) {
	found := false
	foundTemplate := catalog.Template{}
	err := forEachTemplate(ctx, func(item catalog.Template) error {
		if found {
			return nil
		}

		templateName, _ := templateNameAndVersion(templateID(item))
		if templateName == name {
			found = true
			foundTemplate = item
		}

		return nil
	})
	if !found && err == nil {
		err = fmt.Errorf("Failed to find template %s", name)
	}
	return foundTemplate, err
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

	idx := strings.LastIndex(config.URL, "/v2-beta")
	if idx == -1 {
		return nil, fmt.Errorf("Invalid URL %s, must contain /v2-beta", config.URL)
	}

	url := config.URL[:idx] + "/v1-catalog/schemas"
	return catalog.NewRancherClient(&catalog.ClientOpts{
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
		Url:       url,
	})
}

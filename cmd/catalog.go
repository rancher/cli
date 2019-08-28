package cmd

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/rancher/go-rancher/catalog"
	"github.com/rancher/go-rancher/v2"
	"github.com/rancher/rancher-compose-executor/utils"
	"github.com/urfave/cli"
)

const (
	orchestrationSupported = "io.rancher.orchestration.supported"
	catalogProto           = "catalog://"
)

func CatalogCommand() cli.Command {
	catalogLsFlags := []cli.Flag{
		cli.BoolFlag{
			Name:  "quiet,q",
			Usage: "Only display IDs",
		},
		cli.StringFlag{
			Name:  "format",
			Usage: "'json' or Custom format: '{{.ID}} {{.Template.Id}}'",
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
				Name:        "show",
				Usage:       "Show catalog template versions",
				Description: "\nShow all catalog template versions in the current $RANCHER_ENVIRONMENT. Use `--env <envID>` or `--env <envName>` to select a different environment.\n\nExample:\n\t$ rancher --env k8slab catalog show <TEMPLATE_ID>\n",
				ArgsUsage:   "[TEMPLATE_ID]...",
				Action:      catalogShow,
				Flags:       catalogLsFlags,
			},
			cli.Command{
				Name:        "install",
				Usage:       "Install catalog template",
				Description: "\nInstall a catalog template in the current $RANCHER_ENVIRONMENT. \nUse `--env <envID>` or `--env <envName>` to select a different environment.\n\nExample:\n\t$ rancher --env k8slab catalog install <TEMPLATE_VERSION_ID>\n",
				Action:      catalogInstall,
				ArgsUsage:   "[TEMPLATE_VERSION_ID]...",
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
			cli.Command{
				Name:        "upgrade",
				Usage:       "Upgrade stack with new catalog template version",
				Description: "\nUpgrade stack with new catalog template version in the current $RANCHER_ENVIRONMENT. \nUse `--env <envID>` or `--env <envName>` to select a different environment.\n\nExample:\n\t$ rancher --env k8slab catalog upgrade <TEMPLATE_VERSION_ID> --stack id\n",
				Action:      catalogUpgrade,
				ArgsUsage:   "[TEMPLATE_VERSION_ID]...",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "answers,a",
						Usage: "Answer file",
					},
					cli.StringFlag{
						Name:  "stack,s",
						Usage: "Stack id to upgrade",
					},
					cli.BoolFlag{
						Name:  "confirm",
						Usage: "Wait for upgrade and confirm it",
					},
				},
			},
		},
	}
}

func catalogLs(ctx *cli.Context) error {
	operator, err := NewCatalogOperator(ctx)
	if err != nil {
		return err
	}

	return operator.Ls()
}

func catalogShow(ctx *cli.Context) error {
	operator, err := NewCatalogOperator(ctx)
	if err != nil {
		return err
	}

	return operator.Show()
}

func catalogInstall(ctx *cli.Context) error {
	operator, err := NewCatalogOperator(ctx)
	if err != nil {
		return err
	}

	return operator.Install()
}

func catalogUpgrade(ctx *cli.Context) error {
	operator, err := NewCatalogOperator(ctx)
	if err != nil {
		return err
	}

	return operator.Upgrade()
}

func normalizeTemplateID(id string) (string, error) {
	result := strings.Replace(id, catalogProto, "", 1)

	if len(strings.Split(result, ":")) != 2 {
		return "", fmt.Errorf("Bad template id format [%s]", result)
	}

	return result, nil
}

func normalizeTemplateVersionID(id string) (string, error) {
	result := strings.Replace(id, catalogProto, "", 1)

	if len(strings.Split(result, ":")) != 3 {
		return "", fmt.Errorf("Bad templateVersion id format [%s]", result)
	}

	return result, nil
}

type CatalogOperator struct {
	cclient  *catalog.RancherClient
	client   *client.RancherClient
	project  *client.Project
	config   Config
	context  *cli.Context
	listOpts *catalog.ListOpts
}

type CatalogData struct {
	ID       string
	Template *catalog.Template
	Category string
}

type CatalogVersionData struct {
	ID              string
	TemplateVersion *catalog.TemplateVersion
}

func (c *CatalogOperator) getEnvFilter() string {
	envFilter := c.project.Orchestration
	if envFilter == "cattle" {
		envFilter = ""
	}
	if c.context.Bool("system") {
		envFilter = "infra"
	}
	return envFilter
}

func (c *CatalogOperator) isOrchestrationSupported(labels map[string]string) bool {
	// Only check for system templates
	if !c.context.Bool("system") {
		return true
	}

	if supported, ok := labels[orchestrationSupported]; ok {
		supportedString := fmt.Sprint(supported)
		if supportedString != "" && !isInCSV(c.project.Orchestration, supportedString) {
			return false
		}
	}

	return true
}

func (c *CatalogOperator) isSupported(item *catalog.Template) bool {
	envFilter := c.getEnvFilter()
	if item.TemplateBase != envFilter {
		return false
	}
	return c.isOrchestrationSupported(item.Labels)
}

func (c *CatalogOperator) Ls() error {
	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "Template.Name"},
		{"CATEGORY", "Category"},
		{"VERSION", "Template.DefaultVersion"},
		{"VERSION ID", "Template.DefaultTemplateVersionId"},
	}, c.context)
	defer writer.Close()

	err := c.forEachTemplate(func(item *catalog.Template) error {
		writer.Write(CatalogData{
			ID:       item.Id,
			Template: item,
			Category: strings.Join(item.Categories, ","),
		})
		return nil
	})
	if err != nil {
		return err
	}
	return writer.Err()
}

func (c *CatalogOperator) Show() error {
	if len(c.context.Args()) != 1 {
		return errors.New("Exactly one argument is required")
	}

	templateID := c.context.Args()[0]

	template, err := c.GetTemplateByID(templateID)
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"VERSION", "TemplateVersion.Version"},
	}, c.context)
	defer writer.Close()

	err = c.forEachTemplateVersion(template, func(item *catalog.TemplateVersion) error {
		writer.Write(CatalogVersionData{
			ID:              item.Id,
			TemplateVersion: item,
		})
		return nil
	})
	if err != nil {
		return err
	}

	return writer.Err()
}

func (c *CatalogOperator) Install() error {
	if len(c.context.Args()) != 1 {
		return errors.New("Exactly one argument is required")
	}

	templateVersionID := c.context.Args()[0]
	templateVersion, err := c.GetTemplateVersionByID(templateVersionID)
	if err != nil {
		return err
	}

	answers, err := c.AskQuestions(templateVersion)

	stackName := c.context.String("name")
	if stackName == "" {
		stackName = strings.Title(strings.Split(templateVersion.Id, ":")[1])
	}

	externalID := fmt.Sprintf("%s%s", catalogProto, templateVersion.Id)
	id := ""
	switch c.project.Orchestration {
	case "cattle":
		var composeFile string
		if templateVersion.Files["docker-compose.yml"] == "" {
			composeFile = toString(templateVersion.Files["docker-compose.yml.tpl"])
		} else {
			composeFile = toString(templateVersion.Files["docker-compose.yml"])
		}
		stack, err := c.client.Stack.Create(&client.Stack{
			Name:           stackName,
			DockerCompose:  composeFile,
			RancherCompose: toString(templateVersion.Files["rancher-compose.yml"]),
			Environment:    answers,
			ExternalId:     externalID,
			System:         c.context.Bool("system"),
			StartOnCreate:  true,
		})
		if err != nil {
			return err
		}
		id = stack.Id
	case "kubernetes":
		stack, err := c.client.KubernetesStack.Create(&client.KubernetesStack{
			Name:        stackName,
			Templates:   utils.ToMapInterface(templateVersion.Files),
			ExternalId:  externalID,
			Environment: answers,
			System:      c.context.Bool("system"),
		})
		if err != nil {
			return err
		}
		id = stack.Id
	}

	return WaitFor(c.context, id)
}

func (c *CatalogOperator) Upgrade() error {
	if len(c.context.Args()) != 1 {
		return errors.New("Exactly one arguments is required")
	}

	if len(c.context.String("stack")) == 0 {
		return errors.New("Stack id is required")
	}

	if c.context.Bool("confirm") {
		c.context.GlobalSet("wait", "true")
	}

	templateVersionID := c.context.Args()[0]
	templateVersion, err := c.GetTemplateVersionByID(templateVersionID)
	if err != nil {
		return err
	}

	var answers map[string]interface{}
	if c.context.String("answers") != "" {
		answers, err = c.AskQuestions(templateVersion)
	}

	stackID := c.context.String("stack")
	externalID := fmt.Sprintf("%s%s", catalogProto, templateVersion.Id)
	switch c.project.Orchestration {
	case "cattle":
		var composeFile string

		stack, err := c.GetStackByID(stackID)
		if err != nil {
			return err
		}
		if !c.CheckTemplateVersionUpgrade(stack.ExternalId, templateVersion.Id) {
			return fmt.Errorf("Stack %s is not upgradable by template version %s", stack.Id, templateVersion.Id)
		}

		if templateVersion.Files["docker-compose.yml"] == "" {
			composeFile = toString(templateVersion.Files["docker-compose.yml.tpl"])
		} else {
			composeFile = toString(templateVersion.Files["docker-compose.yml"])
		}
		stackup := &client.StackUpgrade{
			DockerCompose:  composeFile,
			RancherCompose: toString(templateVersion.Files["rancher-compose.yml"]),
			Environment:    answers,
			ExternalId:     externalID,
		}

		stack, err = c.client.Stack.ActionUpgrade(stack, stackup)
		if err != nil {
			return err
		}
		if c.context.Bool("confirm") {
			fmt.Printf("Upgrading stack %s to template version %s", stack.Id, templateVersion.Id)
			c.context.GlobalSet("wait-state", "upgraded")
			WaitFor(c.context, stack.Id)
			c.context.GlobalSet("wait-state", "active")
			stack, _ := c.GetStackByID(stack.Id)
			stack, err = c.client.Stack.ActionFinishupgrade(stack)
			if err != nil {
				return err
			}
			fmt.Printf("Finishing upgrade of stack %s to template version %s", stack.Id, templateVersion.Id)
		}
	case "kubernetes":
		stack, err := c.GetKubernetesStackByID(stackID)
		if err != nil {
			return err
		}
		if !c.CheckTemplateVersionUpgrade(stack.ExternalId, templateVersion.Id) {
			return fmt.Errorf("Stack %s is not upgradable by template version %s", stack.Id, templateVersion.Id)
		}

		stackup := &client.KubernetesStackUpgrade{
			Templates:   utils.ToMapInterface(templateVersion.Files),
			ExternalId:  externalID,
			Environment: answers,
		}

		stack, err = c.client.KubernetesStack.ActionUpgrade(stack, stackup)
		if err != nil {
			return err
		}
	}

	return WaitFor(c.context, stackID)
}

func (c *CatalogOperator) forEachTemplateVersion(template *catalog.Template, f func(template *catalog.TemplateVersion) error) error {
	var err error
	var templateVersion *catalog.TemplateVersion

	links := make(map[int]string)
	indexes := make([]int, 0, len(template.VersionLinks))
	for _, link := range template.VersionLinks {
		splitedlink := strings.Split(link, ":")
		index, _ := strconv.Atoi(splitedlink[len(splitedlink)-1])
		links[index] = link
		indexes = append(indexes, index)
	}

	sort.Ints(indexes)
	for _, index := range indexes {
		templateVersion, err = c.GetTemplateVersionByLink(links[index])
		if err != nil {
			return err
		}
		if err := f(templateVersion); err != nil {
			return err
		}
	}

	return err
}

func (c *CatalogOperator) forEachTemplate(f func(item *catalog.Template) error) error {
	collection, err := c.cclient.Template.List(c.listOpts)
	if err != nil {
		return err
	}

	collectiondata := collection.Data

	for {
		collection, _ = collection.Next()
		if collection == nil {
			break
		}
		collectiondata = append(collectiondata, collection.Data...)
		if !collection.Pagination.Partial {
			break
		}
	}

	for _, item := range collectiondata {
		if !c.isSupported(&item) {
			continue
		}
		if err := f(&item); err != nil {
			return err
		}
	}

	return err
}

func (c *CatalogOperator) AskQuestions(tver *catalog.TemplateVersion) (map[string]interface{}, error) {
	answers, err := parseAnswers(c.context)
	if err != nil {
		return nil, err
	}

	answers, err = askQuestions(answers, *tver)
	if err != nil {
		return nil, err
	}

	return answers, nil
}

func (c *CatalogOperator) GetTemplateByID(id string) (*catalog.Template, error) {
	normID, err := normalizeTemplateID(id)
	if err != nil {
		return nil, err
	}
	template, err := c.cclient.Template.ById(normID)
	if err != nil {
		return nil, err
	}
	if template == nil || len(template.VersionLinks) == 0 {
		return nil, fmt.Errorf("Template %s not found", id)
	}
	if !c.isSupported(template) {
		return nil, fmt.Errorf("Template %s not supported by env %s", id, c.project.Orchestration)
	}

	return template, nil
}

func (c *CatalogOperator) GetStackByID(id string) (*client.Stack, error) {
	stack, err := c.client.Stack.ById(id)
	if err != nil {
		return nil, err
	}
	if stack == nil {
		return nil, fmt.Errorf("Stack %s not found", id)
	}

	return stack, nil
}

func (c *CatalogOperator) GetKubernetesStackByID(id string) (*client.KubernetesStack, error) {
	stack, err := c.client.KubernetesStack.ById(id)
	if err != nil {
		return nil, err
	}
	if stack == nil {
		return nil, fmt.Errorf("KubernetesStack %s not found", id)
	}

	return stack, nil
}

func (c *CatalogOperator) GetTemplateVersion(version, url string) (*catalog.TemplateVersion, error) {
	templateVersion := &catalog.TemplateVersion{}

	resource := catalog.Resource{}
	resource.Links = make(map[string]string)
	resource.Links[version] = url

	err := c.cclient.RancherBaseClient.GetLink(resource, version, templateVersion)

	if templateVersion == nil || templateVersion.Type != "templateVersion" {
		return nil, fmt.Errorf("Template version not found [%s]", url)
	}

	return templateVersion, err
}

func (c *CatalogOperator) GetTemplateVersionByTemplate(template *catalog.Template, version string) (*catalog.TemplateVersion, error) {
	if version == "" {
		version = template.DefaultVersion
	}
	url := template.VersionLinks[version]

	return c.GetTemplateVersion(version, url)
}

func (c *CatalogOperator) GetTemplateVersionByLink(url string) (*catalog.TemplateVersion, error) {
	version := "byLink"

	return c.GetTemplateVersion(version, url)
}

func (c *CatalogOperator) GetTemplateVersionByID(id string) (*catalog.TemplateVersion, error) {
	normID, err := normalizeTemplateVersionID(id)
	if err != nil {
		return nil, err
	}
	version := "byId"

	schemas := c.cclient.RancherBaseClient.GetSchemas()
	schema, ok := schemas.CheckSchema("template")
	if !ok {
		return nil, fmt.Errorf("Template version schema not found")
	}
	url := schema.Links["collection"] + "/" + normID

	return c.GetTemplateVersion(version, url)
}

func (c *CatalogOperator) GetTemplateVersionUpgradesByID(id string) []string {
	tver, _ := c.GetTemplateVersionByID(id)
	return c.GetTemplateVersionUpgrades(tver)
}

func (c *CatalogOperator) GetTemplateVersionUpgrades(tver *catalog.TemplateVersion) []string {
	if tver == nil || len(tver.UpgradeVersionLinks) == 0 {
		return nil
	}
	upgrades := make([]string, 0, len(tver.UpgradeVersionLinks))
	for _, upgrade := range tver.UpgradeVersionLinks {
		upgradeID := strings.Split(upgrade, "/")
		upgrades = append(upgrades, upgradeID[len(upgradeID)-1])
	}

	sort.Strings(upgrades)

	return upgrades
}

func (c *CatalogOperator) CheckTemplateVersionUpgrade(externalid, newid string) bool {
	upgrades := c.GetTemplateVersionUpgradesByID(externalid)
	for _, upgrade := range upgrades {
		if upgrade == newid {
			return true
		}
	}
	return false
}

func (c *CatalogOperator) getCatalogClient() error {
	var err error

	c.cclient, err = catalog.NewRancherClient(&catalog.ClientOpts{
		AccessKey: c.config.AccessKey,
		SecretKey: c.config.SecretKey,
		Url:       c.config.URL,
	})
	if err != nil {
		return fmt.Errorf("Failed to get rancher catalog client %s %s", c.config.URL, err)
	}

	return err
}

func (c *CatalogOperator) getConfig() error {
	var err error

	c.config, err = lookupConfig(c.context)
	if err != nil {
		return fmt.Errorf("Failed to get rancher config %s", err)
	}

	return err
}

func (c *CatalogOperator) getClient() error {
	var err error

	c.client, err = GetClient(c.context)
	if err != nil {
		return fmt.Errorf("Failed to get rancher client %s", err)
	}

	return err
}

func (c *CatalogOperator) getProject() error {
	var err error

	c.project, err = GetEnvironment(c.config.Environment, c.client)
	if err != nil {
		return fmt.Errorf("Failed to get rancher environment %s %s", c.config.Environment, err)
	}

	return err
}

func (c *CatalogOperator) getListOpts() error {
	opts := catalog.NewListOpts()

	setting, err := c.client.Setting.ById("rancher.server.version")
	if err != nil {
		return err
	}

	if setting != nil && setting.Value != "" {
		opts.Filters["rancherVersion"] = setting.Value
	}

	opts.Filters["category_ne"] = "infra"

	c.listOpts = opts
	return nil
}

func NewCatalogOperator(ctx *cli.Context) (*CatalogOperator, error) {
	c := &CatalogOperator{
		context: ctx,
	}

	err := c.getConfig()
	if err != nil {
		return nil, err
	}

	err = c.getCatalogClient()
	if err != nil {
		return nil, err
	}

	err = c.getClient()
	if err != nil {
		return nil, err
	}

	err = c.getProject()
	if err != nil {
		return nil, err
	}

	err = c.getListOpts()
	if err != nil {
		return nil, err
	}

	return c, nil
}

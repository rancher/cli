package cmd

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	gover "github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/rancher/cli/cliclient"
	"github.com/rancher/norman/clientbase"
	clusterClient "github.com/rancher/rancher/pkg/client/generated/cluster/v3"
	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	projectClient "github.com/rancher/rancher/pkg/client/generated/project/v3"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

const (
	installAppDescription = `
Install an app template in the current Rancher server. This defaults to the newest version of the app template.
Specify a version using '--version' if required.
The app will be installed into a new namespace unless '--namespace' is specified.

Example:
	# Install the redis template without any options
	$ rancher app install redis appFoo

	# Block cli until installation has finished or encountered an error. Use after app install.
	$ rancher wait <app-id>

	# Install the local redis template folder without any options
	$ rancher app install ./redis appFoo

	# Install the redis template and specify an answers file location
	$ rancher app install --answers /example/answers.yaml redis appFoo

	# Install the redis template and set multiple answers and the version to install
	$ rancher app install --set foo=bar --set-string baz=bunk --version 1.0.1 redis appFoo

	# Install the redis template and specify the namespace for the app
	$ rancher app install --namespace bar redis appFoo
`
	upgradeAppDescription = `
Upgrade an existing app to a newer version via app template or app version in the current Rancher server.

Example:
	# Upgrade the 'appFoo' app to latest version without any options
	$ rancher app upgrade appFoo latest

	# Upgrade the 'appFoo' app by local template folder without any options
	$ rancher app upgrade appFoo ./redis

	# Upgrade the 'appFoo' app and set multiple answers and the 0.2.0 version to install
	$ rancher app upgrade --set foo=bar --set-string baz=bunk appFoo 0.2.0
`
)

type AppData struct {
	ID       string
	App      projectClient.App
	Catalog  string
	Template string
	Version  string
}

type TemplateData struct {
	ID       string
	Template managementClient.Template
	Category string
}

type VersionData struct {
	Current string
	Version string
}

type revision struct {
	Current  string
	Name     string
	Created  time.Time
	Human    string
	Catalog  string
	Template string
	Version  string
}

type chartVersion struct {
	chartMetadata `yaml:",inline"`
	Dir           string   `json:"-" yaml:"-"`
	URLs          []string `json:"urls" yaml:"urls"`
	Digest        string   `json:"digest,omitempty" yaml:"digest,omitempty"`
}

type chartMetadata struct {
	Name        string   `json:"name,omitempty" yaml:"name,omitempty"`
	Sources     []string `json:"sources,omitempty" yaml:"sources,omitempty"`
	Version     string   `json:"version,omitempty" yaml:"version,omitempty"`
	KubeVersion string   `json:"kubeVersion,omitempty" yaml:"kubeVersion,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Keywords    []string `json:"keywords,omitempty" yaml:"keywords,omitempty"`
	Icon        string   `json:"icon,omitempty" yaml:"icon,omitempty"`
}

type revSlice []revision

func (s revSlice) Less(i, j int) bool { return s[i].Created.After(s[j].Created) }
func (s revSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s revSlice) Len() int           { return len(s) }

func AppCommand() cli.Command {
	appLsFlags := []cli.Flag{
		formatFlag,
		cli.BoolFlag{
			Name:  "quiet,q",
			Usage: "Only display IDs",
		},
	}

	return cli.Command{
		Name:    "apps",
		Aliases: []string{"app"},
		Usage:   "Operations with apps. Uses helm. Flags prepended with \"helm\" can also be accurately described by helm documentation.",
		Action:  defaultAction(appLs),
		Flags:   appLsFlags,
		Subcommands: []cli.Command{
			{
				Name:        "ls",
				Usage:       "List apps",
				Description: "\nList all apps in the current Rancher server",
				ArgsUsage:   "None",
				Action:      appLs,
				Flags:       appLsFlags,
			},
			{
				Name:      "delete",
				Usage:     "Delete an app",
				Action:    appDelete,
				ArgsUsage: "[APP_NAME/APP_ID]",
			},
			{
				Name:        "install",
				Usage:       "Install an app template",
				Description: installAppDescription,
				Action:      templateInstall,
				ArgsUsage:   "[TEMPLATE_NAME/TEMPLATE_PATH, APP_NAME]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "answers,a",
						Usage: "Path to an answers file, the format of the file is a map with key:value. This supports JSON and YAML.",
					},
					cli.StringFlag{
						Name:  "values",
						Usage: "Path to a helm values file.",
					},
					cli.StringFlag{
						Name:  "namespace,n",
						Usage: "Namespace to install the app into",
					},
					cli.StringSliceFlag{
						Name:  "set",
						Usage: "Set answers for the template, can be used multiple times. Example: --set foo=bar",
					},
					cli.StringSliceFlag{
						Name:  "set-string",
						Usage: "Set string answers for the template (Skips Helm's type conversion), can be used multiple times. Example: --set-string foo=bar",
					},
					cli.StringFlag{
						Name:  "version",
						Usage: "Version of the template to use",
					},
					cli.BoolFlag{
						Name:  "no-prompt",
						Usage: "Suppress asking questions and use the default values when required answers are not provided",
					},
					cli.IntFlag{
						Name:  "helm-timeout",
						Usage: "Amount of time for helm to wait for k8s commands (default is 300 secs). Example: --helm-timeout 600",
						Value: 300,
					},
					cli.BoolFlag{
						Name:  "helm-wait",
						Usage: "Helm will wait for as long as timeout value, for installed resources to be ready (pods, PVCs, deployments, etc.). Example: --helm-wait",
					},
				},
			},
			{
				Name:      "rollback",
				Usage:     "Rollback an app to a previous version",
				Action:    appRollback,
				ArgsUsage: "[APP_NAME/APP_ID, REVISION_ID/REVISION_NAME]",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "show-revisions,r",
						Usage: "Show revisions available to rollback to",
					},
					cli.BoolFlag{
						Name:  "force,f",
						Usage: "Force rollback, deletes and recreates resources if needed during rollback. (default is false)",
					},
				},
			},
			{
				Name:        "upgrade",
				Usage:       "Upgrade an existing app to a newer version",
				Description: upgradeAppDescription,
				Action:      appUpgrade,
				ArgsUsage:   "[APP_NAME/APP_ID VERSION/TEMPLATE_PATH]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "answers,a",
						Usage: "Path to an answers file, the format of the file is a map with key:value. Supports JSON and YAML",
					},
					cli.StringFlag{
						Name:  "values",
						Usage: "Path to a helm values file.",
					},
					cli.StringSliceFlag{
						Name:  "set",
						Usage: "Set answers for the template, can be used multiple times. Example: --set foo=bar",
					},
					cli.StringSliceFlag{
						Name:  "set-string",
						Usage: "Set string answers for the template (Skips Helm's type conversion), can be used multiple times. Example: --set-string foo=bar",
					},
					cli.BoolFlag{
						Name:  "show-versions,v",
						Usage: "Display versions available to upgrade to",
					},
					cli.BoolFlag{
						Name:  "reset",
						Usage: "Reset all catalog app answers",
					},
					cli.BoolFlag{
						Name:  "force,f",
						Usage: "Force upgrade, deletes and recreates resources if needed during upgrade. (default is false)",
					},
				},
			},
			{
				Name:        "list-templates",
				Aliases:     []string{"lt"},
				Usage:       "List templates available for installation",
				Description: "\nList all app templates in the current Rancher server",
				ArgsUsage:   "None",
				Action:      templateLs,
				Flags: []cli.Flag{
					formatFlag,
					cli.StringFlag{
						Name:  "catalog",
						Usage: "Specify the catalog to list templates for",
					},
				},
			},
			{
				Name:        "show-template",
				Aliases:     []string{"st"},
				Usage:       "Show versions available to install for an app template",
				Description: "\nShow all available versions of an app template",
				ArgsUsage:   "[TEMPLATE_ID]",
				Action:      templateShow,
			},
			{
				Name:      "show-app",
				Aliases:   []string{"sa"},
				Usage:     "Show an app's available versions and revisions",
				ArgsUsage: "[APP_NAME/APP_ID]",
				Action:    showApp,
				Flags: []cli.Flag{
					formatFlag,
				},
			},
			{
				Name:      "show-notes",
				Usage:     "Show contents of apps notes.txt",
				Action:    appNotes,
				ArgsUsage: "[APP_NAME/APP_ID]",
			},
		},
	}
}

func appLs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	collection, err := c.ProjectClient.App.List(defaultListOpts(ctx))
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "App.Name"},
		{"STATE", "App.State"},
		{"CATALOG", "Catalog"},
		{"TEMPLATE", "Template"},
		{"VERSION", "Version"},
	}, ctx)

	defer writer.Close()

	for _, item := range collection.Data {
		appExternalID := item.ExternalID
		appTemplateFiles := make(map[string]string)
		if appExternalID == "" {
			// add namespace prefix to AppRevisionID to create a Rancher API style ID
			appRevisionID := strings.Replace(item.ID, item.Name, item.AppRevisionID, -1)

			appRevision, err := c.ProjectClient.AppRevision.ByID(appRevisionID)
			if err != nil {
				return err
			}

			if appRevision.Status != nil {
				appTemplateFiles = appRevision.Status.Files
			}
		}

		parsedInfo, err := parseTemplateInfo(appExternalID, appTemplateFiles)
		if err != nil {
			return err
		}

		appData := &AppData{
			ID:       item.ID,
			App:      item,
			Catalog:  parsedInfo["catalog"],
			Template: parsedInfo["template"],
			Version:  parsedInfo["version"],
		}
		writer.Write(appData)
	}
	return writer.Err()

}

func parseTemplateInfo(appExternalID string, appTemplateFiles map[string]string) (map[string]string, error) {
	if appExternalID != "" {
		parsedExternal, parseErr := parseExternalID(appExternalID)
		if parseErr != nil {
			return nil, errors.Wrap(parseErr, "failed to parse ExternalID from app")
		}

		return parsedExternal, nil
	}

	for fileName, fileContent := range appTemplateFiles {
		if strings.HasSuffix(fileName, "/Chart.yaml") || strings.HasSuffix(fileName, "/Chart.yml") {
			content, decodeErr := base64.StdEncoding.DecodeString(fileContent)
			if decodeErr != nil {
				return nil, errors.Wrap(decodeErr, "failed to decode Chart.yaml from app")
			}

			version := &chartVersion{}
			unmarshalErr := yaml.Unmarshal(content, version)
			if unmarshalErr != nil {
				return nil, errors.Wrap(unmarshalErr, "failed to parse Chart.yaml from app")
			}

			return map[string]string{
				"catalog":  "local directory",
				"template": version.Name,
				"version":  version.Version,
			}, nil
		}
	}

	return nil, errors.New("can't parse info from app")
}

func appDelete(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	for _, arg := range ctx.Args() {
		resource, err := Lookup(c, arg, "app")
		if err != nil {
			return err
		}

		app, err := c.ProjectClient.App.ByID(resource.ID)
		if err != nil {
			return err
		}

		err = c.ProjectClient.App.Delete(app)
		if err != nil {
			return err
		}
	}

	return nil

}

func appUpgrade(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	if ctx.Bool("show-versions") {
		return outputVersions(ctx, c)
	}

	if ctx.NArg() < 2 {
		return cli.ShowSubcommandHelp(ctx)
	}

	appName := ctx.Args().First()
	appVersionOrLocalTemplatePath := ctx.Args().Get(1)

	resource, err := Lookup(c, appName, "app")
	if err != nil {
		return err
	}

	app, err := c.ProjectClient.App.ByID(resource.ID)
	if err != nil {
		return err
	}

	answers := app.Answers
	answersSetString := app.AnswersSetString
	values := app.ValuesYaml
	answers, answersSetString, err = processAnswerUpdates(ctx, answers, answersSetString)
	if err != nil {
		return err
	}
	values, err = processValueUpgrades(ctx, values)
	if err != nil {
		return err
	}

	force := ctx.Bool("force")

	au := &projectClient.AppUpgradeConfig{
		Answers:          answers,
		AnswersSetString: answersSetString,
		ForceUpgrade:     force,
		ValuesYaml:       values,
	}

	if resolveTemplatePath(appVersionOrLocalTemplatePath) {
		// if it is a path, upgrade install charts locally
		localTemplatePath := appVersionOrLocalTemplatePath
		_, files, err := walkTemplateDirectory(localTemplatePath)
		if err != nil {
			return err
		}

		au.Files = files
	} else {
		appVersion := appVersionOrLocalTemplatePath
		externalID, err := updateExternalIDVersion(app.ExternalID, appVersion)
		if err != nil {
			return err
		}

		filter := defaultListOpts(ctx)
		filter.Filters["externalId"] = externalID

		template, err := c.ManagementClient.TemplateVersion.List(filter)
		if err != nil {
			return err
		}
		if len(template.Data) == 0 {
			return fmt.Errorf("version %s is not valid", appVersion)
		}

		au.ExternalID = template.Data[0].ExternalID
	}

	return c.ProjectClient.App.ActionUpgrade(app, au)
}

func updateExternalIDVersion(externalID string, version string) (string, error) {
	u, err := url.Parse(externalID)
	if err != nil {
		return "", err
	}

	oldVersionQuery := fmt.Sprintf("version=%s", u.Query().Get("version"))
	newVersionQuery := fmt.Sprintf("version=%s", version)
	return strings.Replace(externalID, oldVersionQuery, newVersionQuery, 1), nil
}

func appRollback(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	if ctx.Bool("show-revisions") {
		return outputRevisions(ctx, c)
	}

	if ctx.NArg() < 2 {
		return cli.ShowSubcommandHelp(ctx)
	}

	force := ctx.Bool("force")

	resource, err := Lookup(c, ctx.Args().First(), "app")
	if err != nil {
		return err
	}

	app, err := c.ProjectClient.App.ByID(resource.ID)
	if err != nil {
		return err
	}

	revisionResource, err := Lookup(c, ctx.Args().Get(1), "appRevision")
	if err != nil {
		return err
	}

	revision, err := c.ProjectClient.AppRevision.ByID(revisionResource.ID)
	if err != nil {
		return err
	}

	rr := &projectClient.RollbackRevision{
		ForceUpgrade: force,
		RevisionID:   revision.Name,
	}

	return c.ProjectClient.App.ActionRollback(app, rr)
}

func templateLs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	filter := defaultListOpts(ctx)
	if ctx.String("app") != "" {
		resource, err := Lookup(c, ctx.String("app"), "app")
		if err != nil {
			return err
		}
		filter.Filters["appId"] = resource.ID
	}

	collection, err := c.ManagementClient.Template.List(filter)
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "Template.Name"},
		{"CATEGORY", "Category"},
	}, ctx)

	defer writer.Close()

	for _, item := range collection.Data {
		writer.Write(&TemplateData{
			ID:       item.ID,
			Template: item,
			Category: strings.Join(item.Categories, ","),
		})
	}

	return writer.Err()
}

func templateShow(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	resource, err := Lookup(c, ctx.Args().First(), "template")
	if err != nil {
		return err
	}

	template, err := getFilteredTemplate(ctx, c, resource.ID)
	if err != nil {
		return err
	}

	sortedVersions, err := sortTemplateVersions(template)
	if err != nil {
		return err
	}

	if len(sortedVersions) == 0 {
		fmt.Println("No app versions available to install for this version of Rancher server")
	}

	for _, version := range sortedVersions {
		fmt.Println(version)
	}

	return nil
}

func templateInstall(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}
	templateName := ctx.Args().First()
	appName := ctx.Args().Get(1)

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	app := &projectClient.App{
		Name: appName,
	}
	if resolveTemplatePath(templateName) {
		// if it is a path, install charts locally
		chartName, files, err := walkTemplateDirectory(templateName)
		if err != nil {
			return err
		}
		answers, answersSetString, err := processAnswerInstall(ctx, nil, nil, nil, false, false)
		if err != nil {
			return err
		}
		values, err := processValueInstall(ctx, nil, "")
		if err != nil {
			return err
		}

		app.Files = files
		app.Answers = answers
		app.AnswersSetString = answersSetString
		app.ValuesYaml = values
		namespace := ctx.String("namespace")
		if namespace == "" {
			namespace = chartName + "-" + RandomLetters(5)
		}
		err = createNamespace(c, namespace)
		if err != nil {
			return err
		}
		app.TargetNamespace = namespace
	} else {
		resource, err := Lookup(c, templateName, "template")
		if err != nil {
			return err
		}

		template, err := getFilteredTemplate(ctx, c, resource.ID)
		if err != nil {
			return err
		}

		latestVersion, err := getTemplateLatestVersion(template)
		if err != nil {
			return err
		}

		templateVersionID := templateVersionIDFromVersionLink(template.VersionLinks[latestVersion])
		userVersion := ctx.String("version")
		if userVersion != "" {
			if link, ok := template.VersionLinks[userVersion]; ok {
				templateVersionID = templateVersionIDFromVersionLink(link)
			} else {
				return fmt.Errorf(
					"version %s for template %s is invalid, run 'rancher app show-template %s' for a list of versions",
					userVersion,
					templateName,
					templateName,
				)
			}
		}

		templateVersion, err := c.ManagementClient.TemplateVersion.ByID(templateVersionID)
		if err != nil {
			return err
		}

		interactive := !ctx.Bool("no-prompt")
		answers, answersSetString, err := processAnswerInstall(ctx, templateVersion, nil, nil, interactive, false)
		if err != nil {
			return err
		}
		values, err := processValueInstall(ctx, templateVersion, "")
		if err != nil {
			return err
		}
		namespace := ctx.String("namespace")
		if namespace == "" {
			namespace = template.Name + "-" + RandomLetters(5)
		}
		err = createNamespace(c, namespace)
		if err != nil {
			return err
		}
		app.Answers = answers
		app.AnswersSetString = answersSetString
		app.ValuesYaml = values
		app.ExternalID = templateVersion.ExternalID
		app.TargetNamespace = namespace
	}

	app.Wait = ctx.Bool("helm-wait")
	app.Timeout = ctx.Int64("helm-timeout")

	madeApp, err := c.ProjectClient.App.Create(app)
	if err != nil {
		return err
	}

	fmt.Printf("run \"app show-notes %s\" to view app notes once app is ready\n", madeApp.Name)

	return nil
}

// appNotes prints notes from app's notes.txt file
func appNotes(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	if ctx.NArg() < 1 {
		return cli.ShowSubcommandHelp(ctx)
	}

	resource, err := Lookup(c, ctx.Args().First(), "app")
	if err != nil {
		return err
	}

	app, err := c.ProjectClient.App.ByID(resource.ID)
	if err != nil {
		return err
	}

	if len(app.Notes) > 0 {
		fmt.Println(app.Notes)
	} else {
		fmt.Println("no notes to print")
	}

	return nil
}

func resolveTemplatePath(templateName string) bool {
	return templateName == "." || strings.Contains(templateName, "\\\\") || strings.Contains(templateName, "/")
}

func walkTemplateDirectory(templatePath string) (string, map[string]string, error) {
	templateAbsPath, parsedErr := filepath.Abs(templatePath)
	if parsedErr != nil {
		return "", nil, parsedErr
	}
	if _, statErr := os.Stat(templateAbsPath); statErr != nil {
		return "", nil, statErr
	}

	var (
		chartName string
		files     = make(map[string]string)
		err       error
	)
	err = filepath.Walk(templateAbsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if !strings.EqualFold(info.Name(), "Chart.yaml") {
			return nil
		}
		version := &chartVersion{}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		rootDir := filepath.Dir(path)
		if err := yaml.Unmarshal(content, version); err != nil {
			return err
		}
		chartName = version.Name
		err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if len(content) > 0 {
				key := filepath.Join(chartName, strings.TrimPrefix(path, rootDir+"/"))
				files[key] = base64.StdEncoding.EncodeToString(content)
			}
			return nil
		})
		if err != nil {
			return err
		}

		return filepath.SkipDir
	})

	return chartName, files, err
}

func showApp(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	err = outputRevisions(ctx, c)
	if err != nil {
		return err
	}

	fmt.Println()

	err = outputVersions(ctx, c)
	if err != nil {
		return err
	}
	return nil
}

func outputVersions(ctx *cli.Context, c *cliclient.MasterClient) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	resource, err := Lookup(c, ctx.Args().First(), "app")
	if err != nil {
		return err
	}

	app, err := c.ProjectClient.App.ByID(resource.ID)
	if err != nil {
		return err
	}

	externalID := app.ExternalID
	if externalID == "" {
		// local folder app doesn't show any version information
		return nil
	}

	externalInfo, err := parseExternalID(externalID)
	if err != nil {
		return err
	}

	template, err := getFilteredTemplate(ctx, c, "cattle-global-data:"+externalInfo["catalog"]+"-"+externalInfo["template"])
	if err != nil {
		return err
	}

	sortedVersions, err := sortTemplateVersions(template)
	if err != nil {
		return err
	}

	if len(sortedVersions) == 0 {
		fmt.Println("No app versions available to install for this version of Rancher server")
		return nil
	}

	writer := NewTableWriter([][]string{
		{"CURRENT", "Current"},
		{"VERSION", "Version"},
	}, ctx)

	defer writer.Close()

	for _, version := range sortedVersions {
		var current string
		if version.String() == externalInfo["version"] {
			current = "*"
		}
		writer.Write(&VersionData{
			Current: current,
			Version: version.String(),
		})
	}
	return writer.Err()
}

func outputRevisions(ctx *cli.Context, c *cliclient.MasterClient) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	resource, err := Lookup(c, ctx.Args().First(), "app")
	if err != nil {
		return err
	}

	app, err := c.ProjectClient.App.ByID(resource.ID)
	if err != nil {
		return err
	}

	revisions := &projectClient.AppRevisionCollection{}
	err = c.ProjectClient.GetLink(*resource, "revision", revisions)
	if err != nil {
		return err
	}

	var sorted revSlice
	for _, rev := range revisions.Data {
		parsedTime, err := time.Parse(time.RFC3339, rev.Created)
		if err != nil {
			return err
		}

		parsedInfo, err := parseTemplateInfo(rev.Status.ExternalID, rev.Status.Files)
		if err != nil {
			return err
		}

		reversionData := revision{
			Name:     rev.Name,
			Created:  parsedTime,
			Catalog:  parsedInfo["catalog"],
			Template: parsedInfo["template"],
			Version:  parsedInfo["version"],
		}
		sorted = append(sorted, reversionData)
	}

	sort.Sort(sorted)

	writer := NewTableWriter([][]string{
		{"CURRENT", "Current"},
		{"REVISION", "Name"},
		{"CATALOG", "Catalog"},
		{"TEMPLATE", "Template"},
		{"VERSION", "Version"},
		{"CREATED", "Human"},
	}, ctx)

	defer writer.Close()

	for _, rev := range sorted {
		if rev.Name == app.AppRevisionID {
			rev.Current = "*"
		}
		rev.Human = rev.Created.Format("02 Jan 2006 15:04:05 MST")

		writer.Write(rev)
	}
	return writer.Err()
}

func templateVersionIDFromVersionLink(s string) string {
	pieces := strings.Split(s, "/")
	return pieces[len(pieces)-1]
}

// parseExternalID gives back a map with the keys catalog, template and version
func parseExternalID(e string) (map[string]string, error) {
	parsed := make(map[string]string)
	u, err := url.Parse(e)
	if err != nil {
		return parsed, err
	}
	q := u.Query()
	for key, value := range q {
		if len(value) > 0 {
			parsed[key] = value[0]
		}
	}
	return parsed, nil
}

// getFilteredTemplate uses the rancherVersion in the template request to get the
// filtered template with incompatable versions dropped
func getFilteredTemplate(ctx *cli.Context, c *cliclient.MasterClient, templateID string) (*managementClient.Template, error) {
	ver, err := getRancherServerVersion(c)
	if err != nil {
		return nil, err
	}

	filter := defaultListOpts(ctx)
	filter.Filters["id"] = templateID
	filter.Filters["rancherVersion"] = ver

	template, err := c.ManagementClient.Template.List(filter)
	if err != nil {
		return nil, err
	}

	if len(template.Data) == 0 {
		return nil, fmt.Errorf("template %v not found", templateID)
	}
	return &template.Data[0], nil
}

// getTemplateLatestVersion returns the newest version of the template
func getTemplateLatestVersion(template *managementClient.Template) (string, error) {
	if len(template.VersionLinks) == 0 {
		return "", errors.New("no versions found for this template (the chart you are trying to install may be intentionally hidden or deprecated for your Rancher version)")
	}
	sorted, err := sortTemplateVersions(template)
	if err != nil {
		return "", err
	}

	return sorted[len(sorted)-1].String(), nil
}

func sortTemplateVersions(template *managementClient.Template) ([]*gover.Version, error) {
	var versions []*gover.Version
	for key := range template.VersionLinks {
		v, err := gover.NewVersion(key)
		if err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}

	sort.Sort(gover.Collection(versions))
	return versions, nil
}

// createNamespace checks if a namespace exists and creates it if needed
func createNamespace(c *cliclient.MasterClient, n string) error {
	filter := defaultListOpts(nil)
	filter.Filters["name"] = n
	namespaces, err := c.ClusterClient.Namespace.List(filter)
	if err != nil {
		return err
	}

	if len(namespaces.Data) == 0 {
		newNamespace := &clusterClient.Namespace{
			Name:      n,
			ProjectID: c.UserConfig.Project,
		}

		ns, err := c.ClusterClient.Namespace.Create(newNamespace)
		if err != nil {
			return err
		}

		nsID := ns.ID
		startTime := time.Now()
		for {
			logrus.Debugf("Namespace create wait - Name: %s, State: %s, Transitioning: %s", ns.Name, ns.State, ns.Transitioning)
			if time.Since(startTime)/time.Second > 30 {
				return fmt.Errorf("timed out waiting for new namespace %s", ns.Name)
			}
			ns, err = c.ClusterClient.Namespace.ByID(nsID)
			if err != nil {
				if e, ok := err.(*clientbase.APIError); ok && e.StatusCode == http.StatusForbidden {
					//the new namespace is created successfully but cannot be got when RBAC rules are not ready.
					continue
				}
				return err
			}

			if ns.State == "active" {
				break
			}

			time.Sleep(500 * time.Millisecond)
		}
	} else {
		if namespaces.Data[0].ProjectID != c.UserConfig.Project {
			return fmt.Errorf("namespace %s already exists", n)
		}
	}
	return nil
}

// processValueInstall creates a map of the values file and fills in missing entries with defaults
func processValueInstall(ctx *cli.Context, tv *managementClient.TemplateVersion, existingValues string) (string, error) {
	values, err := processValues(ctx, existingValues)
	if err != nil {
		return existingValues, err
	}
	// add default values if entries missing from map
	err = fillInDefaultAnswers(tv, values)
	if err != nil {
		return existingValues, err
	}

	// change map back into string to be consistent with ui
	existingValues, err = parseMapToYamlString(values)
	if err != nil {
		return existingValues, err
	}
	return existingValues, nil
}

// processValueUpgrades creates map from existing values and applies updates
func processValueUpgrades(ctx *cli.Context, existingValues string) (string, error) {
	values, err := processValues(ctx, existingValues)
	if err != nil {
		return existingValues, err
	}
	// change map back into string to be consistent with ui
	existingValues, err = parseMapToYamlString(values)
	if err != nil {
		return existingValues, err
	}
	return existingValues, nil
}

// processValues creates a map of the values file
func processValues(ctx *cli.Context, existingValues string) (map[string]interface{}, error) {
	var err error
	values := make(map[string]interface{})
	if existingValues != "" {
		// parse values into map to ensure previous values are considered on update
		values, err = createValuesMap([]byte(existingValues))
		if err != nil {
			return values, err
		}
	}
	if ctx.String("values") != "" {
		// if values file passed in, overwrite defaults with new key value pair
		values, err = parseFile(ctx.String("values"))
		if err != nil {
			return values, err
		}
	}
	return values, nil
}

// processAnswerInstall adds answers to given map, and prompts users to answers chart questions if interactive is true
func processAnswerInstall(
	ctx *cli.Context,
	tv *managementClient.TemplateVersion,
	answers,
	answersSetString map[string]string,
	interactive bool,
	multicluster bool,
) (map[string]string, map[string]string, error) {
	var err error
	answers, answersSetString, err = processAnswerUpdates(ctx, answers, answersSetString)
	if err != nil {
		return answers, answersSetString, err
	}
	// interactive occurs before adding defaults to ensure all questions are asked
	if interactive {
		// answers to questions will be added to map
		err := askQuestions(tv, answers)
		if err != nil {
			return answers, answersSetString, err
		}
	}
	if multicluster && !interactive {
		// add default values if answers missing from map
		err = fillInDefaultAnswersStringMap(tv, answers)
		if err != nil {
			return answers, answersSetString, err
		}
	}
	return answers, answersSetString, nil
}

func processAnswerUpdates(ctx *cli.Context, answers, answersSetString map[string]string) (map[string]string, map[string]string, error) {
	logrus.Println("ok")
	if answers == nil || ctx.Bool("reset") {
		// this would not be possible without returning a map
		answers = make(map[string]string)
	}
	if answersSetString == nil || ctx.Bool("reset") {
		// this would not be possible without returning a map
		answersSetString = make(map[string]string)
	}
	if ctx.String("answers") != "" {
		err := parseAnswersFile(ctx.String("answers"), answers)
		if err != nil {
			return answers, answersSetString, err
		}
	}
	for _, answer := range ctx.StringSlice("set") {
		parts := strings.SplitN(answer, "=", 2)
		if len(parts) == 2 {
			answers[parts[0]] = parts[1]
		}
	}
	for _, answer := range ctx.StringSlice("set-string") {
		parts := strings.SplitN(answer, "=", 2)
		logrus.Printf("%v\n", parts)
		if len(parts) == 2 {
			answersSetString[parts[0]] = parts[1]
		}
	}
	return answers, answersSetString, nil
}

// parseMapToYamlString create yaml string from answers map
func parseMapToYamlString(answerMap map[string]interface{}) (string, error) {
	yamlFileString, err := yaml.Marshal(answerMap)
	if err != nil {
		return "", err
	}
	return string(yamlFileString), nil
}

func parseAnswersFile(location string, answers map[string]string) error {
	holder, err := parseFile(location)
	if err != nil {
		return err
	}
	for key, value := range holder {
		switch value.(type) {
		case nil:
			answers[key] = ""
		default:
			answers[key] = fmt.Sprintf("%v", value)
		}
	}
	return nil
}

func parseFile(location string) (map[string]interface{}, error) {
	bytes, err := os.ReadFile(location)
	if err != nil {
		return nil, err
	}
	return createValuesMap(bytes)
}

func createValuesMap(bytes []byte) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	if hasPrefix(bytes, []byte("{")) {
		// this is the check that "readFileReturnJSON" uses to differentiate between JSON and YAML
		if err := json.Unmarshal(bytes, &values); err != nil {
			return nil, err
		}
	} else {
		if err := yaml.Unmarshal(bytes, &values); err != nil {
			return nil, err
		}
	}
	return values, nil
}

func askQuestions(tv *managementClient.TemplateVersion, answers map[string]string) error {
	var asked bool
	var attempts int
	if tv == nil {
		return nil
	}
	for {
		attempts++
		for _, question := range tv.Questions {
			if _, ok := answers[question.Variable]; !ok && checkShowIfStringMap(question.ShowIf, answers) {
				asked = true
				answers[question.Variable] = askQuestion(question)
				if checkShowSubquestionIfStringMap(question, answers) {
					for _, subQuestion := range question.Subquestions {
						// only ask the question if there is not an answer and it passes the ShowIf check
						if _, ok := answers[subQuestion.Variable]; !ok && checkShowIfStringMap(subQuestion.ShowIf, answers) {
							answers[subQuestion.Variable] = askSubQuestion(subQuestion)
						}
					}
				}
			}
		}
		if !asked {
			return nil
		} else if attempts >= 10 {
			return errors.New("attempted questions 10 times")
		}
		asked = false
	}
}

func askQuestion(q managementClient.Question) string {
	if len(q.Description) > 0 {
		fmt.Printf("\nDescription: %s\n", q.Description)
	}

	if len(q.Options) > 0 {
		options := strings.Join(q.Options, ", ")
		fmt.Printf("Accepted Options: %s\n", options)
	}

	fmt.Printf("Name: %s\nVariable Name: %s\nDefault:[%s]\nEnter answer or 'return' for default:", q.Label, q.Variable, q.Default)

	answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return ""
	}

	answer = strings.TrimSpace(answer)
	if answer == "" {
		answer = q.Default
	}

	return answer
}

func askSubQuestion(q managementClient.SubQuestion) string {
	if len(q.Description) > 0 {
		fmt.Printf("\nDescription: %s\n", q.Description)
	}

	if len(q.Options) > 0 {
		options := strings.Join(q.Options, ", ")
		fmt.Printf("Accepted Options: %s\n", options)
	}

	fmt.Printf("Name: %s\nVariable Name: %s\nDefault:[%s]\nEnter answer or 'return' for default:", q.Label, q.Variable, q.Default)

	answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return ""
	}

	answer = strings.TrimSpace(answer)
	if answer == "" {
		answer = q.Default
	}

	return answer
}

// fillInDefaultAnswers parses through questions and creates an answer map with default answers if missing from map
func fillInDefaultAnswers(tv *managementClient.TemplateVersion, answers map[string]interface{}) error {
	if tv == nil {
		return nil
	}
	for _, question := range tv.Questions {
		if _, ok := answers[question.Variable]; !ok && checkShowIf(question.ShowIf, answers) {
			answers[question.Variable] = question.Default
			if checkShowSubquestionIf(question, answers) {
				for _, subQuestion := range question.Subquestions {
					// set the sub-question if the showIf check passes
					if _, ok := answers[subQuestion.Variable]; !ok && checkShowIf(subQuestion.ShowIf, answers) {
						answers[subQuestion.Variable] = subQuestion.Default
					}
				}
			}
		}
	}
	if answers == nil {
		return errors.New("could not generate default answers")
	}
	return nil
}

// checkShowIf uses the ShowIf field to determine if a question should be asked
// this field comes in the format <key>=<value> where key is a question id and value is the answer
func checkShowIf(s string, answers map[string]interface{}) bool {
	// No ShowIf so always ask the question
	if len(s) == 0 {
		return true
	}

	pieces := strings.Split(s, "=")
	if len(pieces) != 2 {
		return false
	}

	//if the key exists and the val matches the expression ask the question
	if val, ok := answers[pieces[0]]; ok && fmt.Sprintf("%v", val) == pieces[1] {
		return true
	}
	return false
}

// fillInDefaultAnswersStringMap parses through questions and creates an answer map with default answers if missing from map
func fillInDefaultAnswersStringMap(tv *managementClient.TemplateVersion, answers map[string]string) error {
	if tv == nil {
		return nil
	}
	for _, question := range tv.Questions {
		if _, ok := answers[question.Variable]; !ok && checkShowIfStringMap(question.ShowIf, answers) {
			answers[question.Variable] = question.Default
			if checkShowSubquestionIfStringMap(question, answers) {
				for _, subQuestion := range question.Subquestions {
					// set the sub-question if the showIf check passes
					if _, ok := answers[subQuestion.Variable]; !ok && checkShowIfStringMap(subQuestion.ShowIf, answers) {
						answers[subQuestion.Variable] = subQuestion.Default
					}
				}
			}
		}
	}
	if answers == nil {
		return errors.New("could not generate default answers")
	}
	return nil
}

// checkShowIfStringMap uses the ShowIf field to determine if a question should be asked
// this field comes in the format <key>=<value> where key is a question id and value is the answer
func checkShowIfStringMap(s string, answers map[string]string) bool {
	// No ShowIf so always ask the question
	if len(s) == 0 {
		return true
	}

	pieces := strings.Split(s, "=")
	if len(pieces) != 2 {
		return false
	}

	//if the key exists and the val matches the expression ask the question
	if val, ok := answers[pieces[0]]; ok && val == pieces[1] {
		return true
	}
	return false
}

func checkShowSubquestionIf(q managementClient.Question, answers map[string]interface{}) bool {
	if val, ok := answers[q.Variable]; ok {
		if fmt.Sprintf("%v", val) == q.ShowSubquestionIf {
			return true
		}
	}
	return false
}

func checkShowSubquestionIfStringMap(q managementClient.Question, answers map[string]string) bool {
	if val, ok := answers[q.Variable]; ok {
		if val == q.ShowSubquestionIf {
			return true
		}
	}
	return false
}

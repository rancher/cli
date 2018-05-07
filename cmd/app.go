package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	gover "github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/rancher/cli/cliclient"
	clusterClient "github.com/rancher/types/client/cluster/v3"
	managementClient "github.com/rancher/types/client/management/v3"
	projectClient "github.com/rancher/types/client/project/v3"
	"github.com/urfave/cli"
)

const (
	installAppDescription = `
Install an app template in the current Rancher server. This defaults to the newest version of the app template.
Specify a version using '--version' if required.
					
Example:
	# Install the redis template with no other options
	$ rancher app install redis appFoo

	# Install the redis template and specify an answers file location
	$ rancher app install --answers /example/answers.yaml redis appFoo

	# Install the redis template and set multiple answers and the version to install
	$ rancher app install --set foo=bar --set baz=bunk --version 1.0.1 redis appFoo
`
)

type AppData struct {
	ID      string
	App     projectClient.App
	Version string
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
	Current string
	Name    string
	Created time.Time
	Human   string
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
		Usage:   "Operations with apps",
		Action:  defaultAction(appLs),
		Flags:   appLsFlags,
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "ls",
				Usage:       "List apps",
				Description: "\nList all apps in the current Rancher server",
				ArgsUsage:   "None",
				Action:      appLs,
				Flags:       appLsFlags,
			},
			cli.Command{
				Name:      "delete",
				Usage:     "Delete an app",
				Action:    appDelete,
				ArgsUsage: "[APP_NAME/APP_ID]",
			},
			cli.Command{
				Name:        "install",
				Usage:       "Install an app template",
				Description: installAppDescription,
				Action:      templateInstall,
				ArgsUsage:   "[TEMPLATE_NAME, APP_NAME]...",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "answers,a",
						Usage: "Path to an answers file, the format of the file is a map with key:value. This supports JSON and YAML.",
					},
					cli.StringSliceFlag{
						Name:  "set",
						Usage: "Set answers for the template, can be used multiple times. Example: --set foo=bar",
					},
					cli.StringFlag{
						Name:  "version",
						Usage: "Version of the template to use",
					},
				},
			},
			cli.Command{
				Name:      "rollback",
				Usage:     "Rollback an app to a previous version",
				Action:    appRollback,
				ArgsUsage: "[APP_NAME/APP_ID, REVISION_ID/REVISION_NAME]",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "show-revisions,r",
						Usage: "Show revisions available to rollback to",
					},
				},
			},
			cli.Command{
				Name:      "upgrade",
				Usage:     "Upgrade an app to a newer version",
				Action:    appUpgrade,
				ArgsUsage: "[APP_NAME/APP_ID VERSION]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "answers,a",
						Usage: "Path to an answers file, the format of the file is a map with key:value. Supports JSON and YAML",
					},
					cli.StringSliceFlag{
						Name:  "set",
						Usage: "Set answers for the template, can be used multiple times. Example: --set foo=bar",
					},
					cli.BoolFlag{
						Name:  "show-versions,v",
						Usage: "Display versions available to upgrade to",
					},
				},
			},
			cli.Command{
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
			cli.Command{
				Name:        "show-template",
				Aliases:     []string{"st"},
				Usage:       "Show versions available to install for an app template",
				Description: "\nShow all available versions of an app template",
				ArgsUsage:   "[TEMPLATE_ID]",
				Action:      templateShow,
			},
			cli.Command{
				Name:      "show-app",
				Aliases:   []string{"sa"},
				Usage:     "Show an app's available versions and revisions",
				ArgsUsage: "[APP_NAME/APP_ID]",
				Action:    showApp,
				Flags: []cli.Flag{
					formatFlag,
				},
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
	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "App.Name"},
		{"STATE", "App.State"},
		{"VERSION", "Version"},
	}, ctx)

	defer writer.Close()

	for _, item := range collection.Data {
		parsedExternal, err := parseExternalID(item.ExternalID)
		if err != nil {
			return err
		}
		writer.Write(&AppData{
			ID:      item.ID,
			App:     item,
			Version: parsedExternal["version"],
		})
	}
	return writer.Err()

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

	if ctx.Bool("versions") {
		return outputVersions(ctx, c)
	}

	if ctx.NArg() < 2 {
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

	answers := app.Answers
	err = processAnswers(ctx, c, nil, answers, false)
	if err != nil {
		return err
	}

	u, err := url.Parse(app.ExternalID)
	if err != nil {
		return err
	}

	q := u.Query()

	q.Set("version", ctx.Args().Get(1))

	filter := defaultListOpts(ctx)
	filter.Filters["externalId"] = "catalog://?" + q.Encode()

	template, err := c.ManagementClient.TemplateVersion.List(filter)
	if err != nil {
		return err
	}

	if len(template.Data) == 0 {
		return fmt.Errorf("version %s not valid for app", ctx.Args().Get(1))
	}

	au := &projectClient.AppUpgradeConfig{
		Answers:    answers,
		ExternalID: template.Data[0].ExternalID,
	}

	return c.ProjectClient.App.ActionUpgrade(app, au)
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
		RevisionId: revision.Name,
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

	template, err := c.ManagementClient.Template.ByID(resource.ID)
	if err != nil {
		return err
	}

	sortedVersions, err := sortTemplateVersions(template)
	if err != nil {
		return err
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

	resource, err := Lookup(c, templateName, "template")
	if err != nil {
		return err
	}

	template, err := c.ManagementClient.Template.ByID(resource.ID)

	templateVersionID := templateVersionIDFromVersionLink(template.VersionLinks[template.DefaultVersion])
	userVersion := ctx.String("version")
	if userVersion != "" {
		if link, ok := template.VersionLinks[userVersion]; ok {
			templateVersionID = templateVersionIDFromVersionLink(link)
		} else {
			return fmt.Errorf(
				"version %s for template %s is invalid, run 'rancher show-template %s' for a list of versions",
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

	answers := make(map[string]string)
	err = processAnswers(ctx, c, templateVersion, answers, true)
	if err != nil {
		return err
	}

	err = createNamespace(c, template.Name)
	if err != nil {
		return err
	}

	app := &projectClient.App{
		Answers:         answers,
		ExternalID:      templateVersion.ExternalID,
		Name:            appName,
		TargetNamespace: template.Name,
	}

	madeApp, err := c.ProjectClient.App.Create(app)
	if err != nil {
		return err
	}

	// wait for the app's notes to be populated so we can print them
	var timeout int
	for len(madeApp.Notes) == 0 {
		if timeout == 60 {
			return errors.New("timed out waiting for app notes, the app could still be installing. Run 'rancher apps' to verify")
		}
		timeout++
		time.Sleep(2 * time.Second)
		madeApp, err = c.ProjectClient.App.ByID(madeApp.ID)
		if err != nil {
			return err
		}
	}

	if len(madeApp.Notes) != 0 {
		fmt.Println(madeApp.Notes)
	}

	return nil
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

	externalInfo, err := parseExternalID(app.ExternalID)
	if err != nil {
		return err
	}

	template, err := c.ManagementClient.Template.ByID(externalInfo["catalog"] + "-" + externalInfo["template"])

	writer := NewTableWriter([][]string{
		{"CURRENT", "Current"},
		{"VERSION", "Version"},
	}, ctx)

	defer writer.Close()

	sortedVersions, err := sortTemplateVersions(template)
	if err != nil {
		return err
	}

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
		if nil != err {
			return err
		}
		sorted = append(sorted, revision{Name: rev.Name, Created: parsedTime})
	}

	sort.Sort(sorted)

	writer := NewTableWriter([][]string{
		{"CURRENT", "Current"},
		{"REVISION", "Name"},
		{"CREATED", "Human"},
	}, ctx)

	defer writer.Close()

	for _, rev := range sorted {
		if rev.Name == app.AppRevisionId {
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

		_, err = c.ClusterClient.Namespace.Create(newNamespace)
		if err != nil {
			return err
		}
	}
	return nil
}

func processAnswers(
	ctx *cli.Context,
	c *cliclient.MasterClient,
	tv *managementClient.TemplateVersion,
	answers map[string]string,
	interactive bool,
) error {
	if ctx.String("answers") != "" {
		err := getAnswersFile(ctx.String("answers"), answers)
		if err != nil {
			return err
		}
	}

	for _, answer := range ctx.StringSlice("set") {
		parts := strings.Split(answer, "=")
		if len(parts) == 2 {
			answers[parts[0]] = parts[1]
		}
	}

	if interactive {
		err := askQuestions(tv, answers)
		if err != nil {
			return err
		}
	}

	return nil
}

func getAnswersFile(location string, answers map[string]string) error {
	bytes, err := readFileReturnJSON(location)
	if err != nil {
		return err
	}

	holder := make(map[string]interface{})
	err = json.Unmarshal(bytes, &holder)
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

func askQuestions(tv *managementClient.TemplateVersion, answers map[string]string) error {
	var asked bool
	var attempts int
	for {
		attempts++
		for _, question := range tv.Questions {
			// only ask the question if there is not an answer and it passes the ShowIf check
			if _, ok := answers[question.Variable]; !ok && checkShowIf(question.ShowIf, answers) {
				asked = true
				answers[question.Variable] = askQuestion(question)
				if checkShowSubquestionIf(question, answers) {
					for _, subQuestion := range question.Subquestions {
						// only ask the question if there is not an answer and it passes the ShowIf check
						if _, ok := answers[subQuestion.Variable]; !ok && checkShowIf(subQuestion.ShowIf, answers) {
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

// checkShowIf uses the ShowIf field to determine if a question should be asked
// this field comes in the format <key>=<value> where key is a question id and value is the answer
func checkShowIf(s string, answers map[string]string) bool {
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

func checkShowSubquestionIf(q managementClient.Question, answers map[string]string) bool {
	if val, ok := answers[q.Variable]; ok {
		if val == q.ShowSubquestionIf {
			return true
		}
	}
	return false
}

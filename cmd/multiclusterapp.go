package cmd

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/rancher/cli/cliclient"
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/slice"
	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	installMultiClusterAppDescription = `
Install a multi-cluster app in the current Rancher server. This defaults to the newest version of the app template.
Specify a version using '--version' if required.

Example:
	# Install the redis template with no other options
	$ rancher multiclusterapp install redis appFoo

	# Install the redis template and specify an answers file location
	$ rancher multiclusterapp install --answers /example/answers.yaml redis appFoo

	# Install the redis template and set multiple answers and the version to install
	$ rancher multiclusterapp install --set foo=bar --set-string baz=bunk --version 1.0.1 redis appFoo

	# Install the redis template and set target projects to install
	$ rancher multiclusterapp install --target mycluster:Default --target c-98pjr:p-w6c5f redis appFoo

	# Block cli until installation has finished or encountered an error. Use after multiclusterapp install.
	$ rancher wait <multiclusterapp-id>
`
	upgradeStrategySimultaneously = "simultaneously"
	upgradeStrategyRollingUpdate  = "rolling-update"
	argUpgradeStrategy            = "upgrade-strategy"
	argUpgradeBatchSize           = "upgrade-batch-size"
	argUpgradeBatchInterval       = "upgrade-batch-interval"
)

var (
	memberAccessTypes = []string{"owner", "member", "read-only"}
	upgradeStrategies = []string{upgradeStrategySimultaneously, upgradeStrategyRollingUpdate}
)

type MultiClusterAppData struct {
	ID      string
	App     managementClient.MultiClusterApp
	Version string
	Targets string
}

type scopeAnswers struct {
	Answers          map[string]string
	AnswersSetString map[string]string
}

func MultiClusterAppCommand() cli.Command {
	appLsFlags := []cli.Flag{
		formatFlag,
		cli.BoolFlag{
			Name:  "quiet,q",
			Usage: "Only display IDs",
		},
	}

	return cli.Command{
		Name:    "multiclusterapps",
		Aliases: []string{"multiclusterapp", "mcapps", "mcapp"},
		Usage:   "Operations with multi-cluster apps",
		Action:  defaultAction(multiClusterAppLs),
		Flags:   appLsFlags,
		Subcommands: []cli.Command{
			{
				Name:        "ls",
				Usage:       "List multi-cluster apps",
				Description: "\nList all multi-cluster apps in the current Rancher server",
				ArgsUsage:   "None",
				Action:      multiClusterAppLs,
				Flags:       appLsFlags,
			},
			{
				Name:      "delete",
				Usage:     "Delete a multi-cluster app",
				Action:    multiClusterAppDelete,
				ArgsUsage: "[APP_NAME]",
			},
			{
				Name:        "install",
				Usage:       "Install a multi-cluster app",
				Description: installMultiClusterAppDescription,
				Action:      multiClusterAppTemplateInstall,
				ArgsUsage:   "[TEMPLATE_NAME, APP_NAME]...",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "answers,a",
						Usage: "Path to an answers file, the format of the file is a map with key:value. This supports JSON and YAML.",
					},
					cli.StringFlag{
						Name:  "values",
						Usage: "Path to a helm values file.",
					},
					cli.StringSliceFlag{
						Name: "set",
						Usage: "Set answers for the template, can be used multiple times. You can set overriding answers for specific clusters or projects " +
							"by providing cluster ID or project ID as the prefix. Example: --set foo=bar --set c-rvcrl:foo=bar --set c-rvcrl:p-8w2x8:foo=bar",
					},
					cli.StringSliceFlag{
						Name: "set-string",
						Usage: "Set string answers for the template (Skips Helm's type conversion), can be used multiple times. You can set overriding answers for specific clusters or projects " +
							"by providing cluster ID or project ID as the prefix. Example: --set-string foo=bar --set-string c-rvcrl:foo=bar --set-string c-rvcrl:p-8w2x8:foo=bar",
					},
					cli.StringFlag{
						Name:  "version",
						Usage: "Version of the template to use",
					},
					cli.BoolFlag{
						Name:  "no-prompt",
						Usage: "Suppress asking questions and use the default values when required answers are not provided",
					},
					cli.StringSliceFlag{
						Name:  "target,t",
						Usage: "Target project names/ids to install the app into",
					},
					cli.StringSliceFlag{
						Name: "role",
						Usage: "Set roles required to launch/manage the apps in target projects. For example, set \"project-member\" role when the app needs to manage resources " +
							"in the projects in which it is deployed. Or set \"cluster-owner\" role when the app needs to manage resources in the clusters in which it is deployed. " +
							"(default: \"project-member\")",
					},
					cli.StringSliceFlag{
						Name:  "member",
						Usage: "Set members of the app, with the same access type defined by --member-access-type",
					},
					cli.StringFlag{
						Name:  "member-access-type",
						Usage: "Access type of the members. Specify only one value, and it applies to all members defined by --member. Valid options are 'owner', 'member' and 'read-only'",
						Value: "owner",
					},
					cli.StringFlag{
						Name:  argUpgradeStrategy,
						Usage: "Strategy for upgrade. Valid options are \"rolling-update\" and \"simultaneously\"",
						Value: upgradeStrategySimultaneously,
					},
					cli.Int64Flag{
						Name:  argUpgradeBatchSize,
						Usage: "The number of apps in target projects to be upgraded at a time.  Only used if --upgrade-strategy is rolling-update.",
						Value: 1,
					},
					cli.Int64Flag{
						Name:  argUpgradeBatchInterval,
						Usage: "The number of seconds between updating the next app during upgrade.  Only used if --upgrade-strategy is rolling-update.",
						Value: 1,
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
				Usage:     "Rollback a multi-cluster app to a previous version",
				Action:    multiClusterAppRollback,
				ArgsUsage: "[APP_NAME/APP_ID, REVISION_ID/REVISION_NAME]",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "show-revisions,r",
						Usage: "Show revisions available to rollback to",
					},
				},
			},
			{
				Name:      "upgrade",
				Usage:     "Upgrade an app to a newer version",
				Action:    multiClusterAppUpgrade,
				ArgsUsage: "[APP_NAME/APP_ID VERSION]",
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
						Name: "set",
						Usage: "Set answers for the template, can be used multiple times. You can set overriding answers for specific clusters or projects " +
							"by providing cluster ID or project ID as the prefix. Example: --set foo=bar --set c-rvcrl:foo=bar --set c-rvcrl:p-8w2x8:foo=bar",
					},
					cli.StringSliceFlag{
						Name: "set-string",
						Usage: "Set string answers for the template (Skips Helm's type conversion), can be used multiple times. You can set overriding answers for specific clusters or projects " +
							"by providing cluster ID or project ID as the prefix. Example: --set-string foo=bar --set-string c-rvcrl:foo=bar --set-string c-rvcrl:p-8w2x8:foo=bar",
					},
					cli.BoolFlag{
						Name:  "reset",
						Usage: "Reset all catalog app answers",
					},
					cli.StringSliceFlag{
						Name: "role,r",
						Usage: "Set roles required to launch/manage the apps in target projects. Specified roles on upgrade will override all the original roles. " +
							"For example, provide all existing roles if you want to add additional roles. Leave it empty to keep current roles",
					},
					cli.BoolFlag{
						Name:  "show-versions,v",
						Usage: "Display versions available to upgrade to",
					},
					cli.StringFlag{
						Name:  argUpgradeStrategy,
						Usage: "Strategy for upgrade. Valid options are \"rolling-update\" and \"simultaneously\"",
					},
					cli.Int64Flag{
						Name:  argUpgradeBatchSize,
						Usage: "The number of apps in target projects to be upgraded at a time.  Only used if --upgrade-strategy is rolling-update.",
					},
					cli.Int64Flag{
						Name:  argUpgradeBatchInterval,
						Usage: "The number of seconds between updating the next app during upgrade.  Only used if --upgrade-strategy is rolling-update.",
					},
				},
			},
			{
				Name:        "add-project",
				Usage:       "Add target projects to a multi-cluster app",
				Action:      addMcappTargetProject,
				Description: "Examples:\n #Add 'p1' project in cluster 'mycluster' to target projects of a multi-cluster app named 'myapp'\n rancher multiclusterapp add-project myapp mycluster:p1\n",
				ArgsUsage:   "[APP_NAME/APP_ID, CLUSTER_NAME:PROJECT_NAME/PROJECT_ID...]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "answers,a",
						Usage: "Path to an answers file that provides overriding answers for the new target projects, the format of the file is a map with key:value. Supports JSON and YAML",
					},
					cli.StringFlag{
						Name:  "values",
						Usage: "Path to a helm values file that provides overriding answers for the new target projects",
					},
					cli.StringSliceFlag{
						Name:  "set",
						Usage: "Set overriding answers for the new target projects",
					},
					cli.StringSliceFlag{
						Name:  "set-string",
						Usage: "Set overriding string answers for the new target projects",
					},
				},
			},
			{
				Name:        "delete-project",
				Usage:       "Delete target projects from a multi-cluster app",
				Action:      deleteMcappTargetProject,
				Description: "Examples:\n #Delete 'p1' project in cluster 'mycluster' from target projects of a multi-cluster app named 'myapp'\n rancher multiclusterapp delete-project myapp mycluster:p1\n",
				ArgsUsage:   "[APP_NAME/APP_ID, CLUSTER_NAME:PROJECT_NAME/PROJECT_ID...]",
			},
			{
				Name:        "add-member",
				Usage:       "Add members to a multi-cluster app",
				Action:      addMcappMember,
				Description: "Examples:\n #Add 'user1' and 'user2' as the owners of a multi-cluster app named 'myapp'\n rancher multiclusterapp add-member myapp owner user1 user2\n",
				ArgsUsage:   "[APP_NAME/APP_ID, ACCESS_TYPE, USER_NAME/USER_ID...]",
			},
			{
				Name:        "delete-member",
				Usage:       "Delete members from a multi-cluster app",
				Action:      deleteMcappMember,
				Description: "Examples:\n #Delete the membership of a user named 'user1' from a multi-cluster app named 'myapp'\n rancher multiclusterapp delete-member myapp user1\n",
				ArgsUsage:   "[APP_NAME/APP_ID, USER_NAME/USER_ID...]",
			},
			{
				Name:      "list-members",
				Aliases:   []string{"lm"},
				Usage:     "List current members of a multi-cluster app",
				ArgsUsage: "[APP_NAME/APP_ID]",
				Action:    listMultiClusterAppMembers,
				Flags: []cli.Flag{
					formatFlag,
				},
			},
			{
				Name:      "list-answers",
				Aliases:   []string{"la"},
				Usage:     "List current answers of a multi-cluster app",
				ArgsUsage: "[APP_NAME/APP_ID]",
				Action:    listMultiClusterAppAnswers,
				Flags: []cli.Flag{
					formatFlag,
				},
			},
			{
				Name:        "list-templates",
				Aliases:     []string{"lt"},
				Usage:       "List templates available for installation",
				Description: "\nList all app templates in the current Rancher server",
				ArgsUsage:   "None",
				Action:      globalTemplateLs,
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
				Action:    showMultiClusterApp,
				Flags: []cli.Flag{
					formatFlag,
					cli.BoolFlag{
						Name:  "show-roles",
						Usage: "Show roles required to manage the app",
					},
				},
			},
		},
	}
}

func multiClusterAppLs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	collection, err := c.ManagementClient.MultiClusterApp.List(defaultListOpts(ctx))
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "App.Name"},
		{"STATE", "App.State"},
		{"VERSION", "Version"},
		{"TARGET_PROJECTS", "Targets"},
	}, ctx)

	defer writer.Close()

	clusterCache, projectCache, err := getClusterProjectMap(ctx, c.ManagementClient)
	if err != nil {
		return err
	}

	templateVersionCache := make(map[string]string)
	for _, item := range collection.Data {
		version, err := getTemplateVersion(c.ManagementClient, templateVersionCache, item.TemplateVersionID)
		if err != nil {
			return err
		}
		targetNames := getReadableTargetNames(clusterCache, projectCache, item.Targets)
		writer.Write(&MultiClusterAppData{
			ID:      item.ID,
			App:     item,
			Version: version,
			Targets: strings.Join(targetNames, ","),
		})
	}
	return writer.Err()
}

func getTemplateVersion(client *managementClient.Client, templateVersionCache map[string]string, ID string) (string, error) {
	var version string
	if cachedVersion, ok := templateVersionCache[ID]; ok {
		version = cachedVersion
	} else {
		templateVersion, err := client.TemplateVersion.ByID(ID)
		if err != nil {
			return "", err
		}
		templateVersionCache[templateVersion.ID] = templateVersion.Version
		version = templateVersion.Version
	}
	return version, nil
}

func getClusterProjectMap(ctx *cli.Context, client *managementClient.Client) (map[string]managementClient.Cluster, map[string]managementClient.Project, error) {
	clusters := make(map[string]managementClient.Cluster)
	clusterCollectionData, err := listAllClusters(ctx, client)
	if err != nil {
		return nil, nil, err
	}
	for _, c := range clusterCollectionData {
		clusters[c.ID] = c
	}
	projects := make(map[string]managementClient.Project)
	projectCollectionData, err := listAllProjects(ctx, client)
	if err != nil {
		return nil, nil, err
	}
	for _, p := range projectCollectionData {
		projects[p.ID] = p
	}
	return clusters, projects, nil
}

func listAllClusters(ctx *cli.Context, client *managementClient.Client) ([]managementClient.Cluster, error) {
	clusterCollection, err := client.Cluster.List(defaultListOpts(ctx))
	if err != nil {
		return nil, err
	}
	clusterCollectionData := clusterCollection.Data
	for {
		clusterCollection, err = clusterCollection.Next()
		if err != nil {
			return nil, err
		}
		if clusterCollection == nil {
			break
		}
		clusterCollectionData = append(clusterCollectionData, clusterCollection.Data...)
		if !clusterCollection.Pagination.Partial {
			break
		}
	}
	return clusterCollectionData, nil
}

func listAllProjects(ctx *cli.Context, client *managementClient.Client) ([]managementClient.Project, error) {
	projectCollection, err := client.Project.List(defaultListOpts(ctx))
	if err != nil {
		return nil, err
	}
	projectCollectionData := projectCollection.Data
	for {
		projectCollection, err = projectCollection.Next()
		if err != nil {
			return nil, err
		}
		if projectCollection == nil {
			break
		}
		projectCollectionData = append(projectCollectionData, projectCollection.Data...)
		if !projectCollection.Pagination.Partial {
			break
		}
	}
	return projectCollectionData, nil
}

func getReadableTargetNames(clusterCache map[string]managementClient.Cluster, projectCache map[string]managementClient.Project, targets []managementClient.Target) []string {
	var targetNames []string
	for _, target := range targets {
		projectID := target.ProjectID
		clusterID, _ := parseScope(projectID)
		cluster, ok := clusterCache[clusterID]
		if !ok {
			logrus.Debugf("Cannot get readable name for target %q, showing ID", target.ProjectID)
			targetNames = append(targetNames, target.ProjectID)
			continue
		}
		project, ok := projectCache[projectID]
		if !ok {
			logrus.Debugf("Cannot get readable name for target %q, showing ID", target.ProjectID)
			targetNames = append(targetNames, target.ProjectID)
			continue
		}
		targetNames = append(targetNames, concatScope(cluster.Name, project.Name))
	}
	return targetNames
}

func multiClusterAppDelete(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	for _, name := range ctx.Args() {
		_, app, err := searchForMcapp(c, name)
		if err != nil {
			return err
		}

		err = c.ManagementClient.MultiClusterApp.Delete(app)
		if err != nil {
			return err
		}
	}

	return nil
}

func multiClusterAppUpgrade(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	if ctx.Bool("show-versions") {
		if ctx.NArg() == 0 {
			return cli.ShowSubcommandHelp(ctx)
		}

		_, app, err := searchForMcapp(c, ctx.Args().First())
		if err != nil {
			return err
		}

		return outputMultiClusterAppVersions(ctx, c, app)
	}

	if ctx.NArg() != 2 {
		return cli.ShowSubcommandHelp(ctx)
	}

	upgradeStrategy := strings.ToLower(ctx.String(argUpgradeStrategy))
	if ctx.IsSet(argUpgradeStrategy) && !slice.ContainsString(upgradeStrategies, upgradeStrategy) {
		return fmt.Errorf("invalid upgrade-strategy %q, supported values are \"rolling-update\" and \"simultaneously\"", upgradeStrategy)
	}

	_, app, err := searchForMcapp(c, ctx.Args().First())
	if err != nil {
		return err
	}

	update := make(map[string]interface{})
	answers, answersSetString := fromMultiClusterAppAnswers(app.Answers)
	answers, answersSetString, err = processAnswerUpdates(ctx, answers, answersSetString)
	if err != nil {
		return err
	}
	update["answers"], err = toMultiClusterAppAnswers(c, answers, answersSetString)
	if err != nil {
		return err
	}

	version := ctx.Args().Get(1)
	templateVersion, err := c.ManagementClient.TemplateVersion.ByID(app.TemplateVersionID)
	if err != nil {
		return err
	}
	toUpgradeTemplateversionID := strings.TrimSuffix(templateVersion.ID, templateVersion.Version) + version
	// Check if the template version is valid before applying it
	_, err = c.ManagementClient.TemplateVersion.ByID(toUpgradeTemplateversionID)
	if err != nil {
		templateName := strings.TrimSuffix(toUpgradeTemplateversionID, "-"+version)
		return fmt.Errorf(
			"version %s for template %s is invalid, run 'rancher mcapp show-template %s' for available versions",
			version,
			templateName,
			templateName,
		)
	}
	update["templateVersionId"] = toUpgradeTemplateversionID

	roles := ctx.StringSlice("role")
	if len(roles) > 0 {
		update["roles"] = roles
	} else {
		update["roles"] = app.Roles
	}

	if upgradeStrategy == upgradeStrategyRollingUpdate {
		update["upgradeStrategy"] = &managementClient.UpgradeStrategy{
			RollingUpdate: &managementClient.RollingUpdate{
				BatchSize: ctx.Int64(argUpgradeBatchSize),
				Interval:  ctx.Int64(argUpgradeBatchInterval),
			},
		}
	} else if upgradeStrategy == upgradeStrategySimultaneously {
		update["upgradeStrategy"] = nil
	}

	if _, err := c.ManagementClient.MultiClusterApp.Update(app, update); err != nil {
		return err
	}

	return nil
}

func multiClusterAppRollback(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	resource, app, err := searchForMcapp(c, ctx.Args().First())
	if err != nil {
		return err
	}

	if ctx.Bool("show-revisions") {
		return outputMultiClusterAppRevisions(ctx, c, resource, app)
	}

	if ctx.NArg() != 2 {
		return cli.ShowSubcommandHelp(ctx)
	}

	revisionResource, err := Lookup(c, ctx.Args().Get(1), managementClient.MultiClusterAppRevisionType)
	if err != nil {
		return err
	}

	rr := &managementClient.MultiClusterAppRollbackInput{
		RevisionID: revisionResource.ID,
	}

	if err := c.ManagementClient.MultiClusterApp.ActionRollback(app, rr); err != nil {
		return err
	}

	return nil
}

func multiClusterAppTemplateInstall(ctx *cli.Context) error {
	if ctx.NArg() > 2 {
		return cli.ShowSubcommandHelp(ctx)
	}

	templateName := ctx.Args().First()
	appName := ctx.Args().Get(1)

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	roles := ctx.StringSlice("role")
	if len(roles) == 0 {
		// Handle the default here because the cli default value for stringSlice do not get overridden.
		roles = []string{"project-member"}
	}

	app := &managementClient.MultiClusterApp{
		Name:  appName,
		Roles: roles,
	}

	upgradeStrategy := strings.ToLower(ctx.String(argUpgradeStrategy))
	if !slice.ContainsString(upgradeStrategies, upgradeStrategy) {
		return fmt.Errorf("invalid upgrade-strategy %q, supported values are \"rolling-update\" and \"simultaneously\"", upgradeStrategy)
	} else if upgradeStrategy == upgradeStrategyRollingUpdate {
		app.UpgradeStrategy = &managementClient.UpgradeStrategy{
			RollingUpdate: &managementClient.RollingUpdate{
				BatchSize: ctx.Int64(argUpgradeBatchSize),
				Interval:  ctx.Int64(argUpgradeBatchInterval),
			},
		}
	}

	resource, err := Lookup(c, templateName, managementClient.TemplateType)
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
				"version %s for template %s is invalid, run 'rancher mcapp show-template %s' for a list of versions",
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
	answers, answersSetString, err := processAnswerInstall(ctx, templateVersion, nil, nil, interactive, true)
	if err != nil {
		return err
	}

	projectIDs, err := lookupProjectIDsFromTargets(c, ctx.StringSlice("target"))
	if err != nil {
		return err
	}

	for _, target := range projectIDs {
		app.Targets = append(app.Targets, managementClient.Target{
			ProjectID: target,
		})
	}
	if len(projectIDs) == 0 {
		app.Targets = append(app.Targets, managementClient.Target{
			ProjectID: c.UserConfig.Project,
		})
	}

	app.Answers, err = toMultiClusterAppAnswers(c, answers, answersSetString)
	if err != nil {
		return err
	}
	app.TemplateVersionID = templateVersionID

	accessType := strings.ToLower(ctx.String("member-access-type"))
	if !slice.ContainsString(memberAccessTypes, accessType) {
		return fmt.Errorf("invalid access type %q, supported values are \"owner\",\"member\" and \"read-only\"", accessType)
	}

	members, err := addMembersByNames(ctx, c, app.Members, ctx.StringSlice("member"), accessType)
	if err != nil {
		return err
	}
	app.Members = members

	app.Wait = ctx.Bool("helm-wait")
	app.Timeout = ctx.Int64("helm-timeout")

	app, err = c.ManagementClient.MultiClusterApp.Create(app)
	if err != nil {
		return err
	}

	fmt.Printf("Installing multi-cluster app %q...\n", app.Name)

	return nil
}

func lookupProjectIDsFromTargets(c *cliclient.MasterClient, targets []string) ([]string, error) {
	var projectIDs []string
	for _, target := range targets {
		projectID, err := lookupProjectIDFromProjectScope(c, target)
		if err != nil {
			return nil, err
		}
		projectIDs = append(projectIDs, projectID)
	}
	return projectIDs, nil
}

func lookupClusterIDFromClusterScope(c *cliclient.MasterClient, clusterNameOrID string) (string, error) {
	clusterResource, err := Lookup(c, clusterNameOrID, managementClient.ClusterType)
	if err != nil {
		return "", err
	}
	return clusterResource.ID, nil
}

func lookupProjectIDFromProjectScope(c *cliclient.MasterClient, scope string) (string, error) {
	cluster, project := parseScope(scope)
	clusterResource, err := Lookup(c, cluster, managementClient.ClusterType)
	if err != nil {
		return "", err
	}
	if clusterResource.ID == cluster {
		// Lookup by ID
		projectResource, err := Lookup(c, scope, managementClient.ProjectType)
		if err != nil {
			return "", err
		}
		return projectResource.ID, nil
	}
	// Lookup by clusterName:projectName
	projectResource, err := Lookup(c, project, managementClient.ProjectType)
	if err != nil {
		return "", err
	}
	return projectResource.ID, nil

}

func toMultiClusterAppAnswers(c *cliclient.MasterClient, answers, answersSetString map[string]string) ([]managementClient.Answer, error) {
	answerMap := make(map[string]scopeAnswers)
	var answerSlice []managementClient.Answer
	if err := setValueInAnswerMapByScope(c, answerMap, answers, "Answers"); err != nil {
		return nil, err
	}
	if err := setValueInAnswerMapByScope(c, answerMap, answersSetString, "AnswersSetString"); err != nil {
		return nil, err
	}
	for k, v := range answerMap {
		answer := managementClient.Answer{
			Values:          v.Answers,
			ValuesSetString: v.AnswersSetString,
		}
		if strings.Contains(k, ":") {
			answer.ProjectID = k
		} else if k != "" {
			answer.ClusterID = k
		}
		answerSlice = append(answerSlice, answer)
	}
	return answerSlice, nil
}

func setValueInAnswerMapByScope(c *cliclient.MasterClient, answerMap map[string]scopeAnswers, inputAnswers map[string]string, scopeAnswersFieldStr string) error {
	for k, v := range inputAnswers {
		switch parts := strings.SplitN(k, ":", 3); {
		case len(parts) == 1:
			// Global scope
			setValueInAnswerMap(answerMap, "", "", scopeAnswersFieldStr, k, v)
		case len(parts) == 2:
			// Cluster scope
			clusterNameOrID := parts[0]
			clusterID, err := lookupClusterIDFromClusterScope(c, clusterNameOrID)
			if err != nil {
				return err
			}
			setValueInAnswerMap(answerMap, clusterNameOrID, clusterID, scopeAnswersFieldStr, parts[1], v)
		case len(parts) == 3:
			// Project scope
			projectScope := concatScope(parts[0], parts[1])
			projectID, err := lookupProjectIDFromProjectScope(c, projectScope)
			if err != nil {
				return err
			}
			setValueInAnswerMap(answerMap, projectScope, projectID, scopeAnswersFieldStr, parts[2], v)
		}
	}
	return nil
}

func setValueInAnswerMap(answerMap map[string]scopeAnswers, scope, scopeID, fieldNameToUpdate, key, value string) {
	var exist bool
	if answerMap[scopeID].Answers == nil && answerMap[scopeID].AnswersSetString == nil {
		answerMap[scopeID] = scopeAnswers{
			Answers:          make(map[string]string),
			AnswersSetString: make(map[string]string),
		}
	}
	scopeAnswersStruct := answerMap[scopeID]
	scopeAnswersMap := reflect.ValueOf(&scopeAnswersStruct).Elem().FieldByName(fieldNameToUpdate)
	for _, k := range scopeAnswersMap.MapKeys() {
		if reflect.ValueOf(key) == k {
			exist = true
			break
		}
	}
	if exist {
		// It is possible that there are different forms of the same answer key in aggregated answers
		// In this case, name format from users overrides id format from existing app answers.
		if scope != scopeID {
			scopeAnswersMap.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
		}
	} else {
		scopeAnswersMap.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
	}
}

func fromMultiClusterAppAnswers(answerSlice []managementClient.Answer) (map[string]string, map[string]string) {
	answers := make(map[string]string)
	answersSetString := make(map[string]string)
	for _, answer := range answerSlice {
		for k, v := range answer.Values {
			scopedKey := getAnswerScopedKey(answer, k)
			answers[scopedKey] = v
		}
		for k, v := range answer.ValuesSetString {
			scopedKey := getAnswerScopedKey(answer, k)
			answersSetString[scopedKey] = v
		}
	}
	return answers, answersSetString
}

func getAnswerScopedKey(answer managementClient.Answer, key string) string {
	scope := ""
	if answer.ProjectID != "" {
		scope = answer.ProjectID
	} else if answer.ClusterID != "" {
		scope = answer.ClusterID
	}
	scopedKey := key
	if scope != "" {
		scopedKey = concatScope(scope, key)
	}
	return scopedKey
}

func addMcappTargetProject(ctx *cli.Context) error {
	if len(ctx.Args()) < 2 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	_, app, err := searchForMcapp(c, ctx.Args().First())
	if err != nil {
		return err
	}

	input, err := getTargetInput(ctx, c)
	if err != nil {
		return err
	}

	if err := c.ManagementClient.MultiClusterApp.ActionAddProjects(app, input); err != nil {
		return err
	}

	return nil
}

func deleteMcappTargetProject(ctx *cli.Context) error {
	if len(ctx.Args()) < 2 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	_, app, err := searchForMcapp(c, ctx.Args().First())
	if err != nil {
		return err
	}

	input, err := getTargetInput(ctx, c)
	if err != nil {
		return err
	}
	return c.ManagementClient.MultiClusterApp.ActionRemoveProjects(app, input)
}

func getTargetInput(ctx *cli.Context, c *cliclient.MasterClient) (*managementClient.UpdateMultiClusterAppTargetsInput, error) {
	targets := ctx.Args()[1:]
	projectIDs, err := lookupProjectIDsFromTargets(c, targets)
	if err != nil {
		return nil, err
	}
	answers, answersSetString, err := processAnswerUpdates(ctx, nil, nil)
	if err != nil {
		return nil, err
	}
	mcaAnswers, err := toMultiClusterAppAnswers(c, answers, answersSetString)
	if err != nil {
		return nil, err
	}
	input := &managementClient.UpdateMultiClusterAppTargetsInput{
		Projects: projectIDs,
		Answers:  mcaAnswers,
	}
	return input, nil
}

func addMcappMember(ctx *cli.Context) error {
	if len(ctx.Args()) < 3 {
		return cli.ShowSubcommandHelp(ctx)
	}

	appName := ctx.Args().First()
	accessType := strings.ToLower(ctx.Args().Get(1))
	memberNames := ctx.Args()[2:]

	if !slice.ContainsString(memberAccessTypes, accessType) {
		return fmt.Errorf("invalid access type %q, supported values are \"owner\",\"member\" and \"read-only\"", accessType)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	_, app, err := searchForMcapp(c, appName)
	if err != nil {
		return err
	}

	members, err := addMembersByNames(ctx, c, app.Members, memberNames, accessType)
	if err != nil {
		return err
	}

	update := make(map[string]interface{})
	update["members"] = members
	update["roles"] = app.Roles

	_, err = c.ManagementClient.MultiClusterApp.Update(app, update)
	return err
}

func deleteMcappMember(ctx *cli.Context) error {
	if len(ctx.Args()) < 2 {
		return cli.ShowSubcommandHelp(ctx)
	}

	appName := ctx.Args().First()
	memberNames := ctx.Args()[1:]

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	_, app, err := searchForMcapp(c, appName)
	if err != nil {
		return err
	}

	members, err := deleteMembersByNames(ctx, c, app.Members, memberNames)
	if err != nil {
		return err
	}

	update := make(map[string]interface{})
	update["members"] = members
	update["roles"] = app.Roles

	_, err = c.ManagementClient.MultiClusterApp.Update(app, update)
	return err
}

func showMultiClusterApp(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	resource, app, err := searchForMcapp(c, ctx.Args().First())
	if err != nil {
		return err
	}

	err = outputMultiClusterAppRevisions(ctx, c, resource, app)
	if err != nil {
		return err
	}

	fmt.Println()

	err = outputMultiClusterAppVersions(ctx, c, app)
	if err != nil {
		return err
	}

	if ctx.Bool("show-roles") {
		fmt.Println()

		err = outputMultiClusterAppRoles(ctx, c, app)
		if err != nil {
			return err
		}
	}

	return nil
}

func listMultiClusterAppMembers(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	_, app, err := searchForMcapp(c, ctx.Args().First())
	if err != nil {
		return err
	}

	return outputMembers(ctx, c, app.Members)
}

func listMultiClusterAppAnswers(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	_, app, err := searchForMcapp(c, ctx.Args().First())
	if err != nil {
		return err
	}

	return outputMultiClusterAppAnswers(ctx, c, app)
}

func searchForMcapp(c *cliclient.MasterClient, name string) (*types.Resource, *managementClient.MultiClusterApp, error) {
	resource, err := Lookup(c, name, managementClient.MultiClusterAppType)
	if err != nil {
		return nil, nil, err
	}
	app, err := c.ManagementClient.MultiClusterApp.ByID(resource.ID)
	if err != nil {
		return nil, nil, err
	}
	return resource, app, nil
}

func outputMultiClusterAppVersions(ctx *cli.Context, c *cliclient.MasterClient, app *managementClient.MultiClusterApp) error {
	templateVersion, err := c.ManagementClient.TemplateVersion.ByID(app.TemplateVersionID)
	if err != nil {
		return err
	}

	ver, err := getRancherServerVersion(c)
	if err != nil {
		return err
	}

	filter := defaultListOpts(ctx)
	filter.Filters["rancherVersion"] = ver

	template := &managementClient.Template{}
	if err := c.ManagementClient.Ops.DoGet(templateVersion.Links["template"], filter, template); err != nil {
		return err
	}
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
		if version.String() == templateVersion.Version {
			current = "*"
		}
		writer.Write(&VersionData{
			Current: current,
			Version: version.String(),
		})
	}
	return writer.Err()
}

func outputMultiClusterAppRevisions(ctx *cli.Context, c *cliclient.MasterClient, resource *types.Resource, app *managementClient.MultiClusterApp) error {
	revisions := &managementClient.MultiClusterAppRevisionCollection{}
	if err := c.ManagementClient.GetLink(*resource, "revisions", revisions); err != nil {
		return err
	}

	var sorted revSlice
	for _, rev := range revisions.Data {
		parsedTime, err := time.Parse(time.RFC3339, rev.Created)
		if err != nil {
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
		if rev.Name == app.Status.RevisionID {
			rev.Current = "*"
		}
		rev.Human = rev.Created.Format("02 Jan 2006 15:04:05 MST")
		writer.Write(rev)

	}
	return writer.Err()
}

func outputMultiClusterAppRoles(ctx *cli.Context, c *cliclient.MasterClient, app *managementClient.MultiClusterApp) error {
	writer := NewTableWriter([][]string{
		{"ROLE_NAME", "Name"},
	}, ctx)

	defer writer.Close()

	for _, r := range app.Roles {
		writer.Write(map[string]string{"Name": r})
	}
	return writer.Err()
}

func outputMultiClusterAppAnswers(ctx *cli.Context, c *cliclient.MasterClient, app *managementClient.MultiClusterApp) error {
	writer := NewTableWriter([][]string{
		{"SCOPE", "Scope"},
		{"QUESTION", "Question"},
		{"ANSWER", "Answer"},
	}, ctx)

	defer writer.Close()

	answers := app.Answers
	// Sort answers by scope in the Global-Cluster-Project order
	sort.Slice(answers, func(i, j int) bool {
		if answers[i].ClusterID == "" && answers[i].ProjectID == "" {
			return true
		} else if answers[i].ClusterID != "" && answers[j].ProjectID != "" {
			return true
		}
		return false
	})

	var scope string
	for _, r := range answers {
		scope = "Global"
		if r.ClusterID != "" {
			cluster, err := getClusterByID(c, r.ClusterID)
			if err != nil {
				return err
			}
			scope = fmt.Sprintf("All projects in cluster %s", cluster.Name)
		} else if r.ProjectID != "" {
			project, err := getProjectByID(c, r.ProjectID)
			if err != nil {
				return err
			}
			scope = fmt.Sprintf("Project %s", project.Name)
		}
		for key, value := range r.Values {
			writer.Write(map[string]string{
				"Scope":    scope,
				"Question": key,
				"Answer":   value,
			})
		}
		for key, value := range r.ValuesSetString {
			writer.Write(map[string]string{
				"Scope":    scope,
				"Question": key,
				"Answer":   fmt.Sprintf("\"%s\"", value),
			})
		}
	}
	return writer.Err()
}

func globalTemplateLs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	filter := defaultListOpts(ctx)
	if ctx.String("catalog") != "" {
		resource, err := Lookup(c, ctx.String("catalog"), managementClient.CatalogType)
		if err != nil {
			return err
		}
		filter.Filters["catalogId"] = resource.ID
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
		// Skip non-global catalogs
		if item.CatalogID == "" {
			continue
		}
		writer.Write(&TemplateData{
			ID:       item.ID,
			Template: item,
			Category: strings.Join(item.Categories, ","),
		})
	}

	return writer.Err()
}

func concatScope(scope, key string) string {
	return fmt.Sprintf("%s:%s", scope, key)
}

func parseScope(ref string) (scope string, key string) {
	parts := strings.SplitN(ref, ":", 2)
	if len(parts) == 1 {
		return "", parts[0]
	}
	return parts[0], parts[1]
}

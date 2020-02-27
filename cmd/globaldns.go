package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rancher/cli/cliclient"
	managementClient "github.com/rancher/types/client/management/v3"
	"github.com/urfave/cli"
)

const (
	alidnsProvider     = "alidns"
	cloudflareProvider = "cloudflare"
	route53Provider    = "route53"

	envAlibabaCloudAccessKey = "ALI_ACCESS_KEY_ID"
	envAlibabaCloudSecretKey = "ALI_ACCESS_KEY_SECRET"
	envAwsAccessKey          = "AWS_ACCESS_KEY_ID"
	envAwsSecretKey          = "AWS_SECRET_ACCESS_KEY"
	envCloudflareAPIEmail    = "CF_API_EMAIL"
	envCloudflareAPIKey      = "CF_API_KEY"

	argAlibabaCloudAccessKey = "alibabacloud-access-key-id"
	argAlibabaCloudSecretKey = "alibabacloud-access-key-secret"
	argAwsAccessKey          = "aws-access-key"
	argAwsSecretKey          = "aws-secret-key"
	argCloudflareAPIEmail    = "cloudflare-api-email"
	argCloudflareAPIKey      = "cloudflare-api-key"

	argFQDN            = "fqdn"
	argMember          = "member"
	argMultiClusterApp = "multi-cluster-app"
	argProject         = "project"
	argProvider        = "provider"
	argRootDomain      = "root-domain"
	argTTL             = "ttl"
	argType            = "type"

	memberAccessTypeOwner = "owner"
)

type globalDNSHolder struct {
	ID       string
	Provider managementClient.GlobalDNSProvider
	Entry    managementClient.GlobalDNS
	Target   string
}

var globalDNSProviderCredentialFlags = []cli.Flag{
	cli.StringFlag{
		Name:   argAlibabaCloudAccessKey,
		EnvVar: envAlibabaCloudAccessKey,
		Usage:  "Alibaba Cloud access key ID for alidns provider",
	},
	cli.StringFlag{
		Name:   argAlibabaCloudSecretKey,
		EnvVar: envAlibabaCloudSecretKey,
		Usage:  "Alibaba Cloud access key secret for alidns provider",
	},
	cli.StringFlag{
		Name:   argAwsAccessKey,
		EnvVar: envAwsAccessKey,
		Usage:  "AWS access key for route53 provider",
	},
	cli.StringFlag{
		Name:   argAwsSecretKey,
		EnvVar: envAwsSecretKey,
		Usage:  "AWS secret key for route53 provider",
	},
	cli.StringFlag{
		Name:   argCloudflareAPIEmail,
		EnvVar: envCloudflareAPIEmail,
		Usage:  "API email for Cloudflare provider",
	},
	cli.StringFlag{
		Name:   argCloudflareAPIKey,
		EnvVar: envCloudflareAPIKey,
		Usage:  "API key for Cloudflare provider",
	},
}

func GlobalDNSCommand() cli.Command {
	return cli.Command{
		Name:  "globaldns",
		Usage: "Operations on global DNS providers and entries",
		Subcommands: []cli.Command{
			{
				Name:    "providers",
				Aliases: []string{"provider"},
				Usage:   "Operations on global DNS providers",
				Action:  defaultAction(globalDNSProviderLs),
				Subcommands: []cli.Command{
					{
						Name:   "ls",
						Usage:  "List global DNS providers",
						Action: globalDNSProviderLs,
						Flags: []cli.Flag{
							formatFlag,
							quietFlag,
						},
					},
					{
						Name:      "create",
						Usage:     "Create a global DNS provider",
						Action:    globalDNSProviderCreate,
						ArgsUsage: "[NAME]",
						Flags: append([]cli.Flag{
							cli.StringFlag{
								Name:  argType,
								Usage: "Global DNS provider type, available options are \"alidns\", \"cloudflare\" and \"route53\"",
							},
							cli.StringFlag{
								Name:  argRootDomain,
								Usage: "Set root domain of a global DNS provider",
							},
							cli.StringSliceFlag{
								Name:  argMember,
								Usage: "Set members of the global DNS provider, can be used multiple times",
							},
						}, globalDNSProviderCredentialFlags...),
					},
					{
						Name:      "update",
						Usage:     "Update a global DNS provider",
						Action:    globalDNSProviderUpdate,
						ArgsUsage: "[PROVIDER_NAME/PROVIDER_ID]",
						Flags: append([]cli.Flag{
							cli.StringFlag{
								Name:  argRootDomain,
								Usage: "Set root domain of a global DNS provider",
							},
						}, globalDNSProviderCredentialFlags...),
					},
					{
						Name:      "delete",
						Aliases:   []string{"rm"},
						Usage:     "Delete a global DNS provider",
						Action:    globalDNSProviderDelete,
						ArgsUsage: "[PROVIDER_NAME/PROVIDER_ID...]",
					},
					{
						Name:      "list-members",
						Aliases:   []string{"lm"},
						Usage:     "List members of a global DNS provider",
						Action:    listGlobalDNSProviderMembers,
						ArgsUsage: "[PROVIDER_NAME/PROVIDER_ID]",
						Flags: []cli.Flag{
							formatFlag,
						},
					},
					{
						Name:      "add-member",
						Usage:     "Add members to a global DNS provider",
						Action:    addGlobalDNSProviderMembers,
						ArgsUsage: "[PROVIDER_NAME/PROVIDER_ID, USER_NAME/USER_ID...]",
					},
					{
						Name:      "delete-member",
						Usage:     "Delete members from a global DNS provider",
						Action:    deleteGlobalDNSProviderMembers,
						ArgsUsage: "[PROVIDER_NAME/PROVIDER_ID, USER_NAME/USER_ID...]",
					},
				},
			},
			{
				Name:    "entries",
				Aliases: []string{"entry"},
				Usage:   "Operations on global DNS entries",
				Action:  defaultAction(globalDNSLs),
				Subcommands: []cli.Command{
					{
						Name:   "ls",
						Usage:  "List global DNS entries",
						Action: globalDNSLs,
						Flags: []cli.Flag{
							formatFlag,
							quietFlag,
						},
					},
					{
						Name:   "create",
						Usage:  "Create a global DNS entry",
						Action: globalDNSCreate,
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:  argFQDN,
								Usage: "FQDN of a global DNS entry",
							},
							cli.Int64Flag{
								Name:  argTTL,
								Usage: "DNS TTL in seconds",
								Value: 300,
							},
							cli.StringFlag{
								Name:  argProvider,
								Usage: "Global DNS provider for an entry. Run \"rancher globaldns provider ls\" to see available ones",
							},
							cli.StringFlag{
								Name:  argMultiClusterApp,
								Usage: "Set a multi-cluster app as the target to which a global DNS entry resolves",
							},
							cli.StringSliceFlag{
								Name:  argProject,
								Usage: "Set projects as the target to which a global DNS entry resolves, can be used multiple times",
							},
							cli.StringSliceFlag{
								Name:  argMember,
								Usage: "Set members of a global DNS entry, can be used multiple times",
							},
						},
					},
					{
						Name:      "update",
						Usage:     "Update a global DNS entry",
						Action:    globalDNSUpdate,
						ArgsUsage: "[ENTRY_ID]",
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:  argFQDN,
								Usage: "FQDN of a global DNS entry",
							},
							cli.Int64Flag{
								Name:  argTTL,
								Usage: "DNS TTL in seconds",
								Value: 300,
							},
							cli.StringFlag{
								Name:  argProvider,
								Usage: "Global DNS provider for an entry. Run \"rancher globaldns provider ls\" to see available ones",
							},
							cli.StringFlag{
								Name:  argMultiClusterApp,
								Usage: "Set a multi-cluster app as the target to which a global DNS entry resolves",
							},
						},
					},
					{
						Name:      "delete",
						Aliases:   []string{"rm"},
						Usage:     "Delete global DNS entries",
						Action:    globalDNSDelete,
						ArgsUsage: "[ENTRY_ID...]",
					},
					{
						Name:      "list-members",
						Aliases:   []string{"lm"},
						Usage:     "List members of a global DNS entry",
						Action:    listGlobalDNSMembers,
						ArgsUsage: "[ENTRY_ID]",
						Flags: []cli.Flag{
							formatFlag,
						},
					},
					{
						Name:      "add-member",
						Usage:     "Add members to a global DNS entry",
						Action:    addGlobalDNSMembers,
						ArgsUsage: "[ENTRY_ID, USER_NAME/USER_ID...]",
					},
					{
						Name:      "delete-member",
						Usage:     "Delete members from a global DNS entry",
						Action:    deleteGlobalDNSMembers,
						ArgsUsage: "[ENTRY_ID, USER_NAME/USER_ID...]",
					},
					{
						Name:  "add-project",
						Usage: "Add target projects to a global DNS entry",
						Description: "If the global DNS entry uses a multi-cluster app as the target, it will remove " +
							"the multi-cluster app target and take the newly added projects as the target",
						Action:    addGlobalDNSProjects,
						ArgsUsage: "[ENTRY_ID, PROJECT_NAME/PROJECT_ID...]",
					},
					{
						Name:      "delete-project",
						Usage:     "Delete target projects from a global DNS entry",
						Action:    deleteGlobalDNSProjects,
						ArgsUsage: "[ENTRY_ID, PROJECT_NAME/PROJECT_ID...]",
					},
				},
			},
		},
	}
}

func globalDNSProviderLs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	providers, err := c.ManagementClient.GlobalDNSProvider.List(defaultListOpts(ctx))
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "Provider.Name"},
		{"ROOT_DOMAIN", "Provider.RootDomain"},
		{"CREATED", "Provider.Created"},
	}, ctx)

	defer writer.Close()

	for _, provider := range providers.Data {
		writer.Write(&globalDNSHolder{
			ID:       provider.ID,
			Provider: provider,
		})
	}
	return writer.Err()
}

func listGlobalDNSProviderMembers(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	provider, err := searchForGlobalDNSProvider(c, ctx.Args().First())
	if err != nil {
		return err
	}

	return outputMembers(ctx, c, provider.Members)
}

func globalDNSProviderCreate(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	name := ctx.Args().First()
	providerType := strings.ToLower(ctx.String(argType))
	rootDomain := ctx.String(argRootDomain)

	if providerType == "" {
		return errors.New("--type is required")
	}

	provider := &managementClient.GlobalDNSProvider{
		Name:       name,
		RootDomain: rootDomain,
	}
	switch providerType {
	case route53Provider:
		accessKey := ctx.String(argAwsAccessKey)
		if accessKey == "" {
			return errors.New("AWS access key is required for route53 type")
		}
		secretKey := ctx.String(argAwsSecretKey)
		if secretKey == "" {
			return errors.New("AWS secret key is required for route53 type")
		}
		provider.Route53ProviderConfig = &managementClient.Route53ProviderConfig{
			AccessKey: accessKey,
			SecretKey: secretKey,
		}
	case cloudflareProvider:
		apiEmail := ctx.String(argCloudflareAPIEmail)
		if apiEmail == "" {
			return errors.New("API email is required for cloudflare type")
		}
		apiKey := ctx.String(argCloudflareAPIKey)
		if apiKey == "" {
			return errors.New("API key is required for cloudflare type")
		}
		provider.CloudflareProviderConfig = &managementClient.CloudflareProviderConfig{
			APIEmail: apiEmail,
			APIKey:   apiKey,
		}
	case alidnsProvider:
		accessKey := ctx.String(argAlibabaCloudAccessKey)
		if accessKey == "" {
			return errors.New("Alibaba Cloud access key ID is required for alidns type")
		}
		secretKey := ctx.String(argAlibabaCloudSecretKey)
		if secretKey == "" {
			return errors.New("Alibaba Cloud access key secret is required for alidns type")
		}
		provider.AlidnsProviderConfig = &managementClient.AlidnsProviderConfig{
			AccessKey: accessKey,
			SecretKey: secretKey,
		}
	default:
		return fmt.Errorf("unsupported provider type %q", providerType)
	}

	members, err := addMembersByNames(ctx, c, provider.Members, ctx.StringSlice(argMember), memberAccessTypeOwner)
	if err != nil {
		return err
	}
	provider.Members = members

	if _, err = c.ManagementClient.GlobalDNSProvider.Create(provider); nil != err {
		return err
	}

	fmt.Printf("Successfully created global DNS provider %q\n", name)
	return nil
}

func globalDNSProviderUpdate(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	name := ctx.Args().First()

	provider, err := searchForGlobalDNSProvider(c, name)
	if err != nil {
		return err
	}

	update := make(map[string]interface{})

	if ctx.IsSet(argRootDomain) {
		update["rootDomain"] = ctx.String(argRootDomain)
	}

	if provider.Route53ProviderConfig != nil {
		accessKey := ctx.String(argAwsAccessKey)
		if accessKey != "" {
			provider.Route53ProviderConfig.AccessKey = accessKey
		}

		secretKey := ctx.String(argAwsSecretKey)
		if secretKey != "" {
			provider.Route53ProviderConfig.SecretKey = secretKey
		}

		update["route53ProviderConfig"] = provider.Route53ProviderConfig
	} else if provider.CloudflareProviderConfig != nil {
		apiEmail := ctx.String(argCloudflareAPIEmail)
		if apiEmail != "" {
			provider.CloudflareProviderConfig.APIEmail = apiEmail
		}

		apiKey := ctx.String(argCloudflareAPIKey)
		if apiKey != "" {
			provider.CloudflareProviderConfig.APIKey = apiKey
		}

		update["cloudflareProviderConfig"] = provider.CloudflareProviderConfig
	} else if provider.AlidnsProviderConfig != nil {
		accessKey := ctx.String(argAlibabaCloudAccessKey)
		if accessKey != "" {
			provider.AlidnsProviderConfig.AccessKey = accessKey
		}

		secretKey := ctx.String(argAlibabaCloudSecretKey)
		if secretKey != "" {
			provider.AlidnsProviderConfig.SecretKey = secretKey
		}

		update["alidnsProviderConfig"] = provider.AlidnsProviderConfig
	} else {
		return fmt.Errorf("unsupported provider type for %q", name)
	}

	_, err = c.ManagementClient.GlobalDNSProvider.Update(provider, update)
	return err
}

func globalDNSProviderDelete(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	for _, name := range ctx.Args() {
		provider, err := searchForGlobalDNSProvider(c, name)
		if err != nil {
			return err
		}

		if err = c.ManagementClient.GlobalDNSProvider.Delete(provider); nil != err {
			return err
		}
	}

	return nil
}

func addGlobalDNSProviderMembers(ctx *cli.Context) error {
	if ctx.NArg() < 2 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	provider, err := searchForGlobalDNSProvider(c, ctx.Args().First())
	if err != nil {
		return err
	}

	members, err := addMembersByNames(ctx, c, provider.Members, ctx.Args()[1:], memberAccessTypeOwner)
	if err != nil {
		return err
	}

	update := make(map[string][]managementClient.Member)
	update["members"] = members

	_, err = c.ManagementClient.GlobalDNSProvider.Update(provider, update)
	return err
}

func deleteGlobalDNSProviderMembers(ctx *cli.Context) error {
	if ctx.NArg() < 2 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	provider, err := searchForGlobalDNSProvider(c, ctx.Args().First())
	if err != nil {
		return err
	}

	members, err := deleteMembersByNames(ctx, c, provider.Members, ctx.Args()[1:])
	if err != nil {
		return err
	}

	update := make(map[string]interface{})
	update["members"] = members

	_, err = c.ManagementClient.GlobalDNSProvider.Update(provider, update)
	return err
}

func globalDNSLs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	entries, err := c.ManagementClient.GlobalDNS.List(defaultListOpts(ctx))
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"FQDN", "Entry.FQDN"},
		{"PROVIDER_ID", "Entry.ProviderID"},
		{"TARGET", "Target"},
		{"CREATED", "Entry.Created"},
	}, ctx)

	defer writer.Close()

	clusterCache, projectCache, err := getClusterProjectMap(ctx, c.ManagementClient)
	if err != nil {
		return err
	}

	for _, entry := range entries.Data {
		target, err := getEntryTarget(c, clusterCache, projectCache, &entry)
		if err != nil {
			return err
		}
		writer.Write(&globalDNSHolder{
			ID:     entry.ID,
			Entry:  entry,
			Target: target,
		})
	}
	return writer.Err()
}

func getEntryTarget(c *cliclient.MasterClient, clusterCache map[string]managementClient.Cluster, projectCache map[string]managementClient.Project, entry *managementClient.GlobalDNS) (string, error) {
	var target string
	if entry.MultiClusterAppID != "" {
		_, app, err := searchForMcapp(c, entry.MultiClusterAppID)
		if err != nil {
			return "", err
		}
		target = fmt.Sprintf("Multi Cluster App: %s", app.Name)
	} else {
		var targets []managementClient.Target
		for _, projectID := range entry.ProjectIDs {
			targets = append(targets, managementClient.Target{
				ProjectID: projectID,
			})
		}
		targetNames := getReadableTargetNames(clusterCache, projectCache, targets)
		target = strings.Join(targetNames, ",")
	}
	return target, nil
}

func listGlobalDNSMembers(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	entry, err := searchForGlobalDNS(c, ctx.Args().First())
	if err != nil {
		return err
	}

	return outputMembers(ctx, c, entry.Members)
}

func globalDNSCreate(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	fqdn := ctx.String(argFQDN)
	provider := ctx.String(argProvider)
	appName := ctx.String(argMultiClusterApp)
	projects := ctx.StringSlice(argProject)
	ttl := ctx.Int64(argTTL)

	if fqdn == "" {
		return errors.New("--fqdn is required")
	}

	if provider == "" {
		return errors.New("--provider is required")
	}

	if appName != "" && len(projects) > 0 || appName == "" && len(projects) == 0 {
		return fmt.Errorf("please specify either --multi-cluster-app or --project as the global DNS entry target")
	}

	entry := &managementClient.GlobalDNS{
		FQDN: fqdn,
		TTL:  ttl,
	}

	globalDNSProvider, err := searchForGlobalDNSProvider(c, provider)
	if err != nil {
		return err
	}
	entry.ProviderID = globalDNSProvider.ID

	members, err := addMembersByNames(ctx, c, entry.Members, ctx.StringSlice(argMember), memberAccessTypeOwner)
	if err != nil {
		return err
	}
	entry.Members = members

	if appName != "" {
		resource, err := Lookup(c, appName, managementClient.MultiClusterAppType)
		if err != nil {
			return err
		}
		entry.MultiClusterAppID = resource.ID
	} else {
		projectIDs, err := lookupProjectIDsFromTargets(c, projects)
		if err != nil {
			return err
		}
		entry.ProjectIDs = projectIDs
	}

	if _, err = c.ManagementClient.GlobalDNS.Create(entry); nil != err {
		return err
	}

	fmt.Printf("Successfully created global DNS entry for %q\n", fqdn)
	return nil
}

func globalDNSUpdate(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	entry, err := searchForGlobalDNS(c, ctx.Args().First())
	if err != nil {
		return err
	}

	update := make(map[string]interface{})

	if ctx.IsSet(argFQDN) {
		update["fqdn"] = ctx.String(argFQDN)
	}

	if ctx.IsSet(argProvider) {
		globalDNSProvider, err := searchForGlobalDNSProvider(c, ctx.String(argProvider))
		if err != nil {
			return err
		}
		update["providerId"] = globalDNSProvider.ID
	}

	if ctx.IsSet(argMultiClusterApp) {
		resource, err := Lookup(c, ctx.String(argMultiClusterApp), managementClient.MultiClusterAppType)
		if err != nil {
			return err
		}
		update["multiClusterAppId"] = resource.ID
		if len(entry.ProjectIDs) > 0 {
			input := &managementClient.UpdateGlobalDNSTargetsInput{
				ProjectIDs: entry.ProjectIDs,
			}
			if err := c.ManagementClient.GlobalDNS.ActionRemoveProjects(entry, input); err != nil {
				return err
			}
		}
	}

	if ctx.IsSet(argTTL) {
		update["ttl"] = ctx.Int64(argTTL)
	}

	if _, err := c.ManagementClient.GlobalDNS.Update(entry, update); err != nil {
		// Rollback target projects on update failure
		if ctx.IsSet(argMultiClusterApp) && len(entry.ProjectIDs) > 0 {
			input := &managementClient.UpdateGlobalDNSTargetsInput{
				ProjectIDs: entry.ProjectIDs,
			}
			if rbErr := c.ManagementClient.GlobalDNS.ActionAddProjects(entry, input); rbErr != nil {
				return fmt.Errorf("failed to update global DNS entry: %v and failed to rollback to previous target projects: %v", err, rbErr)
			}
		}
		return err
	}
	return nil
}

func globalDNSDelete(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	for _, name := range ctx.Args() {
		entry, err := searchForGlobalDNS(c, name)
		if err != nil {
			return err
		}

		if err = c.ManagementClient.GlobalDNS.Delete(entry); nil != err {
			return err
		}
	}

	return nil
}

func addGlobalDNSMembers(ctx *cli.Context) error {
	if ctx.NArg() < 2 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	entry, err := searchForGlobalDNS(c, ctx.Args().First())
	if err != nil {
		return err
	}
	members, err := addMembersByNames(ctx, c, entry.Members, ctx.Args()[1:], memberAccessTypeOwner)
	if err != nil {
		return err
	}

	update := make(map[string]interface{})
	update["members"] = members

	_, err = c.ManagementClient.GlobalDNS.Update(entry, update)
	return err
}

func deleteGlobalDNSMembers(ctx *cli.Context) error {
	if ctx.NArg() < 2 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	entry, err := searchForGlobalDNS(c, ctx.Args().First())
	if err != nil {
		return err
	}

	members, err := deleteMembersByNames(ctx, c, entry.Members, ctx.Args()[1:])
	if err != nil {
		return err
	}

	update := make(map[string]interface{})
	update["members"] = members

	_, err = c.ManagementClient.GlobalDNS.Update(entry, update)
	return err
}

func deleteGlobalDNSProjects(ctx *cli.Context) error {
	if len(ctx.Args()) < 2 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	entry, err := searchForGlobalDNS(c, ctx.Args().First())
	if err != nil {
		return err
	}

	targets := ctx.Args()[1:]
	projectIDs, err := lookupProjectIDsFromTargets(c, targets)
	if err != nil {
		return err
	}

	input := &managementClient.UpdateGlobalDNSTargetsInput{
		ProjectIDs: projectIDs,
	}

	return c.ManagementClient.GlobalDNS.ActionRemoveProjects(entry, input)
}

func addGlobalDNSProjects(ctx *cli.Context) error {
	if len(ctx.Args()) < 2 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	entry, err := searchForGlobalDNS(c, ctx.Args().First())
	if err != nil {
		return err
	}

	targets := ctx.Args()[1:]
	projectIDs, err := lookupProjectIDsFromTargets(c, targets)
	if err != nil {
		return err
	}

	update := make(map[string]interface{})
	if entry.MultiClusterAppID != "" {
		update["multiClusterAppId"] = ""
		if _, err := c.ManagementClient.GlobalDNS.Update(entry, update); err != nil {
			return err
		}
	}

	input := &managementClient.UpdateGlobalDNSTargetsInput{
		ProjectIDs: projectIDs,
	}

	if err := c.ManagementClient.GlobalDNS.ActionAddProjects(entry, input); err != nil {
		// Rollback target mcapp on add-project failure
		if entry.MultiClusterAppID != "" {
			update["multiClusterAppId"] = entry.MultiClusterAppID
			if _, rbErr := c.ManagementClient.GlobalDNS.Update(entry, update); rbErr != nil {
				return fmt.Errorf("failed to add target projects to the entry: %v and failed to rollback to previous multi-cluster app target: %v", err, rbErr)
			}
		}
		return err
	}
	return nil
}

func searchForGlobalDNSProvider(c *cliclient.MasterClient, name string) (*managementClient.GlobalDNSProvider, error) {
	resource, err := Lookup(c, name, managementClient.GlobalDNSProviderType)
	if err != nil {
		return nil, err
	}

	return c.ManagementClient.GlobalDNSProvider.ByID(resource.ID)
}

func searchForGlobalDNS(c *cliclient.MasterClient, name string) (*managementClient.GlobalDNS, error) {
	resource, err := Lookup(c, name, managementClient.GlobalDNSType)
	if err != nil {
		return nil, err
	}

	return c.ManagementClient.GlobalDNS.ByID(resource.ID)
}

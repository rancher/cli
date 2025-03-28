package cmd

import (
	"fmt"

	"github.com/rancher/cli/cliclient"
	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"github.com/urfave/cli"
)

type UserData struct {
	ID   string
	User managementClient.User
}

var availableRoles = map[string]bool{
	"admin":            true,
	"restricted-admin": true,
	"user":             true,
	"user-base":        true,
}

func UserCommand() cli.Command {
	return cli.Command{
		Name:    "users",
		Aliases: []string{"user"},
		Usage:   "Operations on users",
		Action:  defaultAction(userLs),
		Subcommands: []cli.Command{
			{
				Name:        "ls",
				Usage:       "List users",
				Description: "\nLists all users in the current cluster.",
				ArgsUsage:   "None",
				Action:      userLs,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "format",
						Usage: "'json', 'yaml' or Custom format: '{{.User.ID}} {{.User.Name}}'",
					},
					quietFlag,
				},
			},
			{
				Name:        "create",
				Usage:       "Create a user",
				Description: "\nCreates a user in the current cluster.",
				ArgsUsage:   "[NEWUSERNAME...]",
				Action:      userCreate,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "description",
						Usage: "Description to apply to the user",
					},
					cli.StringFlag{
						Name:  "password",
						Usage: "Password for the user",
					},
					cli.StringSliceFlag{
						Name:  "role",
						Usage: "Role of the user (can be specified multiple times).\n\t\tAvailable values are: 'admin', 'restricted-admin', 'user', 'user-base'",
						Value: &cli.StringSlice{
							"user",
						},
					},
				},
			},
			{
				Name:      "delete",
				Aliases:   []string{"rm"},
				Usage:     "Delete a user by ID",
				ArgsUsage: "[USERID]",
				Action:    userDelete,
			},
		},
	}
}

func userLs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	collection, err := getUserList(ctx, c)
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "User.Username"},
		{"STATE", "User.State"},
	}, ctx)

	defer writer.Close()

	for _, user := range collection.Data {
		writer.Write(&UserData{
			ID:   user.ID,
			User: user,
		})
	}

	return writer.Err()
}

func userCreate(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	enabled := true
	username := ctx.Args().First()
	newUser := &managementClient.User{
		Name:        username,
		Username:    username,
		Password:    ctx.String("password"),
		Description: ctx.String("description"),
		// default settings
		Enabled:            &enabled,
		MustChangePassword: false,
	}
	rancherUser, err := c.ManagementClient.User.Create(newUser)
	if err != nil {
		return err
	}

	// setup GlobalRoleBinding(s) to assign user permissions
	for _, role := range ctx.StringSlice("role") {
		if _, ok := availableRoles[role]; !ok {
			return fmt.Errorf("provided role doesn't exists")
		}
		grBinding := &managementClient.GlobalRoleBinding{
			UserID:       rancherUser.ID,
			GlobalRoleID: role,
		}
		_, err = c.ManagementClient.GlobalRoleBinding.Create(grBinding)
		if err != nil {
			return err
		}
	}

	return nil
}

func userDelete(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	for _, arg := range ctx.Args() {
		user, err := c.ManagementClient.User.ByID(arg)
		if err != nil {
			return err
		}
		err = c.ManagementClient.User.Delete(user)
		if err != nil {
			return err
		}
	}

	return nil
}

func getUserList(
	ctx *cli.Context,
	c *cliclient.MasterClient,
) (*managementClient.UserCollection, error) {
	filter := defaultListOpts(ctx)
	filter.Filters["clusterId"] = c.UserConfig.FocusedCluster()

	collection, err := c.ManagementClient.User.List(filter)
	if err != nil {
		return nil, err
	}
	return collection, nil
}

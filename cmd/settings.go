package cmd

import (
	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type settingHolder struct {
	ID      string
	Setting managementClient.Setting
}

func SettingsCommand() cli.Command {
	return cli.Command{
		Name:        "settings",
		Aliases:     []string{"setting"},
		Usage:       "Show settings for the current server",
		Description: "List get or set settings for the current Rancher server",
		Action:      defaultAction(settingsLs),
		Flags: []cli.Flag{
			formatFlag,
		},
		Subcommands: []cli.Command{
			{
				Name:        "ls",
				Usage:       "List settings",
				Description: "Lists all settings in the current cluster.",
				ArgsUsage:   "[SETTINGNAME]",
				Action:      settingsLs,
				Flags: []cli.Flag{
					formatFlag,
					quietFlag,
				},
			},
			{
				Name:   "get",
				Usage:  "Print a setting",
				Action: settingGet,
				Flags: []cli.Flag{
					formatFlag,
				},
			},
			{
				Name:      "set",
				Usage:     "Set the value for a setting",
				Action:    settingSet,
				ArgsUsage: "[SETTINGNAME VALUE]",
				Flags: []cli.Flag{
					formatFlag,
					cli.BoolFlag{
						Name:  "default",
						Usage: "Reset the setting back to it's default value. If the default value is (blank) it will be set to that.",
					},
				},
			},
		},
	}
}

func settingsLs(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	settings, err := c.ManagementClient.Setting.List(defaultListOpts(ctx))
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "Setting.Name"},
		{"VALUE", "Setting.Value"},
	}, ctx)

	defer writer.Close()

	for _, setting := range settings.Data {
		writer.Write(&settingHolder{
			ID:      setting.ID,
			Setting: setting,
		})
	}
	return writer.Err()
}

func settingGet(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowCommandHelp(ctx, "settings")
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	resource, err := Lookup(c, ctx.Args().First(), "setting")
	if err != nil {
		return err
	}

	setting, err := c.ManagementClient.Setting.ByID(resource.ID)
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "Setting.Name"},
		{"VALUE", "Setting.Value"},
		{"DEFAULT", "Setting.Default"},
		{"CUSTOMIZED", "Setting.Customized"},
	}, ctx)

	defer writer.Close()

	writer.Write(&settingHolder{
		ID:      setting.ID,
		Setting: *setting,
	})

	return writer.Err()
}

func settingSet(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.ShowCommandHelp(ctx, "settings")
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	resource, err := Lookup(c, ctx.Args().First(), "setting")
	if err != nil {
		return err
	}

	setting, err := c.ManagementClient.Setting.ByID(resource.ID)
	if err != nil {
		return err
	}

	update := make(map[string]string)
	if ctx.Bool("default") {
		update["value"] = setting.Default
	} else {
		update["value"] = ctx.Args().Get(1)
	}

	updatedSetting, err := c.ManagementClient.Setting.Update(setting, update)
	if err != nil {
		return err
	}

	var updatedValue string
	if updatedSetting.Value == "" {
		updatedValue = "(blank)"
	} else {
		updatedValue = updatedSetting.Value
	}
	logrus.Infof("Successfully updated setting %s with a new value of: %s", updatedSetting.Name, updatedValue)

	return nil
}

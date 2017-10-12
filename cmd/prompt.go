package cmd

import "github.com/urfave/cli"

func PromptCommand() cli.Command {
	return cli.Command{
		Name:      "prompt",
		Usage:     "Enter rancher cli auto-prompt mode",
		ArgsUsage: "None",
		Action:    promptAction,
		Flags:     []cli.Flag{},
	}
}

func promptAction(ctx *cli.Context) error {
	return promptWrapper(ctx)
}

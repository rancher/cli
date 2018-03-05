package cmd

import (
	"fmt"

	"github.com/rancher/cli/rancher_prompt"

	"github.com/c-bata/go-prompt"
	"github.com/urfave/cli"
)

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
	fmt.Print("rancher cli auto-completion mode")
	defer fmt.Println("Goodbye!")
	p := prompt.New(
		rancherprompt.Executor,
		rancherprompt.Completer,
		prompt.OptionTitle("rancher-prompt: interactive rancher client"),
		prompt.OptionPrefix("rancher$ "),
		prompt.OptionInputTextColor(prompt.Yellow),
		prompt.OptionMaxSuggestion(20),
	)
	p.Run()
	return nil
}

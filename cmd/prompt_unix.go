//+build !windows

package cmd

import (
	"fmt"

	"github.com/c-bata/go-prompt"
	"github.com/rancher/cli/rancher_prompt"
	"github.com/urfave/cli"
)

func promptWrapper(ctx *cli.Context) error {
	fmt.Print("rancher cli auto-completion mode")
	defer fmt.Println("Goodbye!")
	p := prompt.New(
		rancherPrompt.Executor,
		rancherPrompt.Completer,
		prompt.OptionTitle("rancher-prompt: interactive rancher client"),
		prompt.OptionPrefix("rancher$ "),
		prompt.OptionInputTextColor(prompt.Yellow),
		prompt.OptionMaxSuggestion(20),
	)
	p.Run()
	return nil
}

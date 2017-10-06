package rancherPrompt

import (
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/urfave/cli"
)

func optionCompleter(args []string, long bool) []prompt.Suggest {
	l := len(args)
	if l <= 1 {
		if long {
			return prompt.FilterHasPrefix(optionHelp, "--", false)
		}
		return optionHelp
	}
	flagGlobal := getGlobalFlag()

	var suggests []prompt.Suggest
	commandArgs := excludeOptions(args)

	if command, ok := Commands[commandArgs[0]]; ok {
		if len(commandArgs) > 1 && len(command.Subcommands) > 0 {
			for _, sub := range command.Subcommands {
				if sub.Name == commandArgs[1] {
					suggests = append(getFlagsSuggests(sub), flagGlobal...)
					break
				}
			}
		} else {
			suggests = append(getFlagsSuggests(command), flagGlobal...)
		}
	}

	if long {
		return prompt.FilterContains(
			prompt.FilterHasPrefix(suggests, "--", false),
			strings.TrimLeft(args[l-1], "--"),
			true,
		)
	}
	return prompt.FilterHasPrefix(suggests, strings.TrimLeft(args[l-1], "-"), true)
}

var optionHelp = []prompt.Suggest{
	{Text: "-h", Description: "Help Commmand"},
	{Text: "--help", Description: "Help Commmand"},
}

func excludeOptions(args []string) []string {
	ret := make([]string, 0, len(args))
	for i := range args {
		if !strings.HasPrefix(args[i], "-") {
			ret = append(ret, args[i])
		}
	}
	return ret
}

func getGlobalFlag() []prompt.Suggest {
	suggests := []prompt.Suggest{}
	for _, flag := range Flags {
		name := flag.GetName()
		parts := strings.Split(name, ",")
		for _, part := range parts {
			prefix := "--"
			if len(part) == 1 {
				prefix = "-"
			}
			suggests = append(suggests, prompt.Suggest{
				Text:        prefix + strings.TrimSpace(part),
				Description: getUsageForFlag(flag),
			})
		}
	}
	suggests = append(suggests, optionHelp...)
	return suggests
}

func getFlagsSuggests(command cli.Command) []prompt.Suggest {
	suggests := []prompt.Suggest{}
	for _, f := range command.Flags {
		name := f.GetName()
		parts := strings.Split(name, ",")
		for _, part := range parts {
			prefix := "--"
			if len(part) == 1 {
				prefix = "-"
			}
			suggests = append(suggests, prompt.Suggest{
				Text:        prefix + strings.TrimSpace(part),
				Description: getUsageForFlag(f),
			})
		}
	}
	return suggests
}

func getUsageForFlag(flag cli.Flag) string {
	if v, ok := flag.(cli.StringFlag); ok {
		return v.Usage
	}
	if v, ok := flag.(cli.StringSliceFlag); ok {
		return v.Usage
	}
	if v, ok := flag.(cli.IntFlag); ok {
		return v.Usage
	}
	if v, ok := flag.(cli.Int64Flag); ok {
		return v.Usage
	}
	if v, ok := flag.(cli.IntSliceFlag); ok {
		return v.Usage
	}
	if v, ok := flag.(cli.BoolFlag); ok {
		return v.Usage
	}
	return ""
}

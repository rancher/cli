// +build !windows

package rancherPrompt

import (
	"strings"

	"github.com/c-bata/go-prompt"
)

// thanks for the idea from github.com/c-bata/kube-prompt

func Completer(d prompt.Document) []prompt.Suggest {
	if d.TextBeforeCursor() == "" {
		return []prompt.Suggest{}
	}

	args := strings.Split(d.TextBeforeCursor(), " ")
	w := d.GetWordBeforeCursor()

	// If PIPE is in text before the cursor, returns empty suggestions.
	for i := range args {
		if args[i] == "|" {
			return []prompt.Suggest{}
		}
	}

	// If word before the cursor starts with "-", returns CLI flag options.
	if strings.HasPrefix(w, "-") {
		return optionCompleter(args, strings.HasPrefix(w, "--"))
	}

	return argumentsCompleter(excludeOptions(args))
}

func argumentsCompleter(args []string) []prompt.Suggest {
	suggests := []prompt.Suggest{}
	for name, command := range Commands {
		if command.Name != "prompt" {
			suggests = append(suggests, prompt.Suggest{
				Text:        name,
				Description: command.Usage,
			})
		}
	}
	if len(args) <= 1 {
		return prompt.FilterHasPrefix(suggests, args[0], true)
	}

	switch args[0] {
	case "docker":
		if len(args) == 3 {
			subcommands := []prompt.Suggest{
				{Text: "attach", Description: "Attach local standard input, output, and error streams to a running container"},
				{Text: "build", Description: "Build an image from a Dockerfile"},
				{Text: "commit", Description: "Create a new image from a container’s changes"},
				{Text: "cp", Description: "Copy files/folders between a container and the local filesystem"},
				{Text: "create", Description: "Create a new container"},
				{Text: "events", Description: "Get real time events from the server"},
				{Text: "exec", Description: "Run a command in a running container"},
				{Text: "export", Description: "Export a container’s filesystem as a tar archive"},
				{Text: "image", Description: "Manage images"},
				{Text: "images", Description: "List images"},
				{Text: "import", Description: "Import the contents from a tarball to create a filesystem image"},
				{Text: "info", Description: "Display system-wide information"},
				{Text: "inspect", Description: "Return low-level information on Docker objects"},
				{Text: "kill", Description: "Kill one or more running containers"},
				{Text: "load", Description: "Load an image from a tar archive or STDIN"},
				{Text: "login", Description: "Log in to a Docker registry"},
				{Text: "logout", Description: "Log out from a Docker registry"},
				{Text: "logs", Description: "Fetch the logs of a container"},
				{Text: "network", Description: "Manage networks"},
				{Text: "pause", Description: "Pause all processes within one or more containers"},
				{Text: "plugin", Description: "Manage plugins"},
				{Text: "port", Description: "List port mappings or a specific mapping for the container"},
				{Text: "ps", Description: "List containers"},
				{Text: "pull", Description: "Pull an image or a repository from a registry"},
				{Text: "push", Description: "Push an image or a repository to a registry"},
				{Text: "rename", Description: "Rename a container"},
				{Text: "restart", Description: "Restart one or more containers"},
				{Text: "rm", Description: "Remove one or more containers"},
				{Text: "rmi", Description: "Remove one or more images"},
				{Text: "run", Description: "Run a command in a new container"},
				{Text: "save", Description: "Save one or more images to a tar archive (streamed to STDOUT by default)"},
				{Text: "search", Description: "Search the Docker Hub for images"},
				{Text: "start", Description: "Start one or more stopped containers"},
				{Text: "stats", Description: "Display a live stream of container(s) resource usage statistics"},
				{Text: "stop", Description: "Stop one or more running containers"},
				{Text: "tag", Description: "Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE"},
				{Text: "top", Description: "Display the running processes of a container"},
				{Text: "unpause", Description: "Unpause all processes within one or more containers"},
				{Text: "update", Description: "Update configuration of one or more containers"},
				{Text: "version", Description: "Show the Docker version information"},
				{Text: "volume", Description: "Manage volumes"},
				{Text: "wait", Description: "Block until one or more containers stop, then print their exit codes"},
			}
			return prompt.FilterHasPrefix(subcommands, args[2], true)
		}
	default:
		if len(args) == 2 {
			return prompt.FilterHasPrefix(getSubcommandSuggest(args[0]), args[1], true)
		}
	}
	return []prompt.Suggest{}
}

func getSubcommandSuggest(name string) []prompt.Suggest {
	subcommands := []prompt.Suggest{}
	for _, com := range Commands[name].Subcommands {
		subcommands = append(subcommands, prompt.Suggest{
			Text:        com.Name,
			Description: com.Usage,
		})
	}
	return subcommands
}

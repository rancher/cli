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

var commands = []prompt.Suggest{
	{Text: "catalog", Description: "Operations with catalogs"},
	{Text: "config", Description: "Setup client configuration"},
	{Text: "docker", Description: "Run docker CLI on a host"},
	{Text: "env", Description: "Interact with environments"},
	{Text: "environment", Description: "Interact with environments"},
	{Text: "events", Description: "Displays resource change events"},
	{Text: "exec", Description: "Run a command on a container"},
	{Text: "export", Description: "Export configuration yml for a stack as a tar archive or to local files"},
	{Text: "host", Description: "Operations on hosts"},
	{Text: "inspect", Description: "Inspect services/containers/stacks/volumes/hosts/secrets"},
	{Text: "logs", Description: "Fetch the logs of a container"},
	{Text: "ps", Description: "Show services/containers"},
	{Text: "restart", Description: "Restart services/containers in rancher"},
	{Text: "rm", Description: "Delete services/containers/stacks/volumes/hosts/secrets"},
	{Text: "run", Description: "Run services/containers"},
	{Text: "scale", Description: "Set number of containers to run for a service"},
	{Text: "secret", Description: "Operations on secrets"},
	{Text: "secrets", Description: "Operations on secrets"},
	{Text: "ssh", Description: "SSH into host"},
	{Text: "stack", Description: "Operations on stacks"},
	{Text: "start", Description: "Start or activate services/containers/stacks/hosts"},
	{Text: "stop", Description: "Stop or deactivate services/containers/stacks/hosts"},
	{Text: "up", Description: "Rancher up"},
	{Text: "volume", Description: "Operations on volumes"},
	{Text: "volumes", Description: "Operations on volumes"},
	{Text: "wait", Description: "Wait for resources services/containers/stacks/volumes/hosts/secrets"},
	{Text: "help", Description: "Help command"},
	{Text: "exit", Description: "Exit prompt mode"},
}

func argumentsCompleter(args []string) []prompt.Suggest {
	if len(args) <= 1 {
		return prompt.FilterHasPrefix(commands, args[0], true)
	}

	first := args[0]
	switch first {
	case "catalog":
		if len(args) == 2 {
			subcommands := []prompt.Suggest{
				{Text: "ls", Description: "List catalog templates"},
				{Text: "install", Description: "Install catalog template"},
				{Text: "help", Description: "Shows a list of commands or help for one command"},
			}
			return prompt.FilterHasPrefix(subcommands, args[1], true)
		}
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
	case "env", "environment":
		if len(args) == 2 {
			subcommands := []prompt.Suggest{
				{Text: "ls", Description: "List environments"},
				{Text: "create", Description: "Create an environment"},
				{Text: "templates", Description: "Interact with environment templates"},
				{Text: "template", Description: "Interact with environment templates"},
				{Text: "rm", Description: "Remove environment(s)"},
				{Text: "deactivate", Description: "Deactivate environment(s)"},
				{Text: "activate", Description: "Activate environment(s)"},
				{Text: "help", Description: "Shows a list of commands or help for one command"},
			}
			return prompt.FilterHasPrefix(subcommands, args[1], true)
		}
	case "host":
		if len(args) == 2 {
			subcommands := []prompt.Suggest{
				{Text: "ls", Description: "List hosts"},
				{Text: "create", Description: "Create a host"},
			}
			return prompt.FilterHasPrefix(subcommands, args[1], true)
		}
	case "secret":
		if len(args) == 2 {
			subcommands := []prompt.Suggest{
				{Text: "ls", Description: "List secrets"},
				{Text: "create", Description: "Create a secret"},
			}
			return prompt.FilterHasPrefix(subcommands, args[1], true)
		}
	case "volume", "volumes":
		if len(args) == 2 {
			subcommands := []prompt.Suggest{
				{Text: "ls", Description: "List volumes"},
				{Text: "create", Description: "Delete volume"},
				{Text: "rm", Description: "Create volume"},
			}
			return prompt.FilterHasPrefix(subcommands, args[1], true)
		}
	default:
		return []prompt.Suggest{}
	}
	return []prompt.Suggest{}
}

package rancherPrompt

import (
	"strings"

	"github.com/c-bata/go-prompt"
)

func optionCompleter(args []string, long bool) []prompt.Suggest {
	l := len(args)
	if l <= 1 {
		if long {
			return prompt.FilterHasPrefix(optionHelp, "--", false)
		}
		return optionHelp
	}

	var suggests []prompt.Suggest
	commandArgs := excludeOptions(args)
	switch commandArgs[0] {
	case "catalog":
		suggests = append(flagCatalog, flagGlobal...)
	case "config":
		suggests = append(flagConfig, flagGlobal...)
	case "docker":
		suggests = append(flagDocker, flagGlobal...)
	case "env", "environment":
		suggests = append(flagEnv, flagGlobal...)
	case "events":
		suggests = append(flagEvents, flagGlobal...)
	case "exec":
		suggests = append(flagExec, flagGlobal...)
	case "export":
		suggests = append(flagExport, flagGlobal...)
	case "host":
		if len(commandArgs) > 1 {
			if commandArgs[1] == "ls" {
				suggests = append(flagHostLs, flagGlobal...)
			} else if commandArgs[1] == "create" {
				suggests = append(flagHostCreate, flagGlobal...)
			}
		}
	case "logs":
		suggests = append(flagLogs, flagGlobal...)
	case "ps":
		suggests = append(flagPs, flagGlobal...)
	case "restart":
		suggests = append(flagRestart, flagGlobal...)
	case "rm":
		suggests = append(flagRm, flagGlobal...)
	case "run":
		suggests = append(flagRun, flagGlobal...)
	case "secret":
		suggests = append(flagSecret, flagGlobal...)
	case "stack":
		suggests = append(flagStack, flagGlobal...)
	case "start":
		suggests = append(flagStart, flagGlobal...)
	case "stop":
		suggests = append(flagStop, flagGlobal...)
	case "up":
		suggests = append(flagUp, flagGlobal...)
	case "volume", "volumes":
		if len(commandArgs) > 1 {
			if commandArgs[1] == "ls" {
				suggests = append(flagVolumeLs, flagGlobal...)
			} else if commandArgs[1] == "create" {
				suggests = append(flagVolumeCreate, flagGlobal...)
			}
		}
	case "inspect":
		suggests = append(flagInspect, flagGlobal...)
	default:
		suggests = optionHelp
	}

	if long {
		return prompt.FilterContains(
			prompt.FilterHasPrefix(suggests, "--", false),
			strings.TrimLeft(args[l-1], "--"),
			true,
		)
	}
	return prompt.FilterContains(suggests, strings.TrimLeft(args[l-1], "-"), true)
}

var flagGlobal = []prompt.Suggest{
	{Text: "--debug", Description: "Debug logging"},
	{Text: "-c", Description: "Client configuration file (default ${HOME}/.rancher/cli.json)"},
	{Text: "--config", Description: "Client configuration file (default ${HOME}/.rancher/cli.json)"},
	{Text: "--env", Description: "Environment name or ID"},
	{Text: "--environment", Description: "Environment name or ID"},
	{Text: "--url", Description: "Specify the Rancher API endpoint URL"},
	{Text: "--access-key", Description: "Specify Rancher API access key"},
	{Text: "--secret-key", Description: "Specify Rancher API secret key"},
	{Text: "--host", Description: "Host used for docker command"},
	{Text: "-w", Description: "Wait for resource to reach resting state"},
	{Text: "--wait", Description: "Wait for resource to reach resting state"},
	{Text: "--wait-timeout", Description: "Debug logging"},
	{Text: "--wait-state", Description: "Timeout in seconds to wait"},
	{Text: "--help", Description: "Help command"},
}

var flagCatalog = []prompt.Suggest{
	{Text: "-q", Description: "Only display IDs"},
	{Text: "--quiet", Description: "Only display IDs"},
	{Text: "--format", Description: "json or Custom format: ‘{{.ID}} {{.Template.Id}}’"},
	{Text: "-s", Description: "Show system templates, not user"},
	{Text: "--system", Description: "Show system templates, not user"},
}

var flagConfig = []prompt.Suggest{
	{Text: "--print", Description: "Print the current configuration"},
}

var flagDocker = []prompt.Suggest{
	{Text: "--help-docker", Description: "Display the docker --help"},
}

var flagEnv = []prompt.Suggest{
	{Text: "--all", Description: "Show stop/inactive and recently removed resources"},
	{Text: "-a", Description: "Show stop/inactive and recently removed resources"},
	{Text: "--quiet", Description: "Only display IDs"},
	{Text: "-q", Description: "Only display IDs"},
	{Text: "--format", Description: "json or Custom format: ‘{{.ID}} {{.Environment.Name}}’"},
}

var flagEvents = []prompt.Suggest{
	{Text: "--format", Description: "json or Custom format: ‘{{.Name}} {{.Data.resource.kind}}’"},
	{Text: "--reconnect", Description: "Reconnect on error"},
	{Text: "--r", Description: "Reconnect on error"},
}

var flagExec = []prompt.Suggest{
	{Text: "--help-docker", Description: "Display the docker exec --help"},
	{Text: "--detach", Description: "Detached mode: run command in the background"},
	{Text: "-d", Description: "Detached mode: run command in the background"},
	{Text: "--detach-keys", Description: "Override the key sequence for detaching a container"},
	{Text: "--env", Description: "Set environment variables"},
	{Text: "-e", Description: "Set environment variables"},
	{Text: "--interactive", Description: "Keep STDIN open even if not attached"},
	{Text: "-i", Description: "Keep STDIN open even if not attached"},
	{Text: "--privileged", Description: "Give extended privileges to the command"},
	{Text: "--tty", Description: "Allocate a pseudo-TTY"},
	{Text: "-t", Description: "Allocate a pseudo-TTY"},
	{Text: "--user", Description: "Username or UID (format: <name|uid>[:<group|gid>])"},
	{Text: "-u", Description: "Username or UID (format: <name|uid>[:<group|gid>])"},
}

var flagExport = []prompt.Suggest{
	{Text: "--file", Description: "Write to a file, instead of local files, use - to write to STDOUT"},
	{Text: "-f", Description: "Write to a file, instead of local files, use - to write to STDOUT"},
	{Text: "--system", Description: "	If exporting the entire environment, include system"},
	{Text: "-s", Description: "	If exporting the entire environment, include system"},
}

var flagHostLs = []prompt.Suggest{
	{Text: "--all", Description: "Show stop/inactive and recently removed resources"},
	{Text: "-a", Description: "Show stop/inactive and recently removed resources"},
	{Text: "--quiet", Description: "Only display IDs"},
	{Text: "-q", Description: "Only display IDs"},
	{Text: "--format", Description: "json or Custom format: ‘{{.ID}} {{.Host.Hostname}}’"},
}

var flagHostCreate = []prompt.Suggest{
	{Text: "--driver", Description: "Driver to create machine with. [$MACHINE_DRIVER]"},
	{Text: "-d", Description: "Driver to create machine with. [$MACHINE_DRIVER]"},
	{Text: "--engine-env", Description: "Specify environment variables to set in the engine"},
	{Text: "--engine-insecure-registry", Description: "Specify insecure registries to allow with the created engine"},
	{Text: "--engine-install-url", Description: "Custom URL to use for engine installation [$MACHINE_DOCKER_INSTALL_URL]"},
	{Text: "--engine-labe", Description: "Specify labels for the created engine"},
	{Text: "--engine-opt", Description: "Specify arbitrary flags to include with the created engine in the form flag=value"},
	{Text: "--engine-registry-mirror", Description: "Specify registry mirrors to use [$ENGINE_REGISTRY_MIRROR]"},
	{Text: "--engine-storage-driver", Description: "Specify a storage driver to use with the engine"},
}

var flagLogs = []prompt.Suggest{
	{Text: "--service", Description: "Show service logs"},
	{Text: "-s", Description: "Show service logs"},
	{Text: "--sub-log", Description: "Show service sub logs"},
	{Text: "--follow", Description: "Follow log output"},
	{Text: "-f", Description: "Follow log output"},
	{Text: "--tail", Description: "	Number of lines to show from the end of the logs (default: 100)"},
	{Text: "--since", Description: "Show logs since timestamp"},
	{Text: "--timestamps", Description: "Show timestamps"},
}

var flagPs = []prompt.Suggest{
	{Text: "--all", Description: "Show stop/inactive and recently removed resources"},
	{Text: "-a", Description: "Show stop/inactive and recently removed resources"},
	{Text: "--system", Description: "Show system resources"},
	{Text: "-s", Description: "Show system resources"},
	{Text: "--containers", Description: "Display containers"},
	{Text: "-c", Description: "Display containers"},
	{Text: "--quiet", Description: "Only display IDs"},
	{Text: "-q", Description: "Only display IDs"},
	{Text: "--format", Description: "json or Custom format: ‘{{.Service.Id}} {{.Service.Name}} {{.Service.LaunchConfig.ImageUuid}}’"},
}

var flagRestart = []prompt.Suggest{
	{Text: "--type", Description: "Restrict restart to specific types (service, container)"},
	{Text: "--batch-size", Description: "Number of containers to restart at a time (default: 1)"},
	{Text: "--interval", Description: "Interval in millisecond to wait between restarts (default: 1000)"},
}

var flagRm = []prompt.Suggest{
	{Text: "--type", Description: "Restrict delete to specific types"},
	{Text: "--stop", Description: "Stop or deactivate resource first if needed before deleting"},
	{Text: "-s", Description: "Stop or deactivate resource first if needed before deleting"},
}

var flagRun = []prompt.Suggest{
	{Text: "--blkio-weight", Description: "Block IO (relative weight), between 10 and 1000, or 0 to disable (default 0)"},
	{Text: "--cpu-quota", Description: "Limit CPU CFS (Completely Fair Scheduler) quota"},
	{Text: "--cpu-shares", Description: "CPU shares (relative weight)"},
	{Text: "--cap-add", Description: "Add Linux capabilities"},
	{Text: "--cap-drop", Description: "Drop Linux capabilities"},
	{Text: "--cgroup-parent", Description: "Optional parent cgroup for the container"},
	{Text: "--cpu-period", Description: "Limit CPU CFS (Completely Fair Scheduler) period"},
	{Text: "--cpuset-mems", Description: "MEMs in which to allow execution (0-3, 0,1)"},
	{Text: "--device", Description: "Add a host device to the container"},
	{Text: "--dns", Description: "Set custom DNS servers"},
	{Text: "--dns-opt", Description: "Set DNS options"},
	{Text: "--dns-option", Description: "Set DNS options"},
	{Text: "--dns-search", Description: "Set custom DNS search domains"},
	{Text: "--entrypoint", Description: "Overwrite the default ENTRYPOINT of the image"},
	{Text: "--expose", Description: "Expose a port or a range of ports"},
	{Text: "--group-add", Description: "Add additional groups to join"},
	{Text: "--health-cmd", Description: "Command to run to check health"},
	{Text: "--health-interval", Description: "Time between running the check (ms|s|m|h) (default 0s)"},
	{Text: "--health-retries", Description: "Consecutive failures needed to report unhealthy"},
	{Text: "--health-timeout", Description: "Maximum time to allow one check to run (ms|s|m|h) (default 0s)"},
	{Text: "--hostname", Description: "Container host name"},
	{Text: "--init", Description: "Run an init inside the container that forwards signals and reaps processes"},
	{Text: "-i", Description: "Keep STDIN open even if not attached"},
	{Text: "--interactive", Description: "Keep STDIN open even if not attached"},
	{Text: "--ip", Description: "IPv4 address (e.g., 172.30.100.104)"},
	{Text: "--ip6", Description: "IPv6 address (e.g., 2001:db8::33)"},
	{Text: "--ipc", Description: "IPC namespace to use"},
	{Text: "--isolation", Description: "Container isolation technology"},
	{Text: "--kernel-memory", Description: "Kernel memory limit"},
	{Text: "-m", Description: "Memory limit"},
	{Text: "--memory", Description: "Memory limit"},
	{Text: "--memory-reservation", Description: "Memory soft limit"},
	{Text: "--memory-swap", Description: "Swap limit equal to memory plus swap: '-1' to enable unlimited swap"},
	{Text: "--memory-swappiness", Description: "Tune container memory swappiness (0 to 100)"},
	{Text: "--name", Description: "Assign a name to the container"},
	{Text: "--net-alias", Description: "Add network-scoped alias for the container"},
	{Text: "--network-alias", Description: "Add network-scoped alias for the container"},
	{Text: "--oom-kill-disable", Description: "Disable OOM Killer"},
	{Text: "--oom-score-adj", Description: "Tune host’s OOM preferences (-1000 to 1000)"},
	{Text: "-P", Description: "Publish all exposed ports to random ports"},
	{Text: "--publish-all", Description: "Publish all exposed ports to random ports"},
	{Text: "-p", Description: "Publish a container's `port`(s) to the host"},
	{Text: "--publish", Description: "Publish a container's `port`(s) to the host"},
	{Text: "--pid", Description: "PID namespace to use"},
	{Text: "--pids-limit", Description: "Tune container pids limit (set -1 for unlimited)"},
	{Text: "--privileged", Description: "Give extended privileges to this container"},
	{Text: "--read-only", Description: "Mount the container's root filesystem as read only"},
	{Text: "--security-opt", Description: "Security Options"},
	{Text: "--shm-size", Description: "Size of /dev/shm"},
	{Text: "-t", Description: "Allocate a pseudo-TTY"},
	{Text: "--tty", Description: "Allocate a pseudo-TTY"},
	{Text: "-u", Description: "Username or UID (format: <name|uid>[:<group|gid>])"},
	{Text: "--user", Description: "Username or UID (format: <name|uid>[:<group|gid>])"},
	{Text: "--volume-driver", Description: "Optional volume driver for the container"},
	{Text: "-w", Description: "Working directory inside the container"},
	{Text: "--workdir", Description: "Working directory inside the container"},
	{Text: "--log-driver", Description: "Logging driver for container"},
	{Text: "--log-opt", Description: "Log driver options"},
	{Text: "--uts", Description: "UTS namespace to use"},
	{Text: "-v", Description: "Bind mount a volume"},
	{Text: "--volume", Description: "Bind mount a volume"},
	{Text: "--net", Description: "Connect a container to a network: host, none, bridge, managed"},
	{Text: "--scale", Description: "Number of containers to run"},
	{Text: "--schedule-global", Description: "Run 1 container per host"},
	{Text: "--stop-signal", Description: "Signal to stop a container"},
	{Text: "-l", Description: "Add label in the form of key=value"},
	{Text: "--label", Description: "Add label in the form of key=value"},
	{Text: "-e", Description: "Add label in the form of key=value"},
	{Text: "--env", Description: "Set one or more environment variable in the form of key=value, key=, and key"},
	{Text: "--pull", Description: "Always pull image on container start"},
	{Text: "--it", Description: "Combined option for interactive and tty"},
}

var flagSecret = []prompt.Suggest{
	{Text: "--quiet", Description: "Only display IDs"},
	{Text: "-q", Description: "Only display IDs"},
	{Text: "--format", Description: "json or Custom format: ‘{{.ID}} {{.Host.Hostname}}’"},
}

var flagStack = []prompt.Suggest{
	{Text: "--system", Description: "Show system resources"},
	{Text: "-s", Description: "Show system resources"},
	{Text: "--quiet", Description: "Only display IDs"},
	{Text: "-q", Description: "Only display IDs"},
	{Text: "--format", Description: "json or Custom format: ‘{{.ID}} {{.Host.Hostname}}’"},
	{Text: "--start", Description: "Start stack on create"},
	{Text: "-e", Description: "Create an empty stack"},
	{Text: "--empty", Description: "Create an empty stack"},
	{Text: "--docker-compose", Description: "Docker Compose file (default: “docker-compose.yml”)"},
	{Text: "-f", Description: "Docker Compose file (default: “docker-compose.yml”)"},
	{Text: "--rancher-compose", Description: "Rancher Compose file (default: “rancher-compose.yml”)"},
	{Text: "-r", Description: "Rancher Compose file (default: “rancher-compose.yml”)"},
}

var flagStart = []prompt.Suggest{
	{Text: "--type", Description: "Restrict start/activate to specific types (service, container, host, stack)"},
}

var flagStop = []prompt.Suggest{
	{Text: "--type", Description: "Restrict start/activate to specific types (service, container, host, stack)"},
}

var flagUp = []prompt.Suggest{
	{Text: "--pull", Description: ""},
	{Text: "-p", Description: ""},
	{Text: "--upgrade", Description: "Upgrade if service has changed"},
	{Text: "-u", Description: "Upgrade if service has changed"},
	{Text: "--recreate", Description: "Upgrade if service has changed"},
	{Text: "--force-upgrade", Description: "Upgrade regardless if service has changed"},
	{Text: "--force-recreate", Description: "Upgrade regardless if service has changed"},
	{Text: "-c", Description: "Confirm that the upgrade was success and delete old containers"},
	{Text: "--confirm-upgrade", Description: "Confirm that the upgrade was success and delete old containers"},
	{Text: "-r", Description: "Rollback to the previous deployed version"},
	{Text: "--roll-back", Description: "Rollback to the previous deployed version"},
	{Text: "--batch-size", Description: "Number of containers to upgrade at once (default: 2)"},
	{Text: "--interval", Description: "Update interval in milliseconds (default: 1000)"},
	{Text: "--rancher-file", Description: "Specify an alternate Rancher compose file (default: rancher-compose.yml)"},
	{Text: "-e", Description: "Specify a file from which to read environment variables"},
	{Text: "--env-file", Description: "Specify a file from which to read environment variables"},
	{Text: "--file", Description: "Specify one or more alternate compose files (default: docker-compose.yml) [$COMPOSE_FILE]"},
	{Text: "-f", Description: "Specify one or more alternate compose files (default: docker-compose.yml) [$COMPOSE_FILE]"},
	{Text: "-s", Description: "Specify an alternate project name (default: directory name)"},
	{Text: "--stack", Description: "Specify an alternate project name (default: directory name)"},
}

var flagVolumeLs = []prompt.Suggest{
	{Text: "--quiet", Description: "Only display IDs"},
	{Text: "-q", Description: "Only display IDs"},
	{Text: "--format", Description: "json or Custom format: ‘{{.ID}} {{.Host.Hostname}}’"},
	{Text: "-a", Description: "	Show stop/inactive and recently removed resources"},
	{Text: "--all", Description: "	Show stop/inactive and recently removed resources"},
}

var flagVolumeCreate = []prompt.Suggest{
	{Text: "--driver", Description: "Specify volume driver name"},
	{Text: "--opt", Description: "Set driver specific key/value options"},
}

var flagInspect = []prompt.Suggest{
	{Text: "--type", Description: "Restrict inspect to specific types (service, container, host)"},
	{Text: "--links", Description: "Include URLs to actions and links in resource output"},
	{Text: "--format", Description: "json or Custom format: ‘{{.kind}}’ (default: “json”)"},
}

var optionHelp = []prompt.Suggest{
	{Text: "-h"},
	{Text: "--help"},
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

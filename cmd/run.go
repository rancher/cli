package cmd

import (
	"strings"

	"github.com/rancher/go-rancher/v2"

	"fmt"
	"os"

	"github.com/urfave/cli"
)

/*
  -a, --attach=[]                 Attach to STDIN, STDOUT or STDERR
  --add-host=[]                   Add a custom host-to-IP mapping (host:ip)
  --blkio-weight                  Block IO (relative weight), between 10 and 1000
  --blkio-weight-device=[]        Block IO weight (relative device weight)
  --cgroup-parent                 Optional parent cgroup for the container
  --cidfile                       Write the container ID to the file
  --cpu-period                    Limit CPU CFS (Completely Fair Scheduler) period
  --cpu-quota                     Limit CPU CFS (Completely Fair Scheduler) quota
  --cpuset-cpus                   CPUs in which to allow execution (0-3, 0,1)
  --cpuset-mems                   MEMs in which to allow execution (0-3, 0,1)
  -d, --detach                    Run container in background and print container ID
  --detach-keys                   Override the key sequence for detaching a container
  --device-read-bps=[]            Limit read rate (bytes per second) from a device
  --device-read-iops=[]           Limit read rate (IO per second) from a device
  --device-write-bps=[]           Limit write rate (bytes per second) to a device
  --device-write-iops=[]          Limit write rate (IO per second) to a device
  --disable-content-trust=true    Skip image verification
  --dns-opt=[]                    Set DNS options
  -e, --env=[]                    Set environment variables
  --env-file=[]                   Read in a file of environment variables
  --group-add=[]                  Add additional groups to join
  --help                          Print usage
  --ip                            Container IPv4 address (e.g. 172.30.100.104)
  --ip6                           Container IPv6 address (e.g. 2001:db8::33)
  --ipc                           IPC namespace to use
  --isolation                     Container isolation level
  --kernel-memory                 Kernel memory limit
  -l, --label=[]                  Set meta data on a container
  --label-file=[]                 Read in a line delimited file of labels
  --link=[]                       Add link to another container
  --log-driver                    Logging driver for container
  --log-opt=[]                    Log driver options
  --mac-address                   Container MAC address (e.g. 92:d0:c6:0a:29:33)
  --memory-reservation            Memory soft limit
  --memory-swappiness=-1          Tune container memory swappiness (0 to 100)
  --net=default                   Connect a container to a network
  --net-alias=[]                  Add network-scoped alias for the container
  --oom-kill-disable              Disable OOM Killer
  --oom-score-adj                 Tune host's OOM preferences (-1000 to 1000)
  --restart=no                    Restart policy to apply when a container exits
  --rm                            Automatically remove the container when it exits
  --shm-size                      Size of /dev/shm, default value is 64MB
  --sig-proxy=true                Proxy received signals to the process
  --stop-signal=SIGTERM           Signal to stop a container, SIGTERM by default
  --tmpfs=[]                      Mount a tmpfs directory
  --ulimit=[]                     Ulimit options
  --uts                           UTS namespace to use
  -v, --volume=[]                 Bind mount a volume
  --volumes-from=[]               Mount volumes from the specified container(s)
*/

func RunCommand() cli.Command {
	return cli.Command{
		Name:   "run",
		Usage:  "Run services",
		Action: serviceRun,
		Flags: []cli.Flag{
			cli.Int64Flag{
				Name:  "blkio-weight",
				Usage: "Block IO (relative weight), between 10 and 1000, or 0 to disable (default 0)",
			},
			cli.Int64Flag{
				Name:  "cpu-quota",
				Usage: "Limit CPU CFS (Completely Fair Scheduler) quota",
			},
			cli.Int64Flag{
				Name:  "cpu-shares",
				Usage: "CPU shares (relative weight)",
			},
			cli.StringSliceFlag{
				Name:  "cap-add",
				Usage: "Add Linux capabilities",
			},
			cli.StringSliceFlag{
				Name:  "cap-drop",
				Usage: "Drop Linux capabilities",
			},
			cli.StringFlag{
				Name:  "cgroup-parent",
				Usage: "Optional parent cgroup for the container",
			},
			cli.Int64Flag{
				Name:  "cpu-period",
				Usage: "Limit CPU CFS (Completely Fair Scheduler) period",
			},
			cli.StringFlag{
				Name:  "cpuset-mems",
				Usage: "MEMs in which to allow execution (0-3, 0,1)",
			},
			cli.StringSliceFlag{
				Name:  "device",
				Usage: "Add a host device to the container",
			},
			cli.StringSliceFlag{
				Name:  "dns",
				Usage: "Set custom DNS servers",
			},
			cli.StringSliceFlag{
				Name:  "dns-opt, dns-option",
				Usage: "Set DNS options",
			},
			cli.StringSliceFlag{
				Name:  "dns-search",
				Usage: "Set custom DNS search domains",
			},
			cli.StringSliceFlag{
				Name:  "entrypoint",
				Usage: "Overwrite the default ENTRYPOINT of the image",
			},
			cli.StringSliceFlag{
				Name:  "expose",
				Usage: "Expose a port or a range of ports",
			},
			cli.StringSliceFlag{
				Name:  "group-add",
				Usage: "Add additional groups to join",
			},
			cli.StringSliceFlag{
				Name:  "health-cmd",
				Usage: "Command to run to check health",
			},
			cli.Int64Flag{
				Name:  "health-interval",
				Usage: "Time between running the check (ms|s|m|h) (default 0s)",
			},
			cli.Int64Flag{
				Name:  "health-retries",
				Usage: "Consecutive failures needed to report unhealthy",
			},
			cli.Int64Flag{
				Name:  "health-timeout",
				Usage: "Maximum time to allow one check to run (ms|s|m|h) (default 0s)",
			},
			cli.StringFlag{
				Name:  "hostname",
				Usage: "Container host name",
			},
			cli.BoolFlag{
				Name:  "init",
				Usage: "Run an init inside the container that forwards signals and reaps processes",
			},
			cli.BoolFlag{
				Name:  "interactive, i",
				Usage: "Keep STDIN open even if not attached",
			},
			cli.Int64Flag{
				Name:  "kernel-memory",
				Usage: "Kernel memory limit",
			},
			cli.Int64Flag{
				Name:  "memory, m",
				Usage: "Memory limit",
			},
			cli.Int64Flag{
				Name:  "memory-reservation",
				Usage: "Memory soft limit",
			},
			cli.Int64Flag{
				Name:  "memory-swap",
				Usage: "Swap limit equal to memory plus swap: '-1' to enable unlimited swap",
			},
			cli.Int64Flag{
				Name:  "memory-swappiness",
				Usage: "Tune container memory swappiness (0 to 100)",
			},
			cli.StringFlag{
				Name:  "name",
				Usage: "Assign a name to the container",
			},
			cli.BoolFlag{
				Name:  "oom-kill-disable",
				Usage: "Disable OOM Killer",
			},
			cli.Int64Flag{
				Name:  "oom-score-adj",
				Usage: "Tune host’s OOM preferences (-1000 to 1000)",
			},
			cli.BoolFlag{
				Name:  "publish-all, P",
				Usage: "Publish all exposed ports to random ports",
			},
			cli.StringSliceFlag{
				Name:  "publish, p",
				Usage: "Publish a container's `port`(s) to the host",
			},
			cli.StringFlag{
				Name:  "pid",
				Usage: "PID namespace to use",
			},
			cli.Int64Flag{
				Name:  "pids-limit",
				Usage: "Tune container pids limit (set -1 for unlimited)",
			},
			cli.BoolFlag{
				Name:  "privileged",
				Usage: "Give extended privileges to this container",
			},
			cli.BoolFlag{
				Name:  "read-only",
				Usage: "Mount the container's root filesystem as read only",
			},
			cli.StringSliceFlag{
				Name:  "security-opt",
				Usage: "Security Options",
			},
			cli.Int64Flag{
				Name:  "shm-size",
				Usage: "Size of /dev/shm",
			},
			cli.BoolFlag{
				Name:  "tty, t",
				Usage: "Allocate a pseudo-TTY",
			},
			cli.StringFlag{
				Name:  "user, u",
				Usage: "Username or UID (format: <name|uid>[:<group|gid>])",
			},
			cli.StringFlag{
				Name:  "volume-driver",
				Usage: "Optional volume driver for the container",
			},
			cli.StringFlag{
				Name:  "workdir, w",
				Usage: "Working directory inside the container",
			},
			cli.StringFlag{
				Name:  "log-driver",
				Usage: "Logging driver for container",
			},
			cli.StringSliceFlag{
				Name:  "log-opt",
				Usage: "Log driver options",
			},
			cli.StringFlag{
				Name:  "uts",
				Usage: "UTS namespace to use",
			},
			cli.StringSliceFlag{
				Name:  "volume, v",
				Usage: "Bind mount a volume",
			},
			cli.StringFlag{
				Name:  "net",
				Usage: "Connect a container to a network: host, none, bridge, managed",
				Value: "managed",
			},
			cli.IntFlag{
				Name:  "scale",
				Usage: "Number of containers to run",
				Value: 1,
			},
			cli.BoolFlag{
				Name:  "schedule-global",
				Usage: "Run 1 container per host",
			},
			cli.StringFlag{
				Name:  "stop-signal",
				Usage: "Signal to stop a container",
			},
			cli.StringSliceFlag{
				Name:  "label,l",
				Usage: "Add label in the form of key=value",
			},
			cli.StringSliceFlag{
				Name:  "env,e",
				Usage: "Set one or more environment variable in the form of key=value, key=, and key",
			},
			cli.BoolFlag{
				Name:  "pull",
				Usage: "Always pull image on container start",
			},
		},
	}
}

func ParseName(c *client.RancherClient, name string) (*client.Stack, string, error) {
	stackName := ""
	serviceName := name

	parts := strings.SplitN(name, "/", 2)
	if len(parts) == 2 {
		stackName = parts[0]
		serviceName = parts[1]
	}

	stack, err := GetOrCreateDefaultStack(c, stackName)
	if err != nil {
		return stack, "", err
	}

	if serviceName == "" {
		serviceName = RandomName()
	}

	return stack, serviceName, nil
}

func serviceRun(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if ctx.NArg() < 1 {
		return cli.NewExitError("Image name is required as the first argument", 1)
	}

	if err != nil {
		return err
	}
	launchConfig := &client.LaunchConfig{
		//BlkioDeviceOptions:
		BlkioWeight: ctx.Int64("blkio-weight"),
		CapAdd:      ctx.StringSlice("cap-add"),
		CapDrop:     ctx.StringSlice("cap-drop"),
		//CpuSet: ctx.String(""),
		CgroupParent:   ctx.String("cgroup-parent"),
		CpuSetMems:     ctx.String("cpuset-mems"),
		CpuPeriod:      ctx.Int64("cpu-period"),
		CpuQuota:       ctx.Int64("cpu-quota"),
		CpuShares:      ctx.Int64("cpu-shares"),
		Devices:        ctx.StringSlice("device"),
		Dns:            ctx.StringSlice("dns"),
		DnsOpt:         ctx.StringSlice("dns-opt"),
		DnsSearch:      ctx.StringSlice("dns-search"),
		EntryPoint:     ctx.StringSlice("entrypoint"),
		Expose:         ctx.StringSlice("expose"),
		GroupAdd:       ctx.StringSlice("group-add"),
		HealthCmd:      ctx.StringSlice("health-cmd"),
		HealthTimeout:  ctx.Int64("health-timeout"),
		HealthInterval: ctx.Int64("health-interval"),
		HealthRetries:  ctx.Int64("health-retries"),
		Hostname:       ctx.String("hostname"),
		ImageUuid:      "docker:" + ctx.Args()[0],
		KernelMemory:   ctx.Int64("kernel-memory"),
		Labels:         map[string]interface{}{},
		Environment:    map[string]interface{}{},
		//LogConfig:
		Memory:            ctx.Int64("memory"),
		MemoryReservation: ctx.Int64("memory-reservation"),
		MemorySwap:        ctx.Int64("memory-swap"),
		MemorySwappiness:  ctx.Int64("memory-swappiness"),
		//NetworkIds: ctx.StringSlice("networkids"),
		NetworkMode:     ctx.String("net"),
		OomKillDisable:  ctx.Bool("oom-kill-disable"),
		OomScoreAdj:     ctx.Int64("oom-score-adj"),
		PidMode:         ctx.String("pid"),
		PidsLimit:       ctx.Int64("pids-limit"),
		Ports:           ctx.StringSlice("publish"),
		Privileged:      ctx.Bool("privileged"),
		PublishAllPorts: ctx.Bool("publish-all"),
		ReadOnly:        ctx.Bool("read-only"),
		RunInit:         ctx.Bool("init"),
		SecurityOpt:     ctx.StringSlice("security-opt"),
		ShmSize:         ctx.Int64("shm-size"),
		StdinOpen:       ctx.Bool("interactive"),
		StopSignal:      ctx.String("stop-signal"),
		Tty:             ctx.Bool("tty"),
		User:            ctx.String("user"),
		Uts:             ctx.String("uts"),
		VolumeDriver:    ctx.String("volume-driver"),
		WorkingDir:      ctx.String("workdir"),
		DataVolumes:     ctx.StringSlice("volume"),
	}

	if ctx.String("log-driver") != "" || len(ctx.StringSlice("log-opt")) > 0 {
		launchConfig.LogConfig = &client.LogConfig{
			Driver: ctx.String("log-driver"),
			Config: map[string]interface{}{},
		}
		for _, opt := range ctx.StringSlice("log-opt") {
			parts := strings.SplitN(opt, "=", 2)
			if len(parts) > 1 {
				launchConfig.LogConfig.Config[parts[0]] = parts[1]
			} else {
				launchConfig.LogConfig.Config[parts[0]] = ""
			}
		}
	}

	for _, label := range ctx.StringSlice("label") {
		parts := strings.SplitN(label, "=", 2)
		value := ""
		if len(parts) > 1 {
			value = parts[1]
		}
		launchConfig.Labels[parts[0]] = value
	}

	for _, env := range ctx.StringSlice("env") {
		parts := strings.SplitN(env, "=", 2)
		value := ""

		if len(parts) > 1 {
			value = parts[1]

			if parts[0] == "" {
				errMsg := fmt.Sprintf("invalid argument \"%s\" for e: invalid environment variable: %s\nSee 'rancher run --help'.", env, env)
				return cli.NewExitError(errMsg, 1)
			}
		} else if len(parts) == 1 {
			value = os.Getenv(parts[0])
		}
		launchConfig.Environment[parts[0]] = value
	}

	if ctx.Bool("schedule-global") {
		launchConfig.Labels["io.rancher.scheduler.global"] = "true"
	}

	if ctx.Bool("pull") {
		launchConfig.Labels["io.rancher.container.pull_image"] = "always"
	}

	args := ctx.Args()[1:]

	if len(args) > 0 {
		launchConfig.Command = args
	}

	stack, name, err := ParseName(c, ctx.String("name"))
	if err != nil {
		return err
	}

	service := &client.Service{
		Name:          name,
		StackId:       stack.Id,
		LaunchConfig:  launchConfig,
		StartOnCreate: true,
		Scale:         int64(ctx.Int("scale")),
	}

	service, err = c.Service.Create(service)
	if err != nil {
		return err
	}

	return WaitFor(ctx, service.Id)
}

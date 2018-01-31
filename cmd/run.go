package cmd

import (
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
  --disable-content-trust=true    Skip Image verification
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
				Usage: "Overwrite the default ENTRYPOINT of the Image",
			},
			cli.StringSliceFlag{
				Name:  "expose",
				Usage: "Expose a port or a range of ports",
			},
			cli.StringSliceFlag{
				Name:  "group-add",
				Usage: "Add additional groups to join",
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
				Usage: "Always pull Image on container start",
			},
		},
	}
}

func serviceRun(ctx *cli.Context) error {
	return nil
}

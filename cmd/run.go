package cmd

import (
	"fmt"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/rancher/go-rancher/client"
)

/*
  -a, --attach=[]                 Attach to STDIN, STDOUT or STDERR
  --add-host=[]                   Add a custom host-to-IP mapping (host:ip)
  --blkio-weight                  Block IO (relative weight), between 10 and 1000
  --blkio-weight-device=[]        Block IO weight (relative device weight)
  --cpu-shares                    CPU shares (relative weight)
  --cap-add=[]                    Add Linux capabilities
  --cap-drop=[]                   Drop Linux capabilities
  --cgroup-parent                 Optional parent cgroup for the container
  --cidfile                       Write the container ID to the file
  --cpu-period                    Limit CPU CFS (Completely Fair Scheduler) period
  --cpu-quota                     Limit CPU CFS (Completely Fair Scheduler) quota
  --cpuset-cpus                   CPUs in which to allow execution (0-3, 0,1)
  --cpuset-mems                   MEMs in which to allow execution (0-3, 0,1)
  -d, --detach                    Run container in background and print container ID
  --detach-keys                   Override the key sequence for detaching a container
  --device=[]                     Add a host device to the container
  --device-read-bps=[]            Limit read rate (bytes per second) from a device
  --device-read-iops=[]           Limit read rate (IO per second) from a device
  --device-write-bps=[]           Limit write rate (bytes per second) to a device
  --device-write-iops=[]          Limit write rate (IO per second) to a device
  --disable-content-trust=true    Skip image verification
  --dns=[]                        Set custom DNS servers
  --dns-opt=[]                    Set DNS options
  --dns-search=[]                 Set custom DNS search domains
  -e, --env=[]                    Set environment variables
  --entrypoint                    Overwrite the default ENTRYPOINT of the image
  --env-file=[]                   Read in a file of environment variables
  --expose=[]                     Expose a port or a range of ports
  --group-add=[]                  Add additional groups to join
  -h, --hostname                  Container host name
  --help                          Print usage
  -i, --interactive               Keep STDIN open even if not attached
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
  -m, --memory                    Memory limit
  --mac-address                   Container MAC address (e.g. 92:d0:c6:0a:29:33)
  --memory-reservation            Memory soft limit
  --memory-swap                   Swap limit equal to memory plus swap: '-1' to enable unlimited swap
  --memory-swappiness=-1          Tune container memory swappiness (0 to 100)
  --name                          Assign a name to the container
  --net=default                   Connect a container to a network
  --net-alias=[]                  Add network-scoped alias for the container
  --oom-kill-disable              Disable OOM Killer
  --oom-score-adj                 Tune host's OOM preferences (-1000 to 1000)
  -P, --publish-all               Publish all exposed ports to random ports
  -p, --publish=[]                Publish a container's port(s) to the host
  --pid                           PID namespace to use
  --privileged                    Give extended privileges to this container
  --read-only                     Mount the container's root filesystem as read only
  --restart=no                    Restart policy to apply when a container exits
  --rm                            Automatically remove the container when it exits
  --security-opt=[]               Security Options
  --shm-size                      Size of /dev/shm, default value is 64MB
  --sig-proxy=true                Proxy received signals to the process
  --stop-signal=SIGTERM           Signal to stop a container, SIGTERM by default
  --tmpfs=[]                      Mount a tmpfs directory
  -u, --user                      Username or UID (format: <name|uid>[:<group|gid>])
  --ulimit=[]                     Ulimit options
  --uts                           UTS namespace to use
  -v, --volume=[]                 Bind mount a volume
  --volume-driver                 Optional volume driver for the container
  --volumes-from=[]               Mount volumes from the specified container(s)
  -w, --workdir                   Working directory inside the container
*/

func RunCommand() cli.Command {
	return cli.Command{
		Name:   "run",
		Usage:  "Run services",
		Action: serviceRun,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "interactive, i",
				Usage: "Keep STDIN open even if not attached",
			},
			cli.BoolFlag{
				Name:  "tty, t",
				Usage: "Allocate a pseudo-TTY",
			},
			cli.StringFlag{
				Name:  "name",
				Usage: "Assign a name to the container",
			},
			cli.StringSliceFlag{
				Name:  "publish, p",
				Usage: "Publish a container's `port`(s) to the host",
			},
		},
	}
}

func ParseName(c *client.RancherClient, name string) (*client.Environment, string, error) {
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
		serviceName = strings.Replace(namesgenerator.GetRandomName(0), "_", "-", -1)
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
		StdinOpen: ctx.Bool("interactive"),
		Tty:       ctx.Bool("tty"),
		ImageUuid: "docker:" + ctx.Args()[0],
		Ports:     ctx.StringSlice("publish"),
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
		EnvironmentId: stack.Id,
		LaunchConfig:  launchConfig,
		StartOnCreate: true,
	}

	service, err = c.Service.Create(service)
	if err != nil {
		return err
	}

	fmt.Println(service.Id)
	return nil
}

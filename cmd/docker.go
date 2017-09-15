package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/rancher-docker-api-proxy"
	"github.com/urfave/cli"
)

func DockerCommand() cli.Command {
	return cli.Command{
		Name:        "docker",
		Usage:       "Run docker CLI on a host",
		Description: "\nUses the $RANCHER_DOCKER_HOST to run docker commands. Use `--host <hostID>` or `--host <hostName>` to select a different host.\n\nExample:\n\t$ rancher --host 1h1 docker ps\n",
		Action:      hostDocker,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "help-docker",
				Usage: "Display the 'docker --help'",
			},
		},
		SkipFlagParsing: true,
	}
}

func hostDocker(ctx *cli.Context) error {
	return processExitCode(doDocker(ctx))
}

func doDocker(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		return cli.ShowCommandHelp(ctx, "docker")
	}

	if len(args) > 0 && args[0] == "--help-docker" {
		return runDockerHelp("")
	}

	hostname := ctx.GlobalString("host")
	if hostname == "" {
		return fmt.Errorf("--host is required")
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	return runDocker(hostname, c, args)
}

func runDockerCommand(hostname string, c *client.RancherClient, command string, args []string) error {
	return runDocker(hostname, c, append([]string{command}, args...))
}

func runDocker(hostname string, c *client.RancherClient, args []string) error {
	return runDockerWithOutput(hostname, c, args, os.Stdout, os.Stderr)
}

func determineAPIVersion(host *client.Host) string {
	version := host.Labels["io.rancher.host.docker_version"]
	parts := strings.Split(fmt.Sprint(version), ".")
	if len(parts) != 2 {
		return ""
	}
	num, err := strconv.Atoi(parts[1])
	if err != nil {
		return ""
	}

	return fmt.Sprintf("1.%d", num+12)
}

func runDockerWithOutput(hostname string, c *client.RancherClient, args []string,
	out, outErr io.Writer) error {
	resource, err := Lookup(c, hostname, "host")
	if err != nil {
		return err
	}

	host, err := c.Host.ById(resource.Id)
	if err != nil {
		return err
	}

	state := getHostState(host)
	if state != "active" && state != "inactive" && state != "disconnected" {
		return fmt.Errorf("Can not contact host %s in state %s", hostname, state)
	}

	apiVersion := determineAPIVersion(host)

	tempfile, err := ioutil.TempFile("", "docker-sock")
	if err != nil {
		return err
	}
	defer os.Remove(tempfile.Name())

	if err := tempfile.Close(); err != nil {
		return err
	}

	dockerHost := "unix://" + tempfile.Name()
	proxy := dockerapiproxy.NewProxy(c, host.Id, dockerHost)
	if err := proxy.Listen(); err != nil {
		return err
	}

	go func() {
		logrus.Fatal(proxy.Serve())
	}()

	var cmd *exec.Cmd
	if len(args) > 0 && args[0] == "--" {
		if len(args) > 1 {
			cmd = exec.Command(args[1], args[2:]...)
		} else {
			cmd = exec.Command(os.Getenv("SHELL"))
		}
		cmd.Env = append(os.Environ(), "debian_chroot=docker:"+hostname)
	} else {
		cmd = exec.Command("docker", args...)
		cmd.Env = os.Environ()
	}

	cmd.Env = append(cmd.Env, "DOCKER_HOST="+dockerHost)
	if apiVersion != "" {
		cmd.Env = append(cmd.Env, "DOCKER_API_VERSION="+apiVersion)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = out
	cmd.Stderr = outErr

	signal.Ignore(os.Interrupt)
	return cmd.Run()
}

package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher/client"
	"github.com/rancher/rancher-docker-api-proxy"
	"github.com/urfave/cli"
)

func DockerCommand() cli.Command {
	return cli.Command{
		Name:            "docker",
		Usage:           "Run docker CLI on a host",
		Action:          hostDocker,
		SkipFlagParsing: true,
	}
}

func hostDocker(ctx *cli.Context) error {
	return processExitCode(doDocker(ctx))
}

func doDocker(ctx *cli.Context) error {
	hostname := ctx.GlobalString("host")

	if hostname == "" {
		return fmt.Errorf("--host is required")
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	return runDocker(hostname, c, ctx.Args())
}

func runDockerCommand(hostname string, c *client.RancherClient, command string, args []string) error {
	return runDocker(hostname, c, append([]string{command}, args...))
}

func runDocker(hostname string, c *client.RancherClient, args []string) error {
	return runDockerWithOutput(hostname, c, args, os.Stdout, os.Stderr)
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
	if state != "active" {
		return fmt.Errorf("Can not contact host %s in state %s", hostname, state)
	}

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
	if len(args) == 1 && args[0] == "--" {
		cmd = exec.Command(os.Getenv("SHELL"), args[1:]...)
		cmd.Env = append(os.Environ(), "debian_chroot=docker:"+hostname)
	} else {
		cmd = exec.Command("docker", args...)
		cmd.Env = os.Environ()
	}

	cmd.Env = append(cmd.Env, "DOCKER_HOST="+dockerHost)
	cmd.Stdin = os.Stdin
	cmd.Stdout = out
	cmd.Stderr = outErr

	return cmd.Run()
}

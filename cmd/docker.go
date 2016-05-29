package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/rancher/rancher-docker-api-proxy"
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

	host, err := Lookup(c, hostname, "host")
	if err != nil {
		return err
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
	if len(ctx.Args()) == 1 && ctx.Args()[0] == "--" {
		cmd = exec.Command(os.Getenv("SHELL"), ctx.Args()[1:]...)
		cmd.Env = append(os.Environ(), "debian_chroot=docker:"+hostname)
	} else {
		cmd = exec.Command("docker", ctx.Args()...)
		cmd.Env = os.Environ()
	}

	cmd.Env = append(cmd.Env, "DOCKER_HOST="+dockerHost)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

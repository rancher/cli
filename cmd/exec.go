package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rancher/go-rancher/client"
	"github.com/urfave/cli"
)

func ExecCommand() cli.Command {
	return cli.Command{
		Name:            "exec",
		Usage:           "Run a command on a container",
		Description:     "\nThe command will find the container on the host and use `docker exec` to access the container. Any options that `docker exec` uses can be passed as an option for `rancher exec`.\n\nExample:\n\t$ rancher exec -i -t 1i1\n",
		Action:          execCommand,
		SkipFlagParsing: true,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "help-docker",
				Usage: "Display the 'docker exec --help'",
			},
		},
	}
}

func execCommand(ctx *cli.Context) error {
	return processExitCode(execCommandInternal(ctx))
}

func execCommandInternal(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		return cli.ShowCommandHelp(ctx, "exec")
	}

	if len(args) > 0 && args[0] == "--help-docker" {
		return runDockerHelp("exec")
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	args, hostID, _, err := selectContainer(c, ctx.Args())
	if err != nil {
		return err
	}

	// this is a massive hack. Need to fix the real issue
	args = append([]string{"-i"}, args...)
	return runDockerCommand(hostID, c, "exec", args)
}

func isHelp(args []string) bool {
	for _, i := range args {
		if i == "--help" || i == "-h" {
			return true
		}
	}

	return false
}

func selectContainer(c *client.RancherClient, args []string) ([]string, string, string, error) {
	newArgs := make([]string, len(args))
	copy(newArgs, args)

	name := ""
	index := 0
	for i, val := range newArgs {
		if !strings.HasPrefix(val, "-") {
			name = val
			index = i
			break
		}
	}

	if name == "" {
		return nil, "", "", fmt.Errorf("Please specify container name as an argument")
	}

	resource, err := Lookup(c, name, "container", "service")
	if err != nil {
		return nil, "", "", err
	}

	if _, ok := resource.Links["hosts"]; ok {
		hostID, containerID, err := getHostnameAndContainerID(c, resource.Id)
		if err != nil {
			return nil, "", "", err
		}

		newArgs[index] = containerID
		return newArgs, hostID, containerID, nil
	}

	if _, ok := resource.Links["instances"]; ok {
		var instances client.ContainerCollection
		if err := c.GetLink(*resource, "instances", &instances); err != nil {
			return nil, "", "", err
		}

		hostID, containerID, err := getHostnameAndContainerIDFromList(c, instances)
		if err != nil {
			return nil, "", "", err
		}
		newArgs[index] = containerID
		return newArgs, hostID, containerID, nil
	}

	return nil, "", "", nil
}

func getHostnameAndContainerIDFromList(c *client.RancherClient, containers client.ContainerCollection) (string, string, error) {
	if len(containers.Data) == 0 {
		return "", "", fmt.Errorf("Failed to find a container")
	}

	if len(containers.Data) == 1 {
		return containers.Data[0].HostId, containers.Data[0].ExternalId, nil
	}

	names := []string{}
	for _, container := range containers.Data {
		name := ""
		if container.Name == "" {
			name = container.Id
		} else {
			name = container.Name
		}
		names = append(names, fmt.Sprintf("%s (%s)", name, container.PrimaryIpAddress))
	}

	index := selectFromList("Containers:", names)
	return containers.Data[index].HostId, containers.Data[index].ExternalId, nil
}

func selectFromList(header string, choices []string) int {
	if header != "" {
		fmt.Println(header)
	}

	reader := bufio.NewReader(os.Stdin)
	selected := -1
	for selected <= 0 || selected > len(choices) {
		for i, choice := range choices {
			fmt.Printf("[%d] %s\n", i+1, choice)
		}
		fmt.Print("Select: ")

		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		num, err := strconv.Atoi(text)
		if err == nil {
			selected = num
		}
	}
	return selected - 1
}

func getHostnameAndContainerID(c *client.RancherClient, containerID string) (string, string, error) {
	container, err := c.Container.ById(containerID)
	if err != nil {
		return "", "", err
	}

	var hosts client.HostCollection
	if err := c.GetLink(container.Resource, "hosts", &hosts); err != nil {
		return "", "", err
	}

	if len(hosts.Data) != 1 {
		return "", "", fmt.Errorf("Failed to find host for container %s", container.Name)
	}

	return hosts.Data[0].Id, container.ExternalId, nil
}

func runDockerHelp(subcommand string) error {
	args := []string{"--help"}
	if subcommand != "" {
		args = []string{subcommand, "--help"}
	}
	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/rancher/go-rancher/client"
)

func ExecCommand() cli.Command {
	return cli.Command{
		Name:            "exec",
		Usage:           "Run a command on a container",
		SkipFlagParsing: true,
		HideHelp:        true,
		Action:          execCommand,
	}
}

func execCommand(ctx *cli.Context) error {
	return processExitCode(execCommandInternal(ctx))
}

func execCommandInternal(ctx *cli.Context) error {
	if isHelp(ctx.Args()) {
		return runDockerHelp("exec")
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	args, hostId, _, err := selectContainer(c, ctx.Args())
	if err != nil {
		return err
	}

	return runDockerCommand(hostId, c, "exec", args)
}

func isHelp(args []string) bool {
	for _, i := range args {
		if i == "--help" {
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
		hostId, containerId, err := getHostnameAndContainerId(c, resource.Id)
		if err != nil {
			return nil, "", "", err
		}

		newArgs[index] = containerId
		return newArgs, hostId, containerId, nil
	}

	if _, ok := resource.Links["instances"]; ok {
		var instances client.ContainerCollection
		if err := c.GetLink(*resource, "instances", &instances); err != nil {
			return nil, "", "", err
		}

		hostId, containerId, err := getHostnameAndContainerIdFromList(c, instances)
		if err != nil {
			return nil, "", "", err
		}
		newArgs[index] = containerId
		return newArgs, hostId, containerId, nil
	}

	return nil, "", "", nil
}

func getHostnameAndContainerIdFromList(c *client.RancherClient, containers client.ContainerCollection) (string, string, error) {
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

func getHostnameAndContainerId(c *client.RancherClient, containerId string) (string, string, error) {
	container, err := c.Container.ById(containerId)
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
	cmd := exec.Command("docker", subcommand, "--help")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/go-rancher/v3"
	"github.com/urfave/cli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

func ExecCommand() cli.Command {
	return cli.Command{
		Name:        "exec",
		Usage:       "Run a command on a container",
		Description: "\nThe command will find the container on the host and use `docker exec` to access the container. Any options that `docker exec` uses can be passed as an option for `rancher exec`.\n\nExample:\n\t$ rancher exec -it 1i1 bash\n",
		Action:      execCommand,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "interactive,i",
				Usage: "Keep STDIN open even if not attached",
			},
			cli.BoolFlag{
				Name:  "tty,t",
				Usage: "Allocate a pseudo-TTY",
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

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	args, hostID, containerID, err := selectContainer(c, ctx.Args())
	if err != nil {
		return err
	}

	containers, err := c.Container.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"externalId":   containerID,
			"removed_null": "true",
		},
	})
	if err != nil {
		return err
	}
	if len(containers.Data) == 0 {
		return errors.Errorf("Can't find the container with ID %s", containerID)
	}
	container := containers.Data[0]

	podName := container.Labels[podNameLabel]
	namespace := container.Labels[namespaceLabel]
	containerName := container.Labels[podContainerName]
	if podName != "" {
		return execCommandKube(c, ctx, podName, namespace, containerName)
	}
	if ctx.Bool("interactive") {
		args = append([]string{"-i"}, args...)
	}
	if ctx.Bool("tty") {
		args = append([]string{"-t"}, args...)
	}
	return runDockerCommand(hostID, c, "exec", args)
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

	if _, ok := resource.Links["host"]; ok {
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

	var host client.Host
	if err := c.GetLink(container.Resource, "host", &host); err != nil {
		return "", "", err
	}
	if host.Id == "" {
		return "", "", fmt.Errorf("Failed to find host for container %s", container.Name)
	}

	return host.Id, container.ExternalId, nil
}

func execCommandKube(c *client.RancherClient, ctx *cli.Context, podName, namespace, containerName string) error {
	conf, err := constructRestConfig(ctx, c)
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(conf)
	if err != nil {
		return err
	}
	pod, err := clientset.CoreV1Client.Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	restClient, err := restclient.RESTClientFor(conf)
	if err != nil {
		return err
	}
	t := setupTTY(ctx)
	stderr := "true"
	var sizeQueue remotecommand.TerminalSizeQueue
	if t.Raw {
		stderr = "false"
		sizeQueue = t.MonitorSize(t.GetSize())
	}
	commands := ctx.Args()
	if len(commands) < 2 {
		fmt.Println("error: no command is provided. example:\n\t$ rancher exec -it 1i1 bash")
		return nil
	}
	fn := func() error {
		req := restClient.Post().
			Resource("pods").
			Name(pod.Name).
			Namespace(pod.Namespace).
			SubResource("exec").
			Param("container", containerName).
			Param("stdin", strconv.FormatBool(ctx.Bool("interactive"))).
			Param("stdout", "true").
			Param("stderr", stderr).
			Param("tty", strconv.FormatBool(ctx.Bool("tty")))
		for _, command := range commands[1:] {
			req.Param("command", command)
		}
		exec, err := remotecommand.NewExecutor(conf, "POST", req.URL())
		if err != nil {
			return err
		}
		streamOpts := remotecommand.StreamOptions{
			Tty:               t.Raw,
			Stdout:            t.Out,
			TerminalSizeQueue: sizeQueue,
		}
		if ctx.Bool("interactive") {
			streamOpts.Stdin = t.In
		}
		if !t.Raw {
			streamOpts.Stderr = os.Stderr
		}
		return exec.Stream(streamOpts)
	}
	return t.Safe(fn)
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

package cmd

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/containerd/console"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/rancher/go-rancher/v2"
	"github.com/urfave/cli"
)

type execPayload struct {
	AttachStdin  bool     `json:"attachStdin"`
	AttachStdout bool     `json:"attachStdout"`
	Command      []string `json:"command"`
	TTY          bool     `json:"tty"`
}

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

	args, execURL, err := selectContainer(c, ctx.Args())
	if err != nil {
		return err
	}

	if execURL == "" {
		return errors.New("not authorized to exec into this container")
	}

	fmt.Println(execURL)

	payload := execPayload{
		AttachStdin:  true,
		AttachStdout: true,
		Command:      args,
		TTY:          true,
	}

	payloadJSON, err := json.Marshal(payload)
	if nil != err {
		return fmt.Errorf("json error: %v", err)
	}

	wsURL, err := resolveWebsocketURL(ctx, execURL, payloadJSON)
	if nil != err {
		return err
	}

	logrus.Debugf("websocket full URL: %s", wsURL)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{})
	if nil != err {
		logrus.Errorf("websocket dial error: %v", err)
		return err
	}
	defer conn.Close()

	current, err := console.ConsoleFromFile(os.Stdin)
	if err != nil {
		return err
	}
	defer current.Reset()

	err = current.SetRaw()
	if err != nil {
		return err
	}

	errChannel := make(chan error)

	// get input from stdin
	go getInput(conn, errChannel)
	// print out from the websocket
	go streamWebsocket(conn, errChannel)

	err = <-errChannel

	if err != nil && reflect.TypeOf(err).String() == "*websocket.CloseError" &&
		err.Error() == "websocket: close 1000 (normal)" {
		return nil
	}
	return err
}

func streamWebsocket(conn *websocket.Conn, errChannel chan<- error) {
	for {
		_, buf, err := conn.NextReader()
		if err != nil {
			errChannel <- err
			return
		}

		p, err := ioutil.ReadAll(buf)
		data, err := base64.StdEncoding.DecodeString(string(p))
		fmt.Print(string(data))
	}
}

func getInput(conn *websocket.Conn, errChannel chan<- error) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanBytes)
	for scanner.Scan() {
		if len(scanner.Bytes()) == 1 && scanner.Bytes()[0] == 3 {
			errChannel <- nil
			return
		}
		data := base64.StdEncoding.EncodeToString(scanner.Bytes())
		err := conn.WriteMessage(websocket.TextMessage, []byte(data))
		if nil != err {
			errChannel <- err
			return
		}

	}
}

func selectContainer(c *client.RancherClient, args []string) ([]string, string, error) {
	newArgs := make([]string, len(args))
	copy(newArgs, args)

	var name string
	for i, val := range newArgs {
		if !strings.HasPrefix(val, "-") {
			name = val
			newArgs = newArgs[i+1:]
			break
		}
	}

	if name == "" {
		return nil, "", fmt.Errorf("Please specify container name as an argument")
	}

	resource, err := Lookup(c, name, "container", "service")
	if err != nil {
		return nil, "", err
	}

	var containerExecURL string
	if _, ok := resource.Links["instances"]; ok {
		var instances client.ContainerCollection
		if err := c.GetLink(*resource, "instances", &instances); err != nil {
			return nil, "", err
		}

		containerExecURL, err = getContainerIDFromList(c, instances)
		if err != nil {
			return nil, "", err
		}
	} else {
		containerExecURL = resource.Actions["execute"]
	}

	return newArgs, containerExecURL, nil
}

func getContainerIDFromList(c *client.RancherClient, containers client.ContainerCollection) (string, error) {
	if len(containers.Data) == 0 {
		return "", fmt.Errorf("failed to find a container")
	}

	if len(containers.Data) == 1 {
		return containers.Data[0].Actions["execute"], nil
	}

	var names []string
	for _, container := range containers.Data {
		name := ""
		if container.Name == "" {
			name = container.Actions["execute"]
		} else {
			name = container.Name
		}
		names = append(names, fmt.Sprintf("%s (%s)", name, container.PrimaryIpAddress))
	}

	index := selectFromList("Containers:", names)
	return containers.Data[index].Actions["execute"], nil
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

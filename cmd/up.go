package cmd

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"io"

	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	dclient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/rancher/cli/monitor"
	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/rancher-docker-api-proxy"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

var colors = []color.Attribute{color.FgGreen, color.FgBlue, color.FgCyan, color.FgMagenta, color.FgRed, color.FgWhite, color.FgYellow}

func UpCommand() cli.Command {
	return cli.Command{
		Name:   "up",
		Usage:  "Bring all services up",
		Action: rancherUp,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "d",
				Usage: "Do not block and log",
			},
			cli.BoolFlag{
				Name:  "w",
				Usage: "wait for all the service to be up(block forever if there is no service)",
			},
			cli.BoolFlag{
				Name:  "rollback, r",
				Usage: "Rollback to the previous deployed version",
			},
			cli.StringSliceFlag{
				Name:   "file,f",
				Usage:  "Specify one or more alternate compose files (default: compose.yml)",
				Value:  &cli.StringSlice{},
				EnvVar: "COMPOSE_FILE",
			},
			cli.StringFlag{
				Name:  "stack,s",
				Usage: "Specify an alternate project name (default: directory name)",
			},
			cli.BoolFlag{
				Name:  "prune",
				Usage: "Prune services that doesn't exist on the current compose files",
			},
		},
	}
}

func rancherUp(ctx *cli.Context) error {
	rancherClient, err := GetClient(ctx)
	if err != nil {
		return err
	}

	// only look for --file or ./compose.yml
	composes, err := resolveComposeFile(ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to resolve compose file")
	}

	//resolve the stackName from --stack or current dir name
	stackName, err := resolveStackName(ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to resolve stackName")
	}

	// get existing stack by stack name
	stacks, err := rancherClient.Stack.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"name":         stackName,
			"removed_null": nil,
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to list stacks")
	}

	// rollback
	if ctx.Bool("rollback") {
		if len(stacks.Data) == 0 {
			return errors.Errorf("Can't find stack %v", stackName)
		}
		_, err := rancherClient.Stack.ActionRollback(&stacks.Data[0])
		if err != nil {
			return errors.Errorf("failed to rollback stack %v", stackName)
		}
		return nil
	}

	stackID, err := doUp(stacks, stackName, composes, ctx, rancherClient)
	if err != nil {
		return err
	}

	if !ctx.Bool("d") {
		watcher := monitor.NewUpWatcher(rancherClient)
		watcher.Subscribe()
		watchErr := make(chan error)
		logErr := make(chan error)
		go func(err chan error) { err <- watcher.Start(stackID) }(watchErr)
		services, err := watchServiceIds(stackID, rancherClient)
		if err != nil {
			return err
		}
		logrus.Infof("Stack %s is up", stackName)
		go func(err chan error) { err <- watchLogs(rancherClient, stackID, services) }(logErr)
		for {
			select {
			case err := <-watchErr:
				return errors.Errorf("Rancher up failed. Exiting Error: %v", err)
			case err := <-logErr:
				if err != nil {
					logrus.Warnf("Failed to watch container logs. Sleep 1 seconds and retry")
				}
				time.Sleep(time.Second * 1)
			}
		}
	}

	if ctx.Bool("w") {
		_, err := watchServiceIds(stackID, rancherClient)
		if err != nil {
			return err
		}
	}

	fmt.Println(stackID)
	return nil
}

func toUnixPath(p string) string {
	return strings.Replace(p, "\\", "/", -1)
}

func getStack(c *client.RancherClient, stackID string) (*client.Stack, error) {
	return c.Stack.ById(stackID)
}

func getServices(c *client.RancherClient, serviceIds []string) ([]client.Service, error) {
	services := []client.Service{}
	for _, serviceID := range serviceIds {
		service, err := c.Service.ById(serviceID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get service (id: [%v])", serviceID)
		}
		services = append(services, *service)
	}
	return services, nil
}

func getLogs(c *client.RancherClient, instanceIds map[string]struct{}) error {
	wg := sync.WaitGroup{}
	instances := []client.Instance{}
	for instanceID := range instanceIds {
		instance, err := c.Instance.ById(instanceID)
		if err != nil {
			return errors.Wrapf(err, "failed to get instance id [%v]", instanceID)
		}
		instances = append(instances, *instance)
	}
	listenSocks := map[string]*dclient.Client{}
	for _, i := range instances {
		if i.ExternalId == "" || i.HostId == "" {
			continue
		}

		if dockerClient, ok := listenSocks[i.HostId]; ok {
			wg.Add(1)
			go func(dockerClient *dclient.Client, i client.Instance) {
				log(i, dockerClient)
				wg.Done()
			}(dockerClient, i)
			continue
		}

		resource, err := Lookup(c, i.HostId, "host")
		if err != nil {
			return err
		}

		host, err := c.Host.ById(resource.Id)
		if err != nil {
			return err
		}

		state := getHostState(host)
		if state != "active" && state != "inactive" {
			logrus.Errorf("Can not contact host %s in state %s", i.HostId, state)
			continue
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

		dockerClient, err := dclient.NewClient(dockerHost, "", nil, nil)
		if err != nil {
			logrus.Errorf("Failed to connect to host %s: %v", i.HostId, err)
			continue
		}

		listenSocks[i.HostId] = dockerClient

		wg.Add(1)
		go func(dockerClient *dclient.Client, i client.Instance) {
			log(i, dockerClient)
			wg.Done()
		}(dockerClient, i)
	}
	wg.Wait()
	return nil
}

func log(instance client.Instance, dockerClient *dclient.Client) error {
	c, err := dockerClient.ContainerInspect(context.Background(), instance.ExternalId)
	if err != nil {
		return err
	}

	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "10",
	}
	responseBody, err := dockerClient.ContainerLogs(context.Background(), c.ID, options)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	if c.Config.Tty {
		_, err = io.Copy(os.Stdout, responseBody)
	} else {
		l := loggerFactory.CreateContainerLogger(instance.Name)
		_, err = stdcopy.StdCopy(writerFunc(l.Out), writerFunc(l.Err), responseBody)
	}
	return err
}

func resolveComposeFile(ctx *cli.Context) (map[string]string, error) {
	composeFiles := ctx.StringSlice("file")
	if len(composeFiles) == 0 {
		composeFiles = []string{"compose.yml"}
	}
	composes := map[string]string{}
	for _, composeFile := range composeFiles {
		fp, err := filepath.Abs(composeFile)
		if err != nil {
			return nil, errors.Wrap(err, "failed to lookup current directory name")
		}
		file, err := os.Open(fp)
		if err != nil {
			return nil, errors.Wrap(err, "Can not find compose.yml")
		}
		defer file.Close()
		buf, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read file")
		}
		composes[composeFile] = string(buf)
	}
	return composes, nil
}

func resolveStackName(ctx *cli.Context) (string, error) {
	if ctx.String("stack") != "" {
		return ctx.String("stack"), nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "Failed to get current working dir for stackName")
	}
	return path.Base(toUnixPath(wd)), nil
}

func doUp(stacks *client.StackCollection, stackName string, composes map[string]string, ctx *cli.Context, rancherClient *client.RancherClient) (string, error) {
	if len(stacks.Data) > 0 {
		// update stacks
		stacks.Data[0].Templates = composes
		prune := ctx.Bool("prune")
		if !ctx.Bool("d") {
			logrus.Infof("Updating stack %v", stackName)
		}
		_, err := rancherClient.Stack.Update(&stacks.Data[0], client.Stack{
			Templates: composes,
			Prune:     prune,
		})
		if err != nil {
			return "", errors.Wrapf(err, "failed to update stack %v", stackName)
		}
		return stacks.Data[0].Id, nil
	}
	// create new stack
	if !ctx.Bool("d") {
		logrus.Infof("Creating Stack %s", stackName)
	}
	prune := ctx.Bool("prune")
	stack, err := rancherClient.Stack.Create(&client.Stack{
		Name:      stackName,
		Templates: composes,
		Prune:     prune,
	})
	if err != nil {
		return "", errors.Wrapf(err, "failed to create stack %v", stackName)
	}
	return stack.Id, nil
}

func watchLogs(rancherClient *client.RancherClient, stackID string, services []client.Service) error {
	instanceIds := map[string]struct{}{}
	for _, service := range services {
		for _, instanceID := range service.InstanceIds {
			instanceIds[instanceID] = struct{}{}
		}
	}
	if err := getLogs(rancherClient, instanceIds); err != nil {
		return errors.Wrapf(err, "failed to get container logs")
	}
	return nil
}

func watchServiceIds(stackID string, rancherClient *client.RancherClient) ([]client.Service, error) {
	for {
		stack, err := getStack(rancherClient, stackID)
		if err != nil {
			return nil, err
		}
		if stack.Transitioning == "error" {
			return nil, errors.Errorf("Failed to up stack %s. Error: %s", stack.Name, stack.TransitioningMessage)
		}
		if len(stack.ServiceIds) != 0 {
			services, err := getServices(rancherClient, stack.ServiceIds)
			if err != nil {
				return nil, err
			}
			isUp := true
			for _, service := range services {
				if service.Transitioning != "no" {
					isUp = false
					break
				}
			}
			if isUp {
				return services, nil
			}
		}
		time.Sleep(time.Second)
		continue
	}
}

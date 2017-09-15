package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	dclient "github.com/docker/docker/client"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/rancher/cli/monitor"
	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/rancher-docker-api-proxy"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

var colors = []color.Attribute{color.FgGreen, color.FgBlack, color.FgBlue, color.FgCyan, color.FgMagenta, color.FgRed, color.FgWhite, color.FgYellow}

func UpCommand() cli.Command {
	return cli.Command{
		Name:   "up",
		Usage:  "Bring all services up",
		Action: rancherUp,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "pull, p",
				Usage: "Before doing the upgrade do an image pull on all hosts that have the image already",
			},
			cli.BoolFlag{
				Name:  "d",
				Usage: "Do not block and log",
			},
			cli.BoolFlag{
				Name:  "render",
				Usage: "Display processed Compose files and exit",
			},
			cli.BoolFlag{
				Name:  "upgrade, u, recreate",
				Usage: "Upgrade if service has changed",
			},
			cli.BoolFlag{
				Name:  "force-upgrade, force-recreate",
				Usage: "Upgrade regardless if service has changed",
			},
			cli.BoolFlag{
				Name:  "confirm-upgrade, c",
				Usage: "Confirm that the upgrade was success and delete old containers",
			},
			cli.BoolFlag{
				Name:  "rollback, r",
				Usage: "Rollback to the previous deployed version",
			},
			cli.IntFlag{
				Name:  "batch-size",
				Usage: "Number of containers to upgrade at once",
				Value: 2,
			},
			cli.IntFlag{
				Name:  "interval",
				Usage: "Update interval in milliseconds",
				Value: 1000,
			},
			cli.StringFlag{
				Name:  "rancher-file",
				Usage: "Specify an alternate Rancher compose file (default: rancher-compose.yml)",
			},
			cli.StringFlag{
				Name:  "env-file,e",
				Usage: "Specify a file from which to read environment variables",
			},
			cli.StringSliceFlag{
				Name:   "file,f",
				Usage:  "Specify one or more alternate compose files (default: docker-compose.yml)",
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
	compose := ""

	composeFile := ctx.String("file")
	if composeFile != "" {
		composeFile = "compose.yml"
	}
	fp, err := filepath.Abs(composeFile)
	if err != nil {
		return errors.Wrapf(err, "failed to lookup current directory name")
	}
	file, err := os.Open(fp)
	if err != nil {
		return errors.Wrapf(err, "Can not find compose.yml")
	}
	defer file.Close()
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return errors.Wrapf(err, "failed to read file")
	}
	compose = string(buf)

	//get stack name
	stackName := ""

	if ctx.String("stack") != "" {
		stackName = ctx.String("stack")
	} else {
		parent := path.Base(path.Dir(fp))
		if parent != "" && parent != "." {
			stackName = parent
		} else if wd, err := os.Getwd(); err != nil {
			return err
		} else {
			stackName = path.Base(toUnixPath(wd))
		}
	}

	stacks, err := rancherClient.Stack.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"name":         stackName,
			"removed_null": nil,
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to list stacks")
	}

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
	if !ctx.Bool("d") {
		watcher := monitor.NewUpWatcher(rancherClient)
		watcher.Subscribe()
		go func() { watcher.Start(stackName) }()
	}

	if len(stacks.Data) > 0 {
		// update stacks
		stacks.Data[0].Templates = map[string]string{
			"compose.yml": compose,
		}
		prune := ctx.Bool("prune")
		logrus.Info("Updating stack")
		_, err := rancherClient.Stack.Update(&stacks.Data[0], client.Stack{
			Templates: map[string]string{
				"compose.yml": compose,
			},
			Prune: prune,
		})
		if err != nil {
			return errors.Wrapf(err, "failed to update stack %v", stackName)
		}
	} else {
		// create new stack
		prune := ctx.Bool("prune")
		_, err := rancherClient.Stack.Create(&client.Stack{
			Name: stackName,
			Templates: map[string]string{
				"compose.yml": compose,
			},
			Prune: prune,
		})
		if err != nil {
			return errors.Wrapf(err, "failed to create stack %v", stackName)
		}
	}

	if !ctx.Bool("d") {
		for {
			stack, err := getStack(rancherClient, stackName)
			if err != nil {
				return err
			}
			if len(stack.ServiceIds) != 0 {
				instanceIds := map[string]struct{}{}
				services, err := getServices(rancherClient, stack.ServiceIds)
				if err != nil {
					return err
				}
				for _, service := range services {
					if service.Transitioning != "no" {
						logrus.Debugf("Service [%v] is not fully up", service.Name)
						time.Sleep(time.Second)
						continue
					}
					for _, instanceID := range service.InstanceIds {
						instanceIds[instanceID] = struct{}{}
					}
				}
				if err := getLogs(rancherClient, instanceIds); err != nil {
					return errors.Wrapf(err, "failed to get container logs")
				}
			}
			time.Sleep(time.Second)
		}
	}

	return nil
}

func toUnixPath(p string) string {
	return strings.Replace(p, "\\", "/", -1)
}

func getStack(c *client.RancherClient, stackName string) (client.Stack, error) {
	stacks, err := c.Stack.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"name":         stackName,
			"removed_null": nil,
		},
	})
	if err != nil {
		return client.Stack{}, errors.Wrap(err, "failed to list stacks")
	}
	if len(stacks.Data) > 0 {
		return stacks.Data[0], nil
	}
	return client.Stack{}, errors.Errorf("Failed to find stacks with name %v", stackName)
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
				if err := log(i, dockerClient); err != nil {
					logrus.Error(err)
				}
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
			if err := log(i, dockerClient); err != nil {
				logrus.Error(err)
			}
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

	scanner := bufio.NewScanner(responseBody)
	cl := getRandomColor()
	for scanner.Scan() {
		text := fmt.Sprintf("[%v]: %v\n", instance.Name, scanner.Text())
		color.New(cl).Fprint(os.Stdout, text)
	}
	return nil
}

func getRandomColor() color.Attribute {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	index := r1.Intn(8)
	return colors[index]
}

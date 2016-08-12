package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/stdcopy"
	dclient "github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/libcompose/cli/logger"
	"github.com/rancher/go-rancher/client"
	"github.com/rancher/rancher-docker-api-proxy"
	"github.com/urfave/cli"
)

var loggerFactory = logger.NewColorLoggerFactory()

func LogsCommand() cli.Command {
	return cli.Command{
		Name:        "logs",
		Usage:       "Fetch the logs of a container",
		Description: "\nExample:\n\t$ rancher logs web\n",
		ArgsUsage:   "[CONTAINERNAME CONTAINERID...]",
		HideHelp:    true,
		Action:      logsCommand,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "follow,f",
				Usage: "Follow log output",
			},
			cli.IntFlag{
				Name:  "tail",
				Value: 100,
				Usage: "Number of lines to show from the end of the logs",
			},
			cli.StringFlag{
				Name:  "since",
				Usage: "Show logs since timestamp",
			},
			//cli.BoolFlag{
			//	Name:  "details",
			//	Usage: "Show extra details provided to logs",
			//},
			cli.BoolFlag{
				Name:  "timestamps,t",
				Usage: "Show timestamps",
			},
		},
	}
}

func logsCommand(ctx *cli.Context) error {
	wg := sync.WaitGroup{}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	if len(ctx.Args()) == 0 {
		return fmt.Errorf("Please pass a container name")
	}

	instances, err := resolveContainers(c, ctx.Args())
	if err != nil {
		return err
	}

	listenSocks := map[string]*dclient.Client{}
	for _, i := range instances {
		if i.ExternalId == "" || i.HostId == "" {
			continue
		}

		if dockerClient, ok := listenSocks[i.HostId]; ok {
			wg.Add(1)
			go func(dockerClient *dclient.Client, i client.Instance) {
				doLog(len(instances) <= 1, ctx, i, dockerClient)
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
		if state != "active" {
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
			doLog(len(instances) <= 1, ctx, i, dockerClient)
			wg.Done()
		}(dockerClient, i)
	}

	wg.Wait()
	return nil
}

func doLog(single bool, ctx *cli.Context, instance client.Instance, dockerClient *dclient.Client) error {
	c, err := dockerClient.ContainerInspect(context.Background(), instance.ExternalId)
	if err != nil {
		return err
	}

	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Since:      ctx.String("since"),
		Timestamps: ctx.Bool("timestamps"),
		Follow:     ctx.Bool("follow"),
		Tail:       ctx.String("tail"),
		//Details:    ctx.Bool("details"),
	}
	responseBody, err := dockerClient.ContainerLogs(context.Background(), c.ID, options)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	if c.Config.Tty {
		_, err = io.Copy(os.Stdout, responseBody)
	} else if single {
		_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, responseBody)
	} else {
		l := loggerFactory.Create(instance.Name)
		_, err = stdcopy.StdCopy(writerFunc(l.Out), writerFunc(l.Err), responseBody)
	}
	return err
}

type writerFunc func(p []byte)

func (f writerFunc) Write(p []byte) (n int, err error) {
	f(p)
	return len(p), nil
}

func resolveContainers(c *client.RancherClient, names []string) ([]client.Instance, error) {
	result := []client.Instance{}

	for _, name := range names {
		resource, err := Lookup(c, name, "container", "service", "environment")
		if err != nil {
			return nil, err
		}
		if resource.Type == "container" {
			i, err := c.Instance.ById(resource.Id)
			if err != nil {
				return nil, err
			}
			result = append(result, *i)
		} else if resource.Type == "environment" {
			services := client.ServiceCollection{}
			err := c.GetLink(*resource, "services", &services)
			if err != nil {
				return nil, err
			}
			serviceIds := []string{}
			for _, s := range services.Data {
				serviceIds = append(serviceIds, s.Id)
			}
			instances, err := resolveContainers(c, serviceIds)
			if err != nil {
				return nil, err
			}
			result = append(result, instances...)
		} else {
			instances := client.InstanceCollection{}
			err := c.GetLink(*resource, "instances", &instances)
			if err != nil {
				return nil, err
			}
			for _, instance := range instances.Data {
				result = append(result, instance)
			}
		}
	}

	return result, nil
}

package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"errors"
	"strconv"

	"bufio"
	"bytes"
	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	dclient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/libcompose/cli/logger"
	"github.com/mitchellh/mapstructure"
	"github.com/rancher/cli/monitor"
	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/rancher-docker-api-proxy"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
	kubeutil "k8s.io/kubernetes/pkg/kubectl/util"
)

var loggerFactory = logger.NewColorLoggerFactory()

func LogsCommand() cli.Command {
	return cli.Command{
		Name:        "logs",
		Usage:       "Fetch the logs of a container",
		Description: "\nExample:\n\t$ rancher logs web\n",
		ArgsUsage:   "[CONTAINERNAME CONTAINERID...] or [SERVICENAME SERVICEID...]",
		Action:      logsCommand,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "service,s",
				Usage: "Show service logs",
			},
			cli.BoolTFlag{
				Name:  "sub-log",
				Usage: "Show service sub logs",
			},
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
			cli.BoolFlag{
				Name:  "timestamps,t",
				Usage: "Show timestamps",
			},
		},
	}
}

func printPastLogs(c *client.RancherClient, nameCache map[string]string, services map[string]bool, ctx *cli.Context) (map[string]bool, error) {
	printed := map[string]bool{}

	listOpts := defaultListOpts(nil)
	listOpts.Filters["sort"] = "id"
	listOpts.Filters["order"] = "desc"
	if !ctx.Bool("sub-log") {
		listOpts.Filters["subLog"] = "0"
	}

	limit := ctx.Int("tail")
	if limit == 0 {
		return printed, nil
	}

	if limit > 0 {
		listOpts.Filters["limit"] = limit
	}

	logs, err := c.ServiceLog.List(listOpts)
	if err != nil {
		return nil, err
	}

	for i := len(logs.Data); i > 0; i-- {
		l := logs.Data[i-1]
		printed[l.Id] = true
		printServiceLog(c, nameCache, services, l)
	}

	return printed, nil
}

func printServiceLog(c *client.RancherClient, nameCache map[string]string, services map[string]bool, log client.ServiceLog) {
	if len(services) > 0 && !services[log.ServiceId] {
		return
	}

	created, _ := time.Parse(time.RFC3339, log.Created)
	endTime, _ := time.Parse(time.RFC3339, log.EndTime)
	duration := endTime.Sub(created)
	durationStr := duration.String()
	if durationStr == "0" || strings.HasPrefix(durationStr, "-") {
		durationStr = "-"
	}
	if log.EndTime == "" {
		durationStr = "?"
	}
	if log.InstanceId == "" {
		log.InstanceId = "-"
	}

	if nameCache[log.ServiceId] == "" {
		service, err := c.Service.ById(log.ServiceId)
		if nameCache[service.StackId] == "" {
			stack, err := c.Stack.ById(service.StackId)
			if err == nil {
				nameCache[service.StackId] = stack.Name
			}
		}
		if err == nil {
			nameCache[log.ServiceId] = service.Name
		}
		nameCache[log.ServiceId] = fmt.Sprintf("%s/%s(%s)", nameCache[service.StackId], nameCache[log.ServiceId], log.ServiceId)
	}

	fmt.Printf("%s %4s %s %s %s %6s %s: %s\n", log.Created, durationStr, strings.SplitN(log.TransactionId, "-", 2)[0],
		strings.ToUpper(log.Level), nameCache[log.ServiceId], log.InstanceId, log.EventType, log.Description)
}

func serviceLogs(c *client.RancherClient, ctx *cli.Context) error {
	nameCache := map[string]string{}
	var sub *monitor.Subscription
	follow := ctx.Bool("follow")

	if follow {
		m := monitor.New(c)
		sub = m.Subscribe()
		go func() {
			logrus.Fatal(m.Start())
		}()
	}

	services, err := resolveServices(c, ctx.Args())
	if err != nil {
		return err
	}

	printed, err := printPastLogs(c, nameCache, services, ctx)
	if err != nil {
		return err
	}

	if follow {
		for event := range sub.C {
			if event.ResourceType != "serviceLog" {
				continue
			}
			if printed[event.ResourceID] {
				continue
			}
			var log client.ServiceLog
			err := mapstructure.Decode(event.Data["resource"], &log)
			if err != nil {
				logrus.Errorf("Failed to convert %#v: %v", event.Data["resource"], err)
			}
			printServiceLog(c, nameCache, services, log)
		}
	}

	return nil
}

func logsCommand(ctx *cli.Context) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	if ctx.Bool("service") {
		return serviceLogs(c, ctx)
	}

	if len(ctx.Args()) == 0 {
		return fmt.Errorf("Please pass a container name")
	}

	containers, err := resolveContainers(c, ctx.Args())
	if err != nil {
		return err
	}

	if len(containers) == 0 {
		return errors.New("No container found")
	}

	podName := containers[0].Labels[podNameLabel]
	if podName != "" {
		wg := sync.WaitGroup{}
		for _, container := range containers {
			wg.Add(1)
			go func(labels map[string]string) {
				podName = labels[podNameLabel]
				podContainerName := labels[podContainerName]
				namespace := labels[namespaceLabel]
				err := logKube(ctx, c, container, podName, namespace, podContainerName)
				if err != nil {
					logrus.Error(err)
				}
				wg.Done()
			}(container.Labels)
		}
		wg.Wait()
		return nil
	}
	return logDocker(c, ctx, containers)
}

func logKube(ctx *cli.Context, c *client.RancherClient, container client.Container, podName, namespace, podContainerName string) error {
	conf, err := constructRestConfig(ctx, c)
	if err != nil {
		return err
	}
	restClient, err := restclient.RESTClientFor(conf)
	if err != nil {
		return err
	}
	follow := ctx.Bool("follow")
	tailLines := ctx.Int("tail")
	sinceTime := ""
	if ctx.String("since") != "" {
		t, err := kubeutil.ParseRFC3339(ctx.String("since"), metav1.Now)
		if err != nil {
			return err
		}
		sinceTime = t.String()
	}
	timeStamp := ctx.Bool("timestamps")
	req := restClient.Get().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("log").
		Param("container", podContainerName)
	if follow {
		req.Param("follow", strconv.FormatBool(follow))
	}
	if tailLines != 0 {
		req.Param("tailLines", strconv.Itoa(tailLines))
	}
	if timeStamp {
		req.Param("timestamps", strconv.FormatBool(timeStamp))
	}
	if sinceTime != "" {
		req.Param("sinceTime", sinceTime)
	}
	readerCloser, err := req.Stream()
	if err != nil {
		return err
	}
	defer readerCloser.Close()

	l := loggerFactory.CreateContainerLogger(podContainerName)
	_, err = io.Copy(writerWithLineBreakFunc(l.Out), readerCloser)
	return err
}

func logDocker(c *client.RancherClient, ctx *cli.Context, containers []client.Container) error {
	wg := sync.WaitGroup{}
	listenSocks := map[string]*dclient.Client{}
	for _, i := range containers {
		if i.ExternalId == "" || i.HostId == "" {
			continue
		}

		if dockerClient, ok := listenSocks[i.HostId]; ok {
			wg.Add(1)
			go func(dockerClient *dclient.Client, i client.Container) {
				doLog(len(containers) <= 1, ctx, i, dockerClient)
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
		go func(dockerClient *dclient.Client, i client.Container) {
			doLog(len(containers) <= 1, ctx, i, dockerClient)
			wg.Done()
		}(dockerClient, i)
	}

	wg.Wait()
	return nil
}

func doLog(single bool, ctx *cli.Context, instance client.Container, dockerClient *dclient.Client) error {
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
		l := loggerFactory.CreateContainerLogger(instance.Name)
		_, err = stdcopy.StdCopy(writerFunc(l.Out), writerFunc(l.Err), responseBody)
	}
	return err
}

type writerFunc func(p []byte)

func (f writerFunc) Write(p []byte) (n int, err error) {
	f(p)
	return len(p), nil
}

type writerWithLineBreakFunc func(p []byte)

func (f writerWithLineBreakFunc) Write(p []byte) (n int, err error) {
	reader := bytes.NewReader(p)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		f(append(scanner.Bytes(), []byte("\n")...))
	}
	return len(p), nil
}

func resolveServices(c *client.RancherClient, names []string) (map[string]bool, error) {
	services := map[string]bool{}
	for _, name := range names {
		resource, err := Lookup(c, name, "service")
		if err != nil {
			return nil, err
		}
		services[resource.Id] = true
	}
	return services, nil
}

func resolveContainers(c *client.RancherClient, names []string) ([]client.Container, error) {
	result := []client.Container{}

	for _, name := range names {
		resource, err := Lookup(c, name, "container", "service")
		if err != nil {
			return nil, err
		}
		if resource.Type == "container" {
			i, err := c.Container.ById(resource.Id)
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
			containers := client.ContainerCollection{}
			err := c.GetLink(*resource, "instances", &containers)
			if err != nil {
				return nil, err
			}
			for _, instance := range containers.Data {
				result = append(result, instance)
			}
		}
	}

	return result, nil
}

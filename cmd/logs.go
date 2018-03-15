package cmd

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/cli/logger"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"github.com/rancher/cli/monitor"
	"github.com/rancher/go-rancher/v2"
	"github.com/urfave/cli"
)

type logOptions struct {
	Follow bool `json:"follow"`
	Lines  int  `json:"lines"`
}

type wsResponse struct {
	Token string `json:"token"`
	URL   string `json:"url"`
}

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
	wg := sync.WaitGroup{}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	config, err := lookupConfig(ctx)
	if err != nil {
		return err
	}

	if ctx.Bool("service") {
		return serviceLogs(c, ctx)
	}

	if len(ctx.Args()) == 0 {
		return fmt.Errorf("Please pass a container name")
	}

	instances, err := resolveContainers(c, ctx.Args())
	if err != nil {
		return err
	}

	payload := logOptions{
		Follow: ctx.Bool("follow"),
		Lines:  ctx.Int("tail"),
	}

	payloadJSON, err := json.Marshal(payload)
	if nil != err {
		return fmt.Errorf("json error: %v", err)
	}

	httpClient := http.DefaultClient

	for _, i := range instances {
		logLink, ok := i.Actions["logs"]
		if !ok {
			logrus.Debugf("no logs action found for %s", i.Name)
			continue
		}

		req, err := http.NewRequest("POST", logLink, bytes.NewBuffer(payloadJSON))
		if nil != err {
			return fmt.Errorf("request error:%v", err)
		}

		req.SetBasicAuth(config.AccessKey, config.SecretKey)

		res, err := httpClient.Do(req)
		if nil != err {
			return err
		}

		bodyBytes, err := ioutil.ReadAll(res.Body)
		if nil != err {
			return err
		}

		res.Body.Close()

		var ws wsResponse
		err = json.Unmarshal(bodyBytes, &ws)
		if nil != err {
			return err
		}

		wsURL := ws.URL + "?token=" + ws.Token

		wg.Add(1)
		go func(wsURL string) {
			doLog(wsURL)
			wg.Done()
		}(wsURL)
	}

	wg.Wait()
	return nil
}

func doLog(wsURL string) error {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{})
	if nil != err {
		logrus.Errorf("websocket dial error: %v", err)
		return err
	}
	defer conn.Close()

	for {
		msgType, buf, err := conn.ReadMessage()
		if nil != err {
			return err
		}
		var text string
		switch msgType {
		case websocket.TextMessage:
			text = string(buf)
		case websocket.BinaryMessage:
			text = bytesToFormattedHex(buf)
		default:
			return fmt.Errorf("unknown websocket frame type: %d", msgType)
		}

		fmt.Println(text)
	}
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

func resolveContainers(c *client.RancherClient, names []string) ([]client.Instance, error) {
	result := []client.Instance{}

	for _, name := range names {
		resource, err := Lookup(c, name, "container", "service")
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

func bytesToFormattedHex(bytes []byte) string {
	text := hex.EncodeToString(bytes)
	return regexp.MustCompile("(..)").ReplaceAllString(text, "$1 ")
}

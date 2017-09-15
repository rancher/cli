package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/cli/monitor"
	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/rancher-compose-executor/project/options"
	"github.com/urfave/cli"
)

var (
	waitTypes = []string{"service", "container", "host", "stack", "machine", "projectTemplate"}
)

func WaitCommand() cli.Command {
	return cli.Command{
		Name:      "wait",
		Usage:     "Wait for resources " + strings.Join(waitTypes, ", "),
		ArgsUsage: "[ID NAME...]",
		Action:    waitForResources,
		Flags:     []cli.Flag{},
	}
}

func WaitFor(ctx *cli.Context, resource string) error {
	w, err := NewWaiter(ctx)
	if err != nil {
		return err
	}
	w.Add(resource)
	return w.Wait()
}

func waitForResources(ctx *cli.Context) error {
	ctx.GlobalSet("wait", "true")

	w, err := NewWaiter(ctx)
	if err != nil {
		return err
	}

	for _, r := range ctx.Args() {
		w.Add(r)
	}

	return w.Wait()
}

func NewWaiter(ctx *cli.Context) (*Waiter, error) {
	client, err := GetClient(ctx)
	if err != nil {
		return nil, err
	}

	waitState := ctx.GlobalString("wait-state")
	if waitState == "" {
		waitState = "active"
	}

	return &Waiter{
		enabled: ctx.GlobalBool("wait"),
		timeout: ctx.GlobalInt("wait-timeout"),
		state:   waitState,
		client:  client,
	}, nil
}

type Waiter struct {
	enabled   bool
	timeout   int
	state     string
	resources []string
	client    *client.RancherClient
	monitor   *monitor.Monitor
}

type ResourceID string

func NewResourceID(resourceType, id string) ResourceID {
	return ResourceID(fmt.Sprintf("%s:%s", resourceType, id))
}

func (r ResourceID) ID() string {
	str := string(r)
	return str[strings.Index(str, ":")+1:]
}

func (r ResourceID) Type() string {
	str := string(r)
	return str[:strings.Index(str, ":")]
}

func (w *Waiter) Add(resources ...string) options.Waiter {
	for _, resource := range resources {
		fmt.Println(resource)
		w.resources = append(w.resources, resource)
	}
	return w
}

func (w *Waiter) done(resourceType, id string) (bool, error) {
	data := map[string]interface{}{}
	ok, err := w.monitor.Get(resourceType, id, &data)
	if err != nil {
		return ok, err
	}

	if ok {
		return w.checkDone(resourceType, id, data)
	}

	if err := w.client.ById(resourceType, id, &data); err != nil {
		return false, err
	}

	return w.checkDone(resourceType, id, data)
}

func (w *Waiter) checkDone(resourceType, id string, data map[string]interface{}) (bool, error) {
	transitioning := fmt.Sprint(data["transitioning"])
	logrus.Debugf("%s:%s transitioning=%s state=%v, healthState=%v waitState=%s", resourceType, id, transitioning,
		data["state"], data["healthState"], w.state)

	switch transitioning {
	case "yes":
		return false, nil
	case "error":
		return false, fmt.Errorf("%s:%s failed: %s", resourceType, id, data["transitioningMessage"])
	}

	if w.state == "" {
		return true, nil
	}

	return (data["state"] == w.state || data["healthState"] == w.state), nil
}

func (w *Waiter) Wait() error {
	if !w.enabled {
		return nil
	}

	watching := map[ResourceID]bool{}
	w.monitor = monitor.New(w.client)
	sub := w.monitor.Subscribe()
	go func() { logrus.Fatal(w.monitor.Start()) }()

	for _, resource := range w.resources {
		r, err := Lookup(w.client, resource, waitTypes...)
		if err != nil {
			return err
		}

		ok, err := w.done(r.Type, r.Id)
		if err != nil {
			return err
		}
		if !ok {
			watching[NewResourceID(r.Type, r.Id)] = true
		}
	}

	timeout := time.After(time.Duration(w.timeout) * time.Second)
	every := time.Tick(10 * time.Second)
	for len(watching) > 0 {
		var event *monitor.Event
		select {
		case event = <-sub.C:
		case <-timeout:
			return fmt.Errorf("Timeout")
		case <-every:
			for resource := range watching {
				ok, err := w.done(resource.Type(), resource.ID())
				if err != nil {
					return err
				}
				if ok {
					delete(watching, resource)
				}
			}
			continue
		}

		resource := NewResourceID(event.ResourceType, event.ResourceID)
		if !watching[resource] {
			continue
		}

		done, err := w.done(event.ResourceType, event.ResourceID)
		if err != nil {
			return err
		}

		if done {
			delete(watching, resource)
		}
	}

	return nil
}

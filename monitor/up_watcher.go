package monitor

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/rancher/go-rancher/v3"
)

type UpWatcher struct {
	sync.Mutex
	c             *client.RancherClient
	subCounter    int
	subscriptions map[int]*Subscription
}

func (m *UpWatcher) Subscribe() *Subscription {
	m.Lock()
	defer m.Unlock()

	m.subCounter++
	sub := &Subscription{
		id: m.subCounter,
		C:  make(chan *Event, 1024),
	}
	m.subscriptions[sub.id] = sub

	return sub
}

func (m *UpWatcher) Unsubscribe(sub *Subscription) {
	m.Lock()
	defer m.Unlock()

	close(sub.C)
	delete(m.subscriptions, sub.id)
}

func NewUpWatcher(c *client.RancherClient) *UpWatcher {
	return &UpWatcher{
		c:             c,
		subscriptions: map[int]*Subscription{},
	}
}

func (m *UpWatcher) Start(stackName string) error {
	schema, ok := m.c.GetSchemas().CheckSchema("subscribe")
	if !ok {
		return fmt.Errorf("Not authorized to subscribe")
	}

	urlString := schema.Links["collection"]
	u, err := url.Parse(urlString)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	}

	q := u.Query()
	q.Add("eventNames", "resource.change")
	q.Add("eventNames", "service.kubernetes.change")

	u.RawQuery = q.Encode()

	conn, resp, err := m.c.Websocket(u.String(), nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != 101 {
		return fmt.Errorf("Bad status code: %d %s", resp.StatusCode, resp.Status)
	}

	logrus.Debugf("Connected to: %s", u.String())

	return m.watch(conn, stackName)
}

func (m *UpWatcher) watch(conn *websocket.Conn, stackName string) error {
	stackID := ""
	serviceIds := map[string]struct{}{}
	lastStackMsg := ""
	lastServiceMsg := ""
	lastContainerMsg := ""
	for {
		v := Event{}
		_, r, err := conn.NextReader()
		if err != nil {
			return err
		}
		if err := json.NewDecoder(r).Decode(&v); err != nil {
			logrus.Errorf("Failed to parse json in message")
			continue
		}

		logrus.Debugf("Event: %s %s %s", v.Name, v.ResourceType, v.ResourceID)
		if v.ResourceType == "stack" {
			stackData := &client.Stack{}
			if err := unmarshalling(v.Data["resource"], stackData); err != nil {
				logrus.Errorf("failed to unmarshalling err: %v", err)
			}
			if stackData.Name == stackName {
				stackID = stackData.Id
				for _, serviceID := range stackData.ServiceIds {
					serviceIds[serviceID] = struct{}{}
				}
				switch stackData.Transitioning {
				case "yes":
					msg := fmt.Sprintf("Stack [%v]: %s", stackData.Name, stackData.TransitioningMessage)
					if msg != lastStackMsg {
						logrus.Info(msg)
					}
					lastStackMsg = msg
				}
			}
		} else if v.ResourceType == "scalingGroup" {
			serviceData := &client.Service{}
			if err := unmarshalling(v.Data["resource"], serviceData); err != nil {
				logrus.Errorf("failed to unmarshalling err: %v", err)
			}
			if serviceData.StackId == stackID {
				switch serviceData.Transitioning {
				case "yes":
					msg := fmt.Sprintf("Service [%v]: %s", serviceData.Name, serviceData.TransitioningMessage)
					if msg != lastServiceMsg {
						logrus.Info(msg)
					}
					lastServiceMsg = msg
				}
			}
		} else if v.ResourceType == "container" {
			containerData := &client.Container{}
			if err := unmarshalling(v.Data["resource"], containerData); err != nil {
				logrus.Errorf("failed to unmarshalling err: %v", err)
			}
			if containerData.StackId == stackID {
				switch containerData.Transitioning {
				case "yes":
					msg := fmt.Sprintf("Container [%v]: %s", containerData.Name, containerData.TransitioningMessage)
					if msg != lastContainerMsg {
						logrus.Info(msg)
					}
					lastContainerMsg = msg
				}
			}
		}
	}
}

func unmarshalling(data interface{}, v interface{}) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return errors.Wrapf(err, "failed to marshall object. Body: %v", data)
	}
	if err := json.Unmarshal(raw, &v); err != nil {
		return errors.Wrapf(err, "failed to unmarshall object. Body: %v", string(raw))
	}
	return nil
}

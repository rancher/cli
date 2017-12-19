package rancher

import (
	"errors"
	"io"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/libcompose/labels"
	"github.com/gorilla/websocket"
	"github.com/rancher/go-rancher/hostaccess"
	"github.com/rancher/go-rancher/v2"
	"github.com/rancher/rancher-compose-executor/config"
	"github.com/rancher/rancher-compose-executor/docker/service"
	"github.com/rancher/rancher-compose-executor/project"
	"github.com/rancher/rancher-compose-executor/project/options"
	rUtils "github.com/rancher/rancher-compose-executor/utils"
)

type Link struct {
	ServiceName, Alias string
}

type IsDone func(*client.Resource) (bool, error)

type ContainerInspect struct {
	Name       string
	Config     *container.Config
	HostConfig *container.HostConfig
}

type RancherService struct {
	name          string
	serviceConfig *config.ServiceConfig
	context       *Context
}

func (r *RancherService) Name() string {
	return r.name
}

func (r *RancherService) Config() *config.ServiceConfig {
	return r.serviceConfig
}

func (r *RancherService) Context() *Context {
	return r.context
}

func (r *RancherService) ID() string {
	service, err := r.FindExisting(r.name)
	if err != nil || service == nil {
		return ""
	}

	return service.Id
}

func NewService(name string, config *config.ServiceConfig, context *Context) *RancherService {
	return &RancherService{
		name:          name,
		serviceConfig: config,
		context:       context,
	}
}

func (r *RancherService) RancherService() (*client.Service, error) {
	return r.FindExisting(r.name)
}

func (r *RancherService) Create(ctx context.Context, options options.Create) error {
	service, err := r.FindExisting(r.name)

	if err == nil && service == nil {
		service, err = r.createService()
	} else if err == nil && service != nil {
		err = r.setupLinks(service, service.State == "inactive")
	}

	return err
}

func (r *RancherService) Up(ctx context.Context, options options.Up) error {
	return r.up(true)
}

func (r *RancherService) Build(ctx context.Context, buildOptions options.Build) error {
	return nil
}

func (r *RancherService) up(create bool) error {
	service, err := r.FindExisting(r.name)
	if err != nil {
		return err
	}

	if r.Context().Rollback {
		if service == nil {
			return nil
		}

		_, err := r.rollback(service)
		return err
	}

	if service != nil && create && r.shouldUpgrade(service) {
		if r.context.Pull {
			if err := r.Pull(context.Background()); err != nil {
				return err
			}
		}

		if service.State == "upgraded" {
			service, err = r.context.Client.Service.ActionFinishupgrade(service)
			if err != nil {
				return err
			}
			err = r.Wait(service)
			if err != nil {
				return err
			}
		}

		service, err = r.upgrade(service, r.context.ForceUpgrade, r.context.Args)
		if err != nil {
			return err
		}
	}

	if service == nil && !create {
		return nil
	}

	if service == nil {
		service, err = r.createService()
	} else {
		err = r.setupLinks(service, true)
	}

	if err != nil {
		return err
	}

	if service.State == "upgraded" && r.context.ConfirmUpgrade {
		service, err = r.context.Client.Service.ActionFinishupgrade(service)
		if err != nil {
			return err
		}
		err = r.Wait(service)
		if err != nil {
			return err
		}
	}

	if service.State == "active" {
		return nil
	}

	if service.Actions["activate"] != "" {
		service, err = r.context.Client.Service.ActionActivate(service)
		err = r.Wait(service)
	}

	// TODO: revisit whether this is the best place to perform this check
	if _, ok := r.serviceConfig.Labels["io.rancher.service.wait_for_healthcheck"]; ok {
		logrus.Debugf("Detected label io.rancher.service.wait_for_healthcheck. Polling for health")
		for {
			logrus.Debugf("Service %s has health state %s", service.Name, service.HealthState)
			if service.HealthState == "healthy" {
				break
			}
			time.Sleep(150 * time.Millisecond)
			err := r.context.Client.Reload(&service.Resource, service)
			if err != nil {
				return err
			}
		}
	}
	return err
}

func (r *RancherService) resolveServiceAndStackId(name string) (string, string, error) {
	parts := strings.SplitN(name, "/", 2)
	if len(parts) == 1 {
		return name, r.context.Stack.Id, nil
	}

	stacks, err := r.context.Client.Stack.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"name":         parts[0],
			"removed_null": nil,
		},
	})

	if err != nil {
		return "", "", err
	}

	if len(stacks.Data) == 0 {
		return "", "", nil
	}

	return parts[1], stacks.Data[0].Id, nil
}

func (r *RancherService) FindExisting(name string) (*client.Service, error) {
	logrus.Debugf("Finding service %s", name)

	name, stackId, err := r.resolveServiceAndStackId(name)
	if err != nil {
		return nil, err
	}

	if stackId == "" {
		return nil, nil
	}

	services, err := r.context.Client.Service.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"stackId":      stackId,
			"name":         name,
			"removed_null": nil,
		},
	})

	if err != nil {
		return nil, err
	}

	if len(services.Data) == 0 {
		return nil, nil
	}

	logrus.Debugf("Found service %s", name)
	return &services.Data[0], nil
}

func (r *RancherService) Metadata() map[string]interface{} {
	return rUtils.NestedMapsToMapInterface(r.serviceConfig.Metadata)
}

func (r *RancherService) HealthCheck(service string) *client.InstanceHealthCheck {
	if service == "" {
		service = r.name
	}
	if config, ok := r.context.Project.ServiceConfigs.Get(service); ok {
		return config.HealthCheck
	}
	return nil
}

func (r *RancherService) getConfiguredScale() int {
	if r.serviceConfig.Scale > 0 {
		return int(r.serviceConfig.Scale)
	}
	return 1
}

func (r *RancherService) createService() (*client.Service, error) {
	logrus.Infof("Creating service %s", r.name)

	factory, err := GetFactory(r)
	if err != nil {
		return nil, err
	}

	if err := factory.Create(r); err != nil {
		return nil, err
	}

	service, err := r.FindExisting(r.name)
	if err != nil {
		return nil, err
	}

	if err := r.setupLinks(service, true); err != nil {
		return nil, err
	}

	return service, r.Wait(service)
}

func (r *RancherService) setupLinks(service *client.Service, update bool) error {
	// Don't modify links for selector based linking, don't want to conflict
	// Don't modify links for load balancers, they're created by cattle
	serviceType := FindServiceType(r)
	if service.SelectorLink != "" || serviceType == ExternalServiceType || serviceType == LbServiceType || serviceType == LegacyLbServiceType {
		return nil
	}

	existingLinks, err := r.context.Client.ServiceConsumeMap.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"serviceId": service.Id,
		},
	})
	if err != nil {
		return err
	}

	if len(existingLinks.Data) > 0 && !update {
		return nil
	}

	links, err := r.getServiceLinks()
	_, err = r.context.Client.Service.ActionSetservicelinks(service, &client.SetServiceLinksInput{
		ServiceLinks: links,
	})
	return err
}

func (r *RancherService) SelectorContainer() string {
	return r.serviceConfig.Labels["io.rancher.service.selector.container"]
}

func (r *RancherService) SelectorLink() string {
	return r.serviceConfig.Labels["io.rancher.service.selector.link"]
}

func (r *RancherService) getServiceLinks() ([]client.ServiceLink, error) {
	links, err := r.getLinks()
	if err != nil {
		return nil, err
	}

	result := []client.ServiceLink{}
	for link, id := range links {
		if rUtils.IsRegionService(link.ServiceName) {
			result = append(result, client.ServiceLink{
				Name:    link.Alias,
				Service: link.ServiceName,
			})
		} else {
			result = append(result, client.ServiceLink{
				Name:      link.Alias,
				ServiceId: id,
			})
		}
	}

	return result, nil
}

func (r *RancherService) getLinks() (map[Link]string, error) {
	result := map[Link]string{}

	for _, link := range append(r.serviceConfig.Links, r.serviceConfig.ExternalLinks...) {
		parts := strings.SplitN(link, ":", 2)
		name := parts[0]
		alias := ""
		if len(parts) == 2 {
			alias = parts[1]
		}

		name = strings.TrimSpace(name)
		alias = strings.TrimSpace(alias)

		if rUtils.IsRegionService(name) {
			result[Link{
				ServiceName: name,
				Alias:       alias,
			}] = name
		} else {
			linkedService, err := r.FindExisting(name)
			if err != nil {
				return nil, err
			}

			if linkedService == nil {
				if _, ok := r.context.Project.ServiceConfigs.Get(name); !ok {
					logrus.Warnf("Failed to find service %s to link to", name)
				}
			} else {
				result[Link{
					ServiceName: name,
					Alias:       alias,
				}] = linkedService.Id
			}
		}
	}

	return result, nil
}

func (r *RancherService) containers() ([]client.Container, error) {
	service, err := r.FindExisting(r.name)
	if err != nil {
		return nil, err
	}

	var instances client.ContainerCollection

	err = r.context.Client.GetLink(service.Resource, "instances", &instances)
	if err != nil {
		return nil, err
	}

	return instances.Data, nil
}

func (r *RancherService) Log(ctx context.Context, follow bool) error {
	service, err := r.FindExisting(r.name)
	if err != nil || service == nil {
		return err
	}

	if service.Type != "service" {
		return nil
	}

	containers, err := r.containers()
	if err != nil {
		logrus.Errorf("Failed to list containers to log: %v", err)
		return err
	}

	for _, container := range containers {
		websocketClient := (*hostaccess.RancherWebsocketClient)(r.context.Client)
		conn, err := websocketClient.GetHostAccess(container.Resource, "logs", nil)
		if err != nil {
			logrus.Errorf("Failed to get logs for %s: %v", container.Name, err)
			continue
		}

		go r.pipeLogs(&container, conn)
	}

	return nil
}

func (r *RancherService) pipeLogs(container *client.Container, conn *websocket.Conn) {
	defer conn.Close()

	log_name := strings.TrimPrefix(container.Name, r.context.ProjectName+"_")
	logger := r.context.LoggerFactory.CreateContainerLogger(log_name)

	for {
		messageType, bytes, err := conn.ReadMessage()

		if err == io.EOF {
			return
		} else if err != nil {
			logrus.Errorf("Failed to read log: %v", err)
			return
		}

		if messageType != websocket.TextMessage || len(bytes) <= 3 {
			continue
		}

		if bytes[len(bytes)-1] != '\n' {
			bytes = append(bytes, '\n')
		}
		message := bytes[3:]

		if "01" == string(bytes[:2]) {
			logger.Out(message)
		} else {
			logger.Err(message)
		}
	}
}

func (r *RancherService) DependentServices() []project.ServiceRelationship {
	result := []project.ServiceRelationship{}

	for _, rel := range service.DefaultDependentServices(r.context.Project, r) {
		if rel.Type == project.RelTypeLink {
			rel.Optional = true
			result = append(result, rel)
		}
	}

	// Load balancers should depend on non-external target services
	lbConfig := r.serviceConfig.LbConfig
	if lbConfig != nil {
		for _, portRule := range lbConfig.PortRules {
			if portRule.Service != "" && !strings.Contains(portRule.Service, "/") {
				result = append(result, project.NewServiceRelationship(portRule.Service, project.RelTypeLink))
			}
		}
	}

	return result
}

func (r *RancherService) Client() *client.RancherClient {
	return r.context.Client
}

func (r *RancherService) pullImage(image string, labels map[string]string) error {
	taskOpts := &client.PullTask{
		Mode:   "all",
		Labels: rUtils.ToMapInterface(labels),
		Image:  image,
	}

	if r.context.PullCached {
		taskOpts.Mode = "cached"
	}

	task, err := r.context.Client.PullTask.Create(taskOpts)
	if err != nil {
		return err
	}

	printed := map[string]string{}
	lastMessage := ""
	r.WaitFor(&task.Resource, task, func() string {
		if task.TransitioningMessage != "" && task.TransitioningMessage != "In Progress" && task.TransitioningMessage != lastMessage {
			printStatus(task.Image, printed, task.Status)
			lastMessage = task.TransitioningMessage
		}

		return task.Transitioning
	})

	if task.Transitioning == "error" {
		return errors.New(task.TransitioningMessage)
	}

	if !printStatus(task.Image, printed, task.Status) {
		return errors.New("Pull failed on one of the hosts")
	}

	logrus.Infof("Finished pulling %s", task.Image)
	return nil
}

func (r *RancherService) Pull(ctx context.Context) (err error) {
	config := r.Config()
	if config.Image == "" || FindServiceType(r) != RancherType {
		return
	}

	toPull := map[string]bool{config.Image: true}
	labels := config.Labels

	if secondaries, ok := r.context.SidekickInfo.primariesToSidekicks[r.name]; ok {
		for _, secondaryName := range secondaries {
			serviceConfig, ok := r.context.Project.ServiceConfigs.Get(secondaryName)
			if !ok {
				continue
			}

			labels = rUtils.MapUnion(labels, serviceConfig.Labels)
			if serviceConfig.Image != "" {
				toPull[serviceConfig.Image] = true
			}
		}
	}

	wg := sync.WaitGroup{}

	for image := range toPull {
		wg.Add(1)
		go func(image string) {
			if pErr := r.pullImage(image, labels); pErr != nil {
				err = pErr
			}
			wg.Done()
		}(image)
	}

	wg.Wait()
	return
}

func appendHash(service *RancherService, existingLabels map[string]interface{}) (map[string]interface{}, error) {
	ret := map[string]interface{}{}
	for k, v := range existingLabels {
		ret[k] = v
	}

	hashValue := "" //, err := hash(service)
	//if err != nil {
	//return nil, err
	//}

	ret[labels.HASH.Str()] = hashValue
	return ret, nil
}

func printStatus(image string, printed map[string]string, current map[string]interface{}) bool {
	good := true
	for host, objStatus := range current {
		status, ok := objStatus.(string)
		if !ok {
			continue
		}

		v := printed[host]
		if status != "Done" {
			good = false
		}

		if v == "" {
			logrus.Infof("Checking for %s on %s...", image, host)
			v = "start"
		} else if printed[host] == "start" && status == "Done" {
			logrus.Infof("Finished %s on %s", image, host)
			v = "done"
		} else if printed[host] == "start" && status != "Pulling" && status != v {
			logrus.Infof("Checking for %s on %s: %s", image, host, status)
			v = status
		}
		printed[host] = v
	}

	return good
}

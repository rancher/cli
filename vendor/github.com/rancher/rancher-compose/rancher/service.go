package rancher

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/labels"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/events"
	"github.com/docker/libcompose/project/options"
	"github.com/gorilla/websocket"
	rancherClient "github.com/rancher/go-rancher/client"
	"github.com/rancher/go-rancher/hostaccess"
	rUtils "github.com/rancher/rancher-compose/utils"
)

type Link struct {
	ServiceName, Alias string
}

type IsDone func(*rancherClient.Resource) (bool, error)

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

func (r *RancherService) RancherConfig() RancherConfig {
	if config, ok := r.context.RancherConfig[r.name]; ok {
		return config
	}
	return RancherConfig{}
}

func NewService(name string, config *config.ServiceConfig, context *Context) *RancherService {
	return &RancherService{
		name:          name,
		serviceConfig: config,
		context:       context,
	}
}

func (r *RancherService) RancherService() (*rancherClient.Service, error) {
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

func (r *RancherService) Start(ctx context.Context) error {
	return r.up(false)
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

	return err
}

func (r *RancherService) Stop(ctx context.Context, timeout int) error {
	service, err := r.FindExisting(r.name)

	if err == nil && service == nil {
		return nil
	}

	if err != nil {
		return err
	}

	if service.State == "inactive" {
		return nil
	}

	service, err = r.context.Client.Service.ActionDeactivate(service)
	return r.Wait(service)
}

func (r *RancherService) Delete(ctx context.Context, options options.Delete) error {
	service, err := r.FindExisting(r.name)

	if err == nil && service == nil {
		return nil
	}

	if err != nil {
		return err
	}

	if service.Removed != "" || service.State == "removing" || service.State == "removed" {
		return nil
	}

	err = r.context.Client.Service.Delete(service)
	if err != nil {
		return err
	}

	return r.Wait(service)
}

func (r *RancherService) resolveServiceAndEnvironmentId(name string) (string, string, error) {
	parts := strings.SplitN(name, "/", 2)
	if len(parts) == 1 {
		return name, r.context.Environment.Id, nil
	}

	envs, err := r.context.Client.Environment.List(&rancherClient.ListOpts{
		Filters: map[string]interface{}{
			"name":         parts[0],
			"removed_null": nil,
		},
	})

	if err != nil {
		return "", "", err
	}

	if len(envs.Data) == 0 {
		return "", "", fmt.Errorf("Failed to find stack: %s", parts[0])
	}

	return parts[1], envs.Data[0].Id, nil
}

func (r *RancherService) FindExisting(name string) (*rancherClient.Service, error) {
	logrus.Debugf("Finding service %s", name)

	name, environmentId, err := r.resolveServiceAndEnvironmentId(name)
	if err != nil {
		return nil, err
	}

	services, err := r.context.Client.Service.List(&rancherClient.ListOpts{
		Filters: map[string]interface{}{
			"environmentId": environmentId,
			"name":          name,
			"removed_null":  nil,
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
	if config, ok := r.context.RancherConfig[r.name]; ok {
		return rUtils.NestedMapsToMapInterface(config.Metadata)
	}
	return map[string]interface{}{}
}

func (r *RancherService) HealthCheck(service string) *rancherClient.InstanceHealthCheck {
	if service == "" {
		service = r.name
	}
	if config, ok := r.context.RancherConfig[service]; ok {
		return config.HealthCheck
	}

	return nil
}

func (r *RancherService) getConfiguredScale() int {
	scale := 1
	if config, ok := r.context.RancherConfig[r.name]; ok {
		if config.Scale > 0 {
			scale = config.Scale
		}
	}

	return scale
}

func (r *RancherService) createService() (*rancherClient.Service, error) {
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

	err = r.Wait(service)
	return service, err
}

func (r *RancherService) setupLinks(service *rancherClient.Service, update bool) error {
	// Don't modify links for selector based linking, don't want to conflict
	if service.SelectorLink != "" || FindServiceType(r) == ExternalServiceType {
		return nil
	}

	var err error
	var links []interface{}

	existingLinks, err := r.context.Client.ServiceConsumeMap.List(&rancherClient.ListOpts{
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

	if service.Type == rancherClient.LOAD_BALANCER_SERVICE_TYPE {
		links, err = r.getLbLinks()
	} else {
		links, err = r.getServiceLinks()
	}

	if err == nil {
		_, err = r.context.Client.Service.ActionSetservicelinks(service, &rancherClient.SetServiceLinksInput{
			ServiceLinks: links,
		})
	}
	return err
}

func (r *RancherService) getLbLinks() ([]interface{}, error) {
	links, err := r.getLinks()
	if err != nil {
		return nil, err
	}

	result := []interface{}{}
	for link, id := range links {
		ports, err := r.getLbLinkPorts(link.ServiceName)
		if err != nil {
			return nil, err
		}

		result = append(result, rancherClient.LoadBalancerServiceLink{
			Ports:     ports,
			ServiceId: id,
		})
	}

	return result, nil
}

func (r *RancherService) SelectorContainer() string {
	return r.serviceConfig.Labels["io.rancher.service.selector.container"]
}

func (r *RancherService) SelectorLink() string {
	return r.serviceConfig.Labels["io.rancher.service.selector.link"]
}

func (r *RancherService) getLbLinkPorts(name string) ([]string, error) {
	labelName := "io.rancher.loadbalancer.target." + name
	v := r.serviceConfig.Labels[labelName]
	if v == "" {
		return []string{}, nil
	}

	return rUtils.TrimSplit(v, ",", -1), nil
}

func (r *RancherService) getServiceLinks() ([]interface{}, error) {
	links, err := r.getLinks()
	if err != nil {
		return nil, err
	}

	result := []interface{}{}
	for link, id := range links {
		result = append(result, rancherClient.ServiceLink{
			Name:      link.Alias,
			ServiceId: id,
		})
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

	return result, nil
}

func (r *RancherService) Scale(ctx context.Context, count int, timeout int) error {
	service, err := r.FindExisting(r.name)
	if err != nil {
		return err
	}

	if service == nil {
		return fmt.Errorf("Failed to find %s to scale", r.name)
	}

	service, err = r.context.Client.Service.Update(service, map[string]interface{}{
		"scale": count,
	})
	if err != nil {
		return err
	}

	return r.Wait(service)
}

func (r *RancherService) Containers(ctx context.Context) ([]project.Container, error) {
	result := []project.Container{}

	containers, err := r.containers()
	if err != nil {
		return nil, err
	}

	for _, c := range containers {
		name := c.Name
		if name == "" {
			name = c.Uuid
		}
		result = append(result, NewContainer(c.Id, name))
	}

	return result, nil
}

func (r *RancherService) containers() ([]rancherClient.Container, error) {
	service, err := r.FindExisting(r.name)
	if err != nil {
		return nil, err
	}

	var instances rancherClient.ContainerCollection

	err = r.context.Client.GetLink(service.Resource, "instances", &instances)
	if err != nil {
		return nil, err
	}

	return instances.Data, nil
}

func (r *RancherService) Restart(ctx context.Context, timeout int) error {
	service, err := r.FindExisting(r.name)
	if err != nil {
		return err
	}

	service, err = r.context.Client.Service.ActionRestart(service, &rancherClient.ServiceRestart{
		RollingRestartStrategy: rancherClient.RollingRestartStrategy{
			BatchSize:      r.context.BatchSize,
			IntervalMillis: r.context.Interval,
		},
	})

	if err != nil {
		logrus.Errorf("Failed to restart %s: %v", r.Name(), err)
		return err
	}

	return r.Wait(service)
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
		conn, err := (*hostaccess.RancherWebsocketClient)(r.context.Client).GetHostAccess(container.Resource, "logs", nil)
		if err != nil {
			logrus.Errorf("Failed to get logs for %s: %v", container.Name, err)
			continue
		}

		go r.pipeLogs(&container, conn)
	}

	return nil
}

func (r *RancherService) pipeLogs(container *rancherClient.Container, conn *websocket.Conn) {
	defer conn.Close()

	log_name := strings.TrimPrefix(container.Name, r.context.ProjectName+"_")
	logger := r.context.LoggerFactory.Create(log_name)

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

	for _, rel := range project.DefaultDependentServices(r.context.Project, r) {
		if rel.Type == project.RelTypeLink {
			rel.Optional = true
			result = append(result, rel)
		}
	}

	return result
}

func (r *RancherService) Client() *rancherClient.RancherClient {
	return r.context.Client
}

func (r *RancherService) Kill(ctx context.Context, signal string) error {
	return project.ErrUnsupported
}

func (r *RancherService) Info(ctx context.Context, qFlag bool) (project.InfoSet, error) {
	return project.InfoSet{}, project.ErrUnsupported
}

func (r *RancherService) pullImage(image string, labels map[string]string) error {
	taskOpts := &rancherClient.PullTask{
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

func (r *RancherService) Pause(ctx context.Context) error {
	return project.ErrUnsupported
}

func (r *RancherService) Unpause(ctx context.Context) error {
	return project.ErrUnsupported
}

func (r *RancherService) Down() error {
	return project.ErrUnsupported
}

func (r *RancherService) Events(ctx context.Context, messages chan events.ContainerEvent) error {
	return project.ErrUnsupported
}

func (r *RancherService) Run(ctx context.Context, commandParts []string, options options.Run) (int, error) {
	return 0, project.ErrUnsupported
}

func (r *RancherService) RemoveImage(ctx context.Context, imageType options.ImageType) error {
	return project.ErrUnsupported
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

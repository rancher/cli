package project

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/logger"
	"github.com/docker/libcompose/utils"
	"github.com/rancher/rancher-compose-executor/config"
	"github.com/rancher/rancher-compose-executor/lookup"
	"github.com/rancher/rancher-compose-executor/project/events"
	"github.com/rancher/rancher-compose-executor/template"
)

type wrapperAction func(*serviceWrapper, map[string]*serviceWrapper)
type serviceAction func(service Service) error

// Project holds libcompose project information.
type Project struct {
	Name              string
	ServiceConfigs    *config.ServiceConfigs
	ContainerConfigs  *config.ServiceConfigs
	DependencyConfigs map[string]*config.DependencyConfig
	VolumeConfigs     map[string]*config.VolumeConfig
	NetworkConfigs    map[string]*config.NetworkConfig
	SecretConfigs     map[string]*config.SecretConfig
	HostConfigs       map[string]*config.HostConfig
	Files             []string
	ReloadCallback    func() error

	dependencies Dependencies
	volumes      Volumes
	secrets      Secrets
	hosts        Hosts
	context      *Context
	reload       []string
	listeners    []chan<- events.Event
	hasListeners bool
}

// NewProject creates a new project with the specified context.
func NewProject(context *Context) *Project {
	p := &Project{
		context:           context,
		ServiceConfigs:    config.NewServiceConfigs(),
		ContainerConfigs:  config.NewServiceConfigs(),
		DependencyConfigs: make(map[string]*config.DependencyConfig),
		VolumeConfigs:     make(map[string]*config.VolumeConfig),
		NetworkConfigs:    make(map[string]*config.NetworkConfig),
		SecretConfigs:     make(map[string]*config.SecretConfig),
		HostConfigs:       make(map[string]*config.HostConfig),
	}

	if context.LoggerFactory == nil {
		context.LoggerFactory = &logger.NullLogger{}
	}

	if context.ResourceLookup == nil {
		context.ResourceLookup = &lookup.FileResourceLookup{}
	}

	context.Project = p

	p.listeners = []chan<- events.Event{NewDefaultListener(p)}

	return p
}

func (p *Project) Open() error {
	return p.context.open()
}

func (p *Project) Parse() error {
	err := p.Open()
	if err != nil {
		return err
	}

	p.Name = p.context.ProjectName

	p.Files = p.context.ComposeFiles

	if len(p.Files) == 1 && p.Files[0] == "-" {
		p.Files = []string{"."}
	}

	if p.context.ComposeBytes != nil {
		for i, composeBytes := range p.context.ComposeBytes {
			file := ""
			if i < len(p.context.ComposeFiles) {
				file = p.Files[i]
			}
			if err := p.load(file, composeBytes); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Project) CreateService(name string) (Service, error) {
	factory := p.context.ServiceFactory
	existing, ok := p.ServiceConfigs.Get(name)
	if !ok {
		factory = p.context.ContainerFactory
		existing, ok = p.ContainerConfigs.Get(name)
		if !ok {
			return nil, fmt.Errorf("Failed to find service or container: %s", name)
		}
	}

	// Copy because we are about to modify the environment
	config := *existing

	// TODO: perform this transformation in config package
	if p.context.EnvironmentLookup != nil {
		parsedEnv := make([]string, 0, len(config.Environment))

		for _, env := range config.Environment {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) > 1 {
				parsedEnv = append(parsedEnv, env)
				continue
			} else {
				env = parts[0]
			}

			for _, value := range p.context.EnvironmentLookup.Lookup(env, &config) {
				parsedEnv = append(parsedEnv, value)
			}
		}

		config.Environment = parsedEnv

		// check the environment for extra build Args that are set but not given a value in the compose file
		for arg, value := range config.Build.Args {
			if value == "\x00" {
				envValue := p.context.EnvironmentLookup.Lookup(arg, &config)
				// depending on what we get back we do different things
				switch l := len(envValue); l {
				case 0:
					delete(config.Build.Args, arg)
				case 1:
					parts := strings.SplitN(envValue[0], "=", 2)
					config.Build.Args[parts[0]] = parts[1]
				default:
					return nil, fmt.Errorf("Tried to set Build Arg %#v to multi-value %#v.", arg, envValue)
				}
			}
		}
	}

	return factory.Create(p, name, &config)
}

func (p *Project) load(file string, bytes []byte) error {
	config, err := config.Merge(p.ServiceConfigs, p.context.EnvironmentLookup, p.context.ResourceLookup, template.StackInfo{
		Version:         p.context.Version,
		PreviousVersion: p.context.PreviousVersion,
		Name:            p.Name,
	}, file, bytes)
	if err != nil {
		log.Errorf("Could not parse config for project %s : %v", p.Name, err)
		return err
	}

	for name, config := range config.Services {
		p.ServiceConfigs.Add(name, config)
		p.reload = append(p.reload, name)
	}
	for name, config := range config.Containers {
		p.ContainerConfigs.Add(name, config)
		p.reload = append(p.reload, name)
	}

	for name, config := range config.Dependencies {
		p.DependencyConfigs[name] = config
	}
	for name, config := range config.Volumes {
		p.VolumeConfigs[name] = config
	}
	for name, config := range config.Networks {
		p.NetworkConfigs[name] = config
	}
	for name, config := range config.Secrets {
		p.SecretConfigs[name] = config
	}
	for name, config := range config.Hosts {
		p.HostConfigs[name] = config
	}

	if p.context.DependenciesFactory != nil {
		dependencies, err := p.context.DependenciesFactory.Create(p.Name, p.DependencyConfigs)
		if err != nil {
			return err
		}
		p.dependencies = dependencies
	}
	if p.context.VolumesFactory != nil {
		volumes, err := p.context.VolumesFactory.Create(p.Name, p.VolumeConfigs, p.ServiceConfigs)
		if err != nil {
			return err
		}
		p.volumes = volumes
	}
	if p.context.SecretsFactory != nil {
		secrets, err := p.context.SecretsFactory.Create(p.Name, p.SecretConfigs)
		if err != nil {
			return err
		}
		p.secrets = secrets
	}
	if p.context.HostsFactory != nil {
		hosts, err := p.context.HostsFactory.Create(p.Name, p.HostConfigs)
		if err != nil {
			return err
		}
		p.hosts = hosts
	}

	return nil
}

func (p *Project) initialize(ctx context.Context) error {
	if p.dependencies != nil {
		if err := p.dependencies.Initialize(ctx); err != nil {
			return err
		}
	}
	if p.volumes != nil {
		if err := p.volumes.Initialize(ctx); err != nil {
			return err
		}
	}
	if p.secrets != nil {
		if err := p.secrets.Initialize(ctx); err != nil {
			return err
		}
	}
	if p.hosts != nil {
		if err := p.hosts.Initialize(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (p *Project) loadWrappers(wrappers map[string]*serviceWrapper, servicesToConstruct []string) error {
	for _, name := range servicesToConstruct {
		wrapper, err := newServiceWrapper(name, p)
		if err != nil {
			return err
		}
		wrappers[name] = wrapper
	}

	return nil
}

func (p *Project) perform(start, done events.EventType, services []string, action wrapperAction, cycleAction serviceAction) error {
	p.Notify(start, "", nil)

	err := p.forEach(services, action, cycleAction)

	p.Notify(done, "", nil)
	return err
}

func isSelected(wrapper *serviceWrapper, selected map[string]bool) bool {
	return len(selected) == 0 || selected[wrapper.name]
}

func (p *Project) forEach(services []string, action wrapperAction, cycleAction serviceAction) error {
	selected := make(map[string]bool)
	wrappers := make(map[string]*serviceWrapper)

	for _, s := range services {
		selected[s] = true
	}

	return p.traverse(true, selected, wrappers, action, cycleAction)
}

func (p *Project) startService(wrappers map[string]*serviceWrapper, history []string, selected, launched map[string]bool, wrapper *serviceWrapper, action wrapperAction, cycleAction serviceAction) error {
	if launched[wrapper.name] {
		return nil
	}

	launched[wrapper.name] = true
	history = append(history, wrapper.name)

	for _, dep := range wrapper.service.DependentServices() {
		target := wrappers[dep.Target]
		if target == nil {
			log.Debugf("Failed to find %s", dep.Target)
			return fmt.Errorf("Service '%s' has a link to service '%s' which is undefined", wrapper.name, dep.Target)
		}

		if utils.Contains(history, dep.Target) {
			cycle := strings.Join(append(history, dep.Target), "->")
			if dep.Optional {
				log.Debugf("Ignoring cycle for %s", cycle)
				wrapper.IgnoreDep(dep.Target)
				if cycleAction != nil {
					var err error
					log.Debugf("Running cycle action for %s", cycle)
					err = cycleAction(target.service)
					if err != nil {
						return err
					}
				}
			} else {
				return fmt.Errorf("Cycle detected in path %s", cycle)
			}

			continue
		}

		err := p.startService(wrappers, history, selected, launched, target, action, cycleAction)
		if err != nil {
			return err
		}
	}

	if isSelected(wrapper, selected) {
		log.Debugf("Launching action for %s", wrapper.name)
		go action(wrapper, wrappers)
	} else {
		wrapper.Ignore()
	}

	return nil
}

func (p *Project) traverse(start bool, selected map[string]bool, wrappers map[string]*serviceWrapper, action wrapperAction, cycleAction serviceAction) error {
	restart := false
	wrapperList := []string{}

	if start {
		for _, name := range p.ServiceConfigs.Keys() {
			wrapperList = append(wrapperList, name)
		}
		for _, name := range p.ContainerConfigs.Keys() {
			wrapperList = append(wrapperList, name)
		}
	} else {
		for _, wrapper := range wrappers {
			if err := wrapper.Reset(); err != nil {
				return err
			}
		}
		wrapperList = p.reload
	}

	p.loadWrappers(wrappers, wrapperList)
	p.reload = []string{}

	// check service name
	for s := range selected {
		if wrappers[s] == nil {
			return errors.New("No such service: " + s)
		}
	}

	launched := map[string]bool{}

	for _, wrapper := range wrappers {
		if err := p.startService(wrappers, []string{}, selected, launched, wrapper, action, cycleAction); err != nil {
			return err
		}
	}

	var firstError error

	for _, wrapper := range wrappers {
		if !isSelected(wrapper, selected) {
			continue
		}
		if err := wrapper.Wait(); err == ErrRestart {
			restart = true
		} else if err != nil {
			log.Errorf("Failed to start: %s : %v", wrapper.name, err)
			if firstError == nil {
				firstError = err
			}
		}
	}

	if restart {
		if p.ReloadCallback != nil {
			if err := p.ReloadCallback(); err != nil {
				log.Errorf("Failed calling callback: %v", err)
			}
		}
		return p.traverse(false, selected, wrappers, action, cycleAction)
	}
	return firstError
}

// AddListener adds the specified listener to the project.
// This implements implicitly events.Emitter.
func (p *Project) AddListener(c chan<- events.Event) {
	if !p.hasListeners {
		for _, l := range p.listeners {
			close(l)
		}
		p.hasListeners = true
		p.listeners = []chan<- events.Event{c}
	} else {
		p.listeners = append(p.listeners, c)
	}
}

// Notify notifies all project listener with the specified eventType, service name and datas.
// This implements implicitly events.Notifier interface.
func (p *Project) Notify(eventType events.EventType, serviceName string, data map[string]string) {
	if eventType == events.NoEvent {
		return
	}

	event := events.Event{
		EventType:   eventType,
		ServiceName: serviceName,
		Data:        data,
	}

	for _, l := range p.listeners {
		l <- event
	}
}

// IsNamedVolume returns whether the specified volume (string) is a named volume or not.
func IsNamedVolume(volume string) bool {
	return !strings.HasPrefix(volume, ".") && !strings.HasPrefix(volume, "/") && !strings.HasPrefix(volume, "~")
}

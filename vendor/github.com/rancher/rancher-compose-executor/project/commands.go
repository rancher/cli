package project

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/rancher/rancher-compose-executor/project/events"
	"github.com/rancher/rancher-compose-executor/project/options"
	"github.com/rancher/rancher-compose-executor/template"
)

func (p *Project) Build(ctx context.Context, buildOptions options.Build, services ...string) error {
	return p.perform(events.ProjectBuildStart, events.ProjectBuildDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, events.ServiceBuildStart, events.ServiceBuild, func(service Service) error {
			return service.Build(ctx, buildOptions)
		})
	}), nil)
}

func (p *Project) Create(ctx context.Context, options options.Create, services ...string) error {
	if options.NoRecreate && options.ForceRecreate {
		return fmt.Errorf("no-recreate and force-recreate cannot be combined")
	}
	if err := p.initialize(ctx); err != nil {
		return err
	}
	return p.perform(events.ProjectCreateStart, events.ProjectCreateDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, events.ServiceCreateStart, events.ServiceCreate, func(service Service) error {
			return service.Create(ctx, options)
		})
	}), nil)
}

func (p *Project) Log(ctx context.Context, follow bool, services ...string) error {
	return p.forEach(services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, events.NoEvent, events.NoEvent, func(service Service) error {
			return service.Log(ctx, follow)
		})
	}), nil)
}

func (p *Project) Up(ctx context.Context, options options.Up, services ...string) error {
	if err := p.initialize(ctx); err != nil {
		return err
	}
	return p.perform(events.ProjectUpStart, events.ProjectUpDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, events.ServiceUpStart, events.ServiceUp, func(service Service) error {
			return service.Up(ctx, options)
		})
	}), func(service Service) error {
		return service.Create(ctx, options.Create)
	})
}

func (p *Project) Render() ([][]byte, error) {
	var renderedComposeBytes [][]byte
	for _, contents := range p.context.ComposeBytes {
		// TODO: figure out story for release variables when using CLI
		contents, err := template.Apply(contents, template.StackInfo{Name: p.Name}, p.context.EnvironmentLookup.Variables())
		if err != nil {
			return nil, err
		}
		renderedComposeBytes = append(renderedComposeBytes, contents)
	}
	return renderedComposeBytes, nil
}

package project

import (
	"golang.org/x/net/context"

	"github.com/rancher/rancher-compose-executor/project/events"
	"github.com/rancher/rancher-compose-executor/project/options"
)

// Up creates and starts the specified services (kinda like docker run).
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

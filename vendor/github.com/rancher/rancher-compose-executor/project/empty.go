package project

import (
	"golang.org/x/net/context"

	"github.com/rancher/rancher-compose-executor/config"
	"github.com/rancher/rancher-compose-executor/project/options"
)

// this ensures EmptyService implements Service
// useful since it's easy to forget adding new functions to EmptyService
var _ Service = (*EmptyService)(nil)

// EmptyService is a struct that implements Service but does nothing.
type EmptyService struct {
}

// Create implements Service.Create but does nothing.
func (e *EmptyService) Create(ctx context.Context, options options.Create) error {
	return nil
}

// Build implements Service.Build but does nothing.
func (e *EmptyService) Build(ctx context.Context, buildOptions options.Build) error {
	return nil
}

// Up implements Service.Up but does nothing.
func (e *EmptyService) Up(ctx context.Context, options options.Up) error {
	return nil
}

// Log implements Service.Log but does nothing.
func (e *EmptyService) Log(ctx context.Context, follow bool) error {
	return nil
}

// DependentServices implements Service.DependentServices with empty slice.
func (e *EmptyService) DependentServices() []ServiceRelationship {
	return []ServiceRelationship{}
}

// Config implements Service.Config with empty config.
func (e *EmptyService) Config() *config.ServiceConfig {
	return &config.ServiceConfig{}
}

// Name implements Service.Name with empty name.
func (e *EmptyService) Name() string {
	return ""
}

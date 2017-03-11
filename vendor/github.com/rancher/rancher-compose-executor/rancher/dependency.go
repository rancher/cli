package rancher

import (
	"golang.org/x/net/context"

	"github.com/rancher/rancher-compose-executor/config"
	"github.com/rancher/rancher-compose-executor/project"
)

type RancherDependenciesFactory struct {
	Context *Context
}

func (f *RancherDependenciesFactory) Create(projectName string, dependencyConfigs map[string]*config.DependencyConfig) (project.Dependencies, error) {
	dependencies := make([]*Dependency, 0, len(dependencyConfigs))
	for name, config := range dependencyConfigs {
		dependencies = append(dependencies, &Dependency{
			context:     f.Context,
			name:        name,
			projectName: projectName,
			template:    config.Template,
			version:     config.Version,
		})
	}
	return &Dependencies{
		dependencies: dependencies,
	}, nil
}

type Dependencies struct {
	dependencies []*Dependency
	Context      *Context
}

func (h *Dependencies) Initialize(ctx context.Context) error {
	for _, dependency := range h.dependencies {
		if err := dependency.EnsureItExists(ctx); err != nil {
			return err
		}
	}
	return nil
}

type Dependency struct {
	context     *Context
	name        string
	projectName string
	template    string
	version     string
}

func (d *Dependency) EnsureItExists(ctx context.Context) error {
	return nil
}

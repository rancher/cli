package project

import (
	"errors"

	"golang.org/x/net/context"

	"github.com/rancher/rancher-compose-executor/config"
	"github.com/rancher/rancher-compose-executor/project/options"
)

// Service defines what a libcompose service provides.
type Service interface {
	Build(ctx context.Context, buildOptions options.Build) error
	Create(ctx context.Context, options options.Create) error
	Log(ctx context.Context, follow bool) error
	Up(ctx context.Context, options options.Up) error

	DependentServices() []ServiceRelationship
	Config() *config.ServiceConfig
	Name() string
	ID() string
}

// ServiceState holds the state of a service.
type ServiceState string

// State definitions
var (
	StateExecuted = ServiceState("executed")
	StateUnknown  = ServiceState("unknown")
)

// Error definitions
var (
	ErrRestart     = errors.New("Restart execution")
	ErrUnsupported = errors.New("UnsupportedOperation")
)

// ServiceFactory is an interface factory to create Service object for the specified
// project, with the specified name and service configuration.
type ServiceFactory interface {
	Create(project *Project, name string, serviceConfig *config.ServiceConfig) (Service, error)
}

// ServiceRelationshipType defines the type of service relationship.
type ServiceRelationshipType string

// RelTypeLink means the services are linked (docker links).
const RelTypeLink = ServiceRelationshipType("")

// RelTypeNetNamespace means the services share the same network namespace.
const RelTypeNetNamespace = ServiceRelationshipType("netns")

// RelTypeIpcNamespace means the service share the same ipc namespace.
const RelTypeIpcNamespace = ServiceRelationshipType("ipc")

// RelTypeVolumesFrom means the services share some volumes.
const RelTypeVolumesFrom = ServiceRelationshipType("volumesFrom")

// RelTypeDependsOn means the dependency was explicitly set using 'depends_on'.
const RelTypeDependsOn = ServiceRelationshipType("dependsOn")

// RelTypeNetworkMode means the services depends on another service on networkMode
const RelTypeNetworkMode = ServiceRelationshipType("networkMode")

// ServiceRelationship holds the relationship information between two services.
type ServiceRelationship struct {
	Target, Alias string
	Type          ServiceRelationshipType
	Optional      bool
}

// NewServiceRelationship creates a new Relationship based on the specified alias
// and relationship type.
func NewServiceRelationship(nameAlias string, relType ServiceRelationshipType) ServiceRelationship {
	name, alias := NameAlias(nameAlias)
	return ServiceRelationship{
		Target: name,
		Alias:  alias,
		Type:   relType,
	}
}

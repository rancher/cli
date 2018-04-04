package client

const (
	InternalNodeStatusType                   = "internalNodeStatus"
	InternalNodeStatusFieldAllocatable       = "allocatable"
	InternalNodeStatusFieldCapacity          = "capacity"
	InternalNodeStatusFieldExternalIPAddress = "externalIpAddress"
	InternalNodeStatusFieldHostname          = "hostname"
	InternalNodeStatusFieldIPAddress         = "ipAddress"
	InternalNodeStatusFieldInfo              = "info"
	InternalNodeStatusFieldNodeConditions    = "nodeConditions"
	InternalNodeStatusFieldVolumesAttached   = "volumesAttached"
	InternalNodeStatusFieldVolumesInUse      = "volumesInUse"
)

type InternalNodeStatus struct {
	Allocatable       map[string]string         `json:"allocatable,omitempty" yaml:"allocatable,omitempty"`
	Capacity          map[string]string         `json:"capacity,omitempty" yaml:"capacity,omitempty"`
	ExternalIPAddress string                    `json:"externalIpAddress,omitempty" yaml:"externalIpAddress,omitempty"`
	Hostname          string                    `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	IPAddress         string                    `json:"ipAddress,omitempty" yaml:"ipAddress,omitempty"`
	Info              *NodeInfo                 `json:"info,omitempty" yaml:"info,omitempty"`
	NodeConditions    []NodeCondition           `json:"nodeConditions,omitempty" yaml:"nodeConditions,omitempty"`
	VolumesAttached   map[string]AttachedVolume `json:"volumesAttached,omitempty" yaml:"volumesAttached,omitempty"`
	VolumesInUse      []string                  `json:"volumesInUse,omitempty" yaml:"volumesInUse,omitempty"`
}

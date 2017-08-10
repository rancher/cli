package convert

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/blkiodev"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/runconfig/opts"
	"github.com/docker/go-connections/nat"
	"github.com/docker/go-units"
	"github.com/docker/libcompose/utils"
	"github.com/rancher/rancher-compose-executor/config"
	"github.com/rancher/rancher-compose-executor/project"
	"github.com/rancher/rancher-compose-executor/yaml"
)

// ConfigWrapper wraps Config, HostConfig and NetworkingConfig for a container.
type ConfigWrapper struct {
	Config           *container.Config
	HostConfig       *container.HostConfig
	NetworkingConfig *network.NetworkingConfig
}

// Filter filters the specified string slice with the specified function.
func Filter(vs []string, f func(string) bool) []string {
	r := make([]string, 0, len(vs))
	for _, v := range vs {
		if f(v) {
			r = append(r, v)
		}
	}
	return r
}

func toMap(vs []string) map[string]struct{} {
	m := map[string]struct{}{}
	for _, v := range vs {
		if v != "" {
			m[v] = struct{}{}
		}
	}
	return m
}

func isBind(s string) bool {
	return strings.ContainsRune(s, ':')
}

func isVolume(s string) bool {
	return !isBind(s)
}

// ConvertToAPI converts a service configuration to a docker API container configuration.
func ConvertToAPI(serviceConfig *config.ServiceConfig, ctx project.Context) (*ConfigWrapper, error) {
	config, hostConfig, err := Convert(serviceConfig, ctx)
	if err != nil {
		return nil, err
	}

	result := ConfigWrapper{
		Config:     config,
		HostConfig: hostConfig,
	}
	return &result, nil
}

func volumes(c *config.ServiceConfig, ctx project.Context) []string {
	if c.Volumes == nil {
		return []string{}
	}
	volumes := make([]string, len(c.Volumes.Volumes))
	for _, v := range c.Volumes.Volumes {
		vol := v
		if len(ctx.ComposeFiles) > 0 && !project.IsNamedVolume(v.Source) {
			sourceVol := ctx.ResourceLookup.ResolvePath(v.String(), ctx.ComposeFiles[0])
			vol.Source = strings.SplitN(sourceVol, ":", 2)[0]
		}
		volumes = append(volumes, vol.String())
	}
	return volumes
}

func restartPolicy(c *config.ServiceConfig) (*container.RestartPolicy, error) {
	restart, err := opts.ParseRestartPolicy(c.Restart)
	if err != nil {
		return nil, err
	}
	return &container.RestartPolicy{Name: restart.Name, MaximumRetryCount: restart.MaximumRetryCount}, nil
}

func ports(c *config.ServiceConfig) (map[nat.Port]struct{}, nat.PortMap, error) {
	ports, binding, err := nat.ParsePortSpecs(c.Ports)
	if err != nil {
		return nil, nil, err
	}

	exPorts, _, err := nat.ParsePortSpecs(c.Expose)
	if err != nil {
		return nil, nil, err
	}

	for k, v := range exPorts {
		ports[k] = v
	}

	exposedPorts := map[nat.Port]struct{}{}
	for k, v := range ports {
		exposedPorts[nat.Port(k)] = v
	}

	portBindings := nat.PortMap{}
	for k, bv := range binding {
		dcbs := make([]nat.PortBinding, len(bv))
		for k, v := range bv {
			dcbs[k] = nat.PortBinding{HostIP: v.HostIP, HostPort: v.HostPort}
		}
		portBindings[nat.Port(k)] = dcbs
	}
	return exposedPorts, portBindings, nil
}

// Convert converts a service configuration to an docker API structures (Config and HostConfig)
func Convert(c *config.ServiceConfig, ctx project.Context) (*container.Config, *container.HostConfig, error) {
	restartPolicy, err := restartPolicy(c)
	if err != nil {
		return nil, nil, err
	}

	exposedPorts, portBindings, err := ports(c)
	if err != nil {
		return nil, nil, err
	}

	deviceMappings, err := parseDevices(c.Devices)
	if err != nil {
		return nil, nil, err
	}

	var volumesFrom []string
	if c.VolumesFrom != nil {
		volumesFrom, err = getVolumesFrom(c.VolumesFrom, ctx.Project.ServiceConfigs, ctx.ProjectName)
		if err != nil {
			return nil, nil, err
		}
	}

	vols := volumes(c, ctx)

	config := &container.Config{
		Entrypoint:   strslice.StrSlice(utils.CopySlice(c.Entrypoint)),
		Hostname:     c.Hostname,
		Domainname:   c.DomainName,
		User:         c.User,
		Env:          utils.CopySlice(c.Environment),
		Cmd:          strslice.StrSlice(utils.CopySlice(c.Command)),
		Image:        c.Image,
		Labels:       utils.CopyMap(c.Labels),
		ExposedPorts: exposedPorts,
		Tty:          c.Tty,
		OpenStdin:    c.StdinOpen,
		WorkingDir:   c.WorkingDir,
		Volumes:      toMap(Filter(vols, isVolume)),
		MacAddress:   c.MacAddress,
		StopSignal:   c.StopSignal,
		StopTimeout:  &[]int64{int64(c.StopGracePeriod)}[0],
	}

	ulimits := []*units.Ulimit{}
	if c.Ulimits.Elements != nil {
		for _, ulimit := range c.Ulimits.Elements {
			ulimits = append(ulimits, &units.Ulimit{
				Name: ulimit.Name,
				Soft: ulimit.Soft,
				Hard: ulimit.Hard,
			})
		}
	}

	memorySwappiness := int64(c.MemSwappiness)

	tmpfs := map[string]string{}
	for _, path := range c.Tmpfs {
		split := strings.SplitN(path, ":", 2)
		if len(split) == 1 {
			tmpfs[split[0]] = ""
		} else if len(split) == 2 {
			tmpfs[split[0]] = split[1]
		}
	}

	blkioWeightDevices := []*blkiodev.WeightDevice{}
	for _, blkioWeightDevice := range c.BlkioWeightDevice {
		split := strings.Split(blkioWeightDevice, ":")
		if len(split) == 2 {
			weight, err := strconv.ParseUint(split[1], 10, 16)
			if err != nil {
				return nil, nil, err
			}
			blkioWeightDevices = append(blkioWeightDevices, &blkiodev.WeightDevice{
				Path:   split[0],
				Weight: uint16(weight),
			})
		}
	}

	blkioDeviceReadBps, err := getThrottleDevice(c.DeviceReadBps)
	if err != nil {
		return nil, nil, err
	}

	blkioDeviceReadIOps, err := getThrottleDevice(c.DeviceReadIOps)
	if err != nil {
		return nil, nil, err
	}

	blkioDeviceWriteBps, err := getThrottleDevice(c.DeviceWriteBps)
	if err != nil {
		return nil, nil, err
	}

	blkioDeviceWriteIOps, err := getThrottleDevice(c.DeviceWriteIOps)
	if err != nil {
		return nil, nil, err
	}

	resources := container.Resources{
		BlkioWeight:          uint16(c.BlkioWeight),
		BlkioWeightDevice:    blkioWeightDevices,
		CgroupParent:         c.CgroupParent,
		Memory:               int64(c.MemLimit),
		MemoryReservation:    int64(c.MemReservation),
		MemorySwap:           int64(c.MemSwapLimit),
		MemorySwappiness:     &memorySwappiness,
		CPUPeriod:            int64(c.CPUPeriod),
		CPUShares:            int64(c.CPUShares),
		CPUQuota:             int64(c.CPUQuota),
		CpusetCpus:           c.CPUSet,
		Ulimits:              ulimits,
		Devices:              deviceMappings,
		OomKillDisable:       &c.OomKillDisable,
		BlkioDeviceReadBps:   blkioDeviceReadBps,
		BlkioDeviceReadIOps:  blkioDeviceReadIOps,
		BlkioDeviceWriteBps:  blkioDeviceWriteBps,
		BlkioDeviceWriteIOps: blkioDeviceWriteIOps,
	}

	hostConfig := &container.HostConfig{
		VolumesFrom: volumesFrom,
		CapAdd:      strslice.StrSlice(utils.CopySlice(c.CapAdd)),
		CapDrop:     strslice.StrSlice(utils.CopySlice(c.CapDrop)),
		GroupAdd:    c.GroupAdd,
		ExtraHosts:  utils.CopySlice(c.ExtraHosts),
		Privileged:  c.Privileged,
		Binds:       Filter(vols, isBind),
		DNS:         utils.CopySlice(c.DNS),
		DNSOptions:  utils.CopySlice(c.DNSOpt),
		DNSSearch:   utils.CopySlice(c.DNSSearch),
		Init:        &c.Init,
		Isolation:   container.Isolation(c.Isolation),
		LogConfig: container.LogConfig{
			Type:   c.Logging.Driver,
			Config: utils.CopyMap(c.Logging.Options),
		},
		NetworkMode:    container.NetworkMode(c.NetworkMode),
		ReadonlyRootfs: c.ReadOnly,
		OomScoreAdj:    int(c.OomScoreAdj),
		PidMode:        container.PidMode(c.Pid),
		UTSMode:        container.UTSMode(c.Uts),
		IpcMode:        container.IpcMode(c.Ipc),
		PortBindings:   portBindings,
		RestartPolicy:  *restartPolicy,
		ShmSize:        int64(c.ShmSize),
		SecurityOpt:    utils.CopySlice(c.SecurityOpt),
		Sysctls:        utils.CopyMap(c.Sysctls),
		Tmpfs:          tmpfs,
		VolumeDriver:   c.VolumeDriver,
		Resources:      resources,
	}

	if config.Labels == nil {
		config.Labels = map[string]string{}
	}

	return config, hostConfig, nil
}

func getThrottleDevice(throttleConfig yaml.MaporColonSlice) ([]*blkiodev.ThrottleDevice, error) {
	var throttleDevice []*blkiodev.ThrottleDevice
	for _, deviceWriteIOps := range throttleConfig {
		split := strings.Split(deviceWriteIOps, ":")
		rate, err := strconv.ParseUint(split[1], 10, 64)
		if err != nil {
			return nil, err
		}

		throttleDevice = append(throttleDevice, &blkiodev.ThrottleDevice{
			Path: split[0],
			Rate: rate,
		})
	}

	return throttleDevice, nil
}

func getVolumesFrom(volumesFrom []string, serviceConfigs *config.ServiceConfigs, projectName string) ([]string, error) {
	volumes := []string{}
	for _, volumeFrom := range volumesFrom {
		if serviceConfig, ok := serviceConfigs.Get(volumeFrom); ok {
			// It's a service - Use the first one
			name := fmt.Sprintf("%s_%s_1", projectName, volumeFrom)
			// If a container name is specified, use that instead
			if serviceConfig.ContainerName != "" {
				name = serviceConfig.ContainerName
			}
			volumes = append(volumes, name)
		} else {
			volumes = append(volumes, volumeFrom)
		}
	}
	return volumes, nil
}

func parseDevices(devices []string) ([]container.DeviceMapping, error) {
	// parse device mappings
	deviceMappings := []container.DeviceMapping{}
	for _, device := range devices {
		v, err := opts.ParseDevice(device)
		if err != nil {
			return nil, err
		}
		deviceMappings = append(deviceMappings, container.DeviceMapping{
			PathOnHost:        v.PathOnHost,
			PathInContainer:   v.PathInContainer,
			CgroupPermissions: v.CgroupPermissions,
		})
	}

	return deviceMappings, nil
}

package rancher

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	"github.com/rancher/go-rancher/v2"
)

type RancherVolumesFactory struct {
	Context *Context
}

func (f *RancherVolumesFactory) Create(projectName string, volumeConfigs map[string]*config.VolumeConfig, serviceConfigs *config.ServiceConfigs, volumeEnabled bool) (project.Volumes, error) {
	volumes := make([]*Volume, 0, len(volumeConfigs))
	for name, config := range volumeConfigs {
		volume := NewVolume(projectName, name, config, f.Context)
		volumes = append(volumes, volume)
	}
	return &Volumes{
		volumes:       volumes,
		volumeEnabled: volumeEnabled,
		Context:       f.Context,
	}, nil
}

type Volumes struct {
	volumes       []*Volume
	volumeEnabled bool
	Context       *Context
}

func (v *Volumes) Initialize(ctx context.Context) error {
	if !v.volumeEnabled {
		return nil
	}
	for _, volume := range v.volumes {
		if err := volume.EnsureItExists(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (v *Volumes) Remove(ctx context.Context) error {
	if !v.volumeEnabled {
		return nil
	}
	for _, volume := range v.volumes {
		if err := volume.Remove(ctx); err != nil {
			return err
		}
	}
	return nil
}

type Volume struct {
	context       *Context
	name          string
	projectName   string
	driver        string
	driverOptions map[string]string
	external      bool
	perContainer  bool
}

// InspectTemplate looks up a volume template
func (v *Volume) InspectTemplate(ctx context.Context) (*client.VolumeTemplate, error) {
	volumes, err := v.context.Client.VolumeTemplate.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"name":    v.name,
			"stackId": v.context.Stack.Id,
		},
	})
	if err != nil {
		return nil, err
	}

	if len(volumes.Data) > 0 {
		return &volumes.Data[0], nil
	}

	return nil, nil
}

// InspectExternal looks up a volume
func (v *Volume) InspectExternal(ctx context.Context) (*client.Volume, error) {
	volumes, err := v.context.Client.Volume.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"name": v.name,
		},
	})
	if err != nil {
		return nil, err
	}

	if len(volumes.Data) > 0 {
		return &volumes.Data[0], nil
	}

	return nil, nil
}

func (v *Volume) Remove(ctx context.Context) error {
	if v.external {
		return nil
	}

	volumeResource, err := v.InspectTemplate(ctx)
	if err != nil {
		return err
	}
	return v.context.Client.VolumeTemplate.Delete(volumeResource)
}

func (v *Volume) EnsureItExists(ctx context.Context) error {
	if v.external {
		volumeResource, err := v.InspectExternal(ctx)
		if err != nil {
			return err
		}

		if volumeResource == nil {
			return fmt.Errorf("Volume %s declared as external, but could not be found. Please create the volume manually and try again.", v.name)
		}

		return nil
	}

	volumeResource, err := v.InspectTemplate(ctx)
	if err != nil {
		return err
	}

	if volumeResource == nil {
		logrus.Infof("Creating volume template %s", v.name)
		return v.create(ctx)
	} else {
		logrus.Infof("Existing volume template found for %s", v.name)
	}

	if v.driver != "" && volumeResource.Driver != v.driver {
		return fmt.Errorf("Volume %q needs to be recreated - driver has changed", v.name)
	}
	return nil
}

func (v *Volume) create(ctx context.Context) error {
	driverOptions := map[string]interface{}{}
	for k, v := range v.driverOptions {
		driverOptions[k] = v
	}
	_, err := v.context.Client.VolumeTemplate.Create(&client.VolumeTemplate{
		Name:         v.name,
		Driver:       v.driver,
		DriverOpts:   driverOptions,
		StackId:      v.context.Stack.Id,
		PerContainer: v.perContainer,
	})
	return err
}

func NewVolume(projectName, name string, config *config.VolumeConfig, context *Context) *Volume {
	return &Volume{
		context:       context,
		name:          name,
		projectName:   projectName,
		driver:        config.Driver,
		driverOptions: config.DriverOpts,
		external:      config.External.External,
		perContainer:  config.PerContainer,
	}
}

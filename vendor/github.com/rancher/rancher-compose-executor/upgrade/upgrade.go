package upgrade

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher/v2"
	"github.com/rancher/rancher-compose-executor/project"
	"github.com/rancher/rancher-compose-executor/project/options"
	"github.com/rancher/rancher-compose-executor/rancher"
)

type UpgradeOpts struct {
	BatchSize      int
	IntervalMillis int
	FinalScale     int
	UpdateLinks    bool
	Wait           bool
	CleanUp        bool
	Pull           bool
}

func Upgrade(p *project.Project, from, to string, opts UpgradeOpts) error {
	fromService, err := p.CreateService(from)
	if err != nil {
		return err
	}

	toService, err := p.CreateService(to)
	if err != nil {
		return err
	}

	rFromService, ok := fromService.(*rancher.RancherService)
	if !ok {
		return fmt.Errorf("%s is not a Rancher service", from)
	}

	source, err := rFromService.RancherService()
	if err != nil {
		return err
	}

	if source == nil {
		return fmt.Errorf("Failed to find service %s", from)
	}

	if source.LaunchConfig.Labels["io.rancher.scheduler.global"] == "true" {
		return fmt.Errorf("Upgrade is not supported for global services")
	}

	rToService, ok := toService.(*rancher.RancherService)
	if !ok {
		return fmt.Errorf("%s is not a Rancher service", to)
	}

	if service, err := rToService.RancherService(); err != nil {
		return err
	} else if service == nil {
		if err := rToService.Create(context.Background(), options.Create{}); err != nil {
			return err
		}

		// TODO timeout shouldn't really be an argument here
		// it's ignored in our implementation anyways
		if err := rToService.Scale(context.Background(), 0, -1); err != nil {
			return err
		}
	}

	if err := rToService.Up(context.Background(), options.Up{}); err != nil {
		return err
	}

	dest, err := rToService.RancherService()
	if err != nil {
		return err
	}

	if dest == nil {
		return fmt.Errorf("Failed to find service %s", to)
	}

	if dest.LaunchConfig.Labels["io.rancher.scheduler.global"] == "true" {
		return fmt.Errorf("Upgrade is not supported for global services")
	}

	upgradeOpts := &client.ServiceUpgrade{
		ToServiceStrategy: &client.ToServiceUpgradeStrategy{
			UpdateLinks:    opts.UpdateLinks,
			FinalScale:     int64(opts.FinalScale),
			BatchSize:      int64(opts.BatchSize),
			IntervalMillis: int64(opts.IntervalMillis),
			ToServiceId:    dest.Id,
		},
	}
	if upgradeOpts.ToServiceStrategy.FinalScale == -1 {
		upgradeOpts.ToServiceStrategy.FinalScale = source.Scale
	}

	client := rFromService.Client()

	if opts.Pull {
		if err := rToService.Pull(context.Background()); err != nil {
			return err
		}
	}

	logrus.Infof("Upgrading %s to %s, scale=%d", from, to, upgradeOpts.ToServiceStrategy.FinalScale)
	service, err := client.Service.ActionUpgrade(source, upgradeOpts)
	if err != nil {
		return err
	}

	if opts.Wait || opts.CleanUp {
		if err := rFromService.Wait(service); err != nil {
			return err
		}
	}

	if opts.CleanUp {
		// Reload source to check scale
		source, err = rFromService.RancherService()
		if err != nil {
			return err
		}

		if source.Scale == 0 {
			if err := rFromService.Delete(context.Background(), options.Delete{}); err != nil {
				return err
			}
		} else {
			logrus.Warnf("Not deleting service %s, scale is not 0 but %d", source.Name, source.Scale)
		}
	}

	return nil
}

func upgradeInfo(up bool, p *project.Project, from, to string, opts UpgradeOpts) (*client.Service, *client.Service, *client.RancherClient, error) {
	fromService, err := p.CreateService(from)
	if err != nil {
		return nil, nil, nil, err
	}

	toService, err := p.CreateService(to)
	if err != nil {
		return nil, nil, nil, err
	}

	rFromService, ok := fromService.(*rancher.RancherService)
	if !ok {
		return nil, nil, nil, fmt.Errorf("%s is not a Rancher service", from)
	}

	rToService, ok := toService.(*rancher.RancherService)
	if !ok {
		return nil, nil, nil, fmt.Errorf("%s is not a Rancher service", to)
	}

	if up {
		if err := rToService.Up(context.Background(), options.Up{}); err != nil {
			return nil, nil, nil, err
		}
	}

	source, err := rFromService.RancherService()
	if err != nil {
		return nil, nil, nil, err
	}

	dest, err := rToService.RancherService()
	if err != nil {
		return nil, nil, nil, err
	}

	return source, dest, rFromService.Client(), nil
}

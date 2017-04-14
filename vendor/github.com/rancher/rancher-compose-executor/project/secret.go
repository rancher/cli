package project

import (
	"github.com/rancher/rancher-compose-executor/config"
	"golang.org/x/net/context"
)

type Secrets interface {
	Initialize(ctx context.Context) error
}

type SecretsFactory interface {
	Create(projectName string, secretConfigs map[string]*config.SecretConfig) (Secrets, error)
}

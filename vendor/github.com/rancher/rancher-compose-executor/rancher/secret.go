package rancher

import (
	"encoding/base64"
	"fmt"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher/v2"
	"github.com/rancher/rancher-compose-executor/config"
	"github.com/rancher/rancher-compose-executor/project"
)

type RancherSecretsFactory struct {
	Context *Context
}

func (f *RancherSecretsFactory) Create(projectName string, secretConfigs map[string]*config.SecretConfig) (project.Secrets, error) {
	secrets := make([]*Secret, 0, len(secretConfigs))
	for name, config := range secretConfigs {
		secrets = append(secrets, &Secret{
			context:     f.Context,
			name:        name,
			projectName: projectName,
			file:        config.File,
			external:    config.External,
		})
	}
	return &Secrets{
		secrets: secrets,
		Context: f.Context,
	}, nil
}

type Secrets struct {
	secrets []*Secret
	Context *Context
}

func (s *Secrets) Initialize(ctx context.Context) error {
	for _, secret := range s.secrets {
		if err := secret.EnsureItExists(ctx); err != nil {
			return err
		}
	}
	return nil
}

type Secret struct {
	context     *Context
	name        string
	projectName string
	file        string
	external    string
}

func (s *Secret) EnsureItExists(ctx context.Context) error {
	existingSecrets, err := s.context.Client.Secret.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"name": s.name,
		},
	})
	if err != nil {
		return err
	}
	if len(existingSecrets.Data) > 0 {
		log.Infof("Secret %s already exists", s.name)
		return nil
	}
	if s.external != "" {
		return fmt.Errorf("Existing secret %s not found", s.name)
	}
	// TODO: use real relative path
	contents, filename, err := s.context.ResourceLookup.Lookup(s.file, "./")
	if err != nil {
		return err
	}
	log.Infof("Creating secret %s with contents from file %s", s.name, filename)
	_, err = s.context.Client.Secret.Create(&client.Secret{
		Name:  s.name,
		Value: base64.StdEncoding.EncodeToString(contents),
	})
	return err
}

package model

import "github.com/rancher/go-rancher/client"

//Template structure defines all properties that can be present in a template
type Template struct {
	client.Resource
	Name           string            `json:"name"`
	UUID           string            `json:"uuid"`
	Category       string            `json:"category"`
	Description    string            `json:"description"`
	Version        string            `json:"version"`
	IconLink       string            `json:"iconLink"`
	VersionLinks   map[string]string `json:"versionLinks"`
	DockerCompose  string            `json:"dockerCompose"`
	RancherCompose string            `json:"rancherCompose"`
	Questions      []Question        `json:"questions"`
	Path           string            `json:"path"`
}

//TemplateCollection holds a collection of templates
type TemplateCollection struct {
	client.Collection
	Data []Template `json:"data,omitempty"`
}

/*var Schemas = Schemas{
	Data: []Schema{
		{},
		{},
	},
}*/

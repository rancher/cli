package model

import "github.com/rancher/go-rancher/client"

//Template structure defines all properties that can be present in a template
type Template struct {
	client.Resource
	CatalogID                        string            `json:"catalogId"`
	Name                             string            `json:"name"`
	Category                         string            `json:"category"`
	IsSystem                         string            `json:"isSystem"`
	Description                      string            `json:"description"`
	Version                          string            `json:"version"`
	DefaultVersion                   string            `json:"defaultVersion"`
	IconLink                         string            `json:"iconLink"`
	VersionLinks                     map[string]string `json:"versionLinks"`
	UpgradeVersionLinks              map[string]string `json:"upgradeVersionLinks"`
	Files                            map[string]string `json:"files"`
	Questions                        []Question        `json:"questions"`
	Path                             string            `json:"path"`
	MinimumRancherVersion            string            `json:"minimumRancherVersion"`
	TemplateVersionRancherVersion    map[string]string
	TemplateVersionRancherVersionGte map[string]string
	Maintainer                       string                 `json:"maintainer"`
	License                          string                 `json:"license"`
	ProjectURL                       string                 `json:"projectURL"`
	ReadmeLink                       string                 `json:"readmeLink"`
	Output                           Output                 `json:"output" yaml:"output,omitempty"`
	TemplateBase                     string                 `json:"templateBase"`
	Labels                           map[string]string      `json:"labels"`
	UpgradeFrom                      string                 `json:"upgradeFrom"`
	Bindings                         map[string]interface{} `json:"bindings"`
	MaximumRancherVersion            string                 `json:"maximumRancherVersion"`
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

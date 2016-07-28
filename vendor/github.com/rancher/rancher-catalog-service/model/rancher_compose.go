package model

//import "github.com/rancher/rancher-compose/rancher"

//Question holds the properties of a question present in rancher-compose.yml file
type Question struct {
	Variable     string   `json:"variable" yaml:"variable,omitempty"`
	Label        string   `json:"label" yaml:"label,omitempty"`
	Description  string   `json:"description" yaml:"description,omitempty"`
	Type         string   `json:"type" yaml:"type,omitempty"`
	Required     bool     `json:"required" yaml:"required,omitempty"`
	Default      string   `json:"default" yaml:"default,omitempty"`
	Group        string   `json:"group" yaml:"group,omitempty"`
	MinLength    int      `json:"minLength" yaml:"min_length,omitempty"`
	MaxLength    int      `json:"maxLength" yaml:"max_length,omitempty"`
	Min          int      `json:"min" yaml:"min,omitempty"`
	Max          int      `json:"max" yaml:"max,omitempty"`
	Options      []string `json:"options" yaml:"options,omitempty"`
	ValidChars   string   `json:"validChars" yaml:"valid_chars,omitempty"`
	InvalidChars string   `json:"invalidChars" yaml:"invalid_chars,omitempty"`
}

//RancherCompose holds the questions array
type RancherCompose struct {
	//rancher.RancherConfig	`yaml:",inline"`
	Name        string     `yaml:"name"`
	UUID        string     `yaml:"uuid"`
	Description string     `yaml:"description"`
	Version     string     `yaml:"version"`
	Questions   []Question `json:"questions" yaml:"questions,omitempty"`
}

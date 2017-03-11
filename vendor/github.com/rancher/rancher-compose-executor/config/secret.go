package config

import "github.com/docker/libcompose/utils"

type SecretReferences []SecretReference

type SecretReference struct {
	Source string `yaml:"source,omitempty"`
	Target string `yaml:"target,omitempty"`
	Uid    string `yaml:"uid,omitempty"`
	Gid    string `yaml:"gid,omitempty"`
	Mode   string `yaml:"mode,omitempty"`
}

func (s *SecretReferences) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var secretReferences []SecretReference

	var sliceType []interface{}
	if err := unmarshal(&sliceType); err == nil {
		for _, elem := range sliceType {
			switch elem := elem.(type) {
			case string:
				secretReferences = append(secretReferences, SecretReference{
					Source: elem,
					Target: elem,
				})
			case map[interface{}]interface{}:
				var secretReference SecretReference
				if err = utils.Convert(elem, &secretReference); err != nil {
					return err
				}
				secretReferences = append(secretReferences, secretReference)
			}
		}
	}

	*s = secretReferences

	return nil
}

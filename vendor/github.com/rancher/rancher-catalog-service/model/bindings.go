package model

import (
	log "github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/config"
	utils "github.com/docker/libcompose/utils"
	libYaml "github.com/docker/libcompose/yaml"
	"github.com/rancher/rancher-compose/preprocess"
	"io/ioutil"
	"os"
)

//MapLabel represents labels from bindings
type MapLabel map[string]interface{}

//PortArray represents ports from bindings
type PortArray []interface{}

//BindingProperty holds bindings
type BindingProperty map[string]interface{}

//ServiceBinding holds the fields for ServiceBinding
type ServiceBinding struct {
	Labels map[string]interface{} `json:"labels"`
	Ports  []interface{}          `json:"ports"`
}

//CreateBindings creates bindings property
func CreateBindings(pathToYml string) (BindingProperty, error) {

	var bindingPropertyMap BindingProperty

	dockerFile := pathToYml + "/docker-compose.yml"
	_, err := os.Stat(dockerFile)
	if os.IsNotExist(err) {
		return BindingProperty{}, nil
	}

	yamlContent, err := ioutil.ReadFile(dockerFile)
	if err != nil {
		log.Errorf("Error in opening file : %v\n", err)
		return nil, err
	}

	bindingPropertyMap, err = ExtractBindings(yamlContent)
	if err != nil {
		return nil, err
	}
	return bindingPropertyMap, nil
}

//ExtractBindings gets bindings from created RawServiceMap
func ExtractBindings(yamlContent []byte) (BindingProperty, error) {
	var rawConfigDocker config.RawServiceMap
	var bindingsMap map[string]ServiceBinding
	var bindingPropertyMap BindingProperty
	var labels libYaml.SliceorMap

	config, err := config.CreateConfig(yamlContent)
	if err != nil {
		return nil, err
	}
	rawConfigDocker = config.Services

	preProcessServiceMap := preprocess.PreprocessServiceMap(nil)
	rawConfigDocker, err = preProcessServiceMap(rawConfigDocker)
	if err != nil {
		log.Errorf("Error during preprocess : %v\n", err)
		return nil, err
	}

	bindingsMap = make(map[string]ServiceBinding)
	bindingPropertyMap = make(map[string]interface{})

	for key := range rawConfigDocker {
		if _, serviceParsed := bindingsMap[key]; serviceParsed {
			log.Debugf("Service bindings already provided")
			continue
		}
		newServiceBinding := ServiceBinding{}

		newServiceBinding.Labels = MapLabel{}
		newServiceBinding.Ports = PortArray{}

		if rawConfigDocker[key]["labels"] != nil {
			err := utils.Convert(rawConfigDocker[key]["labels"], &labels)
			if err != nil {
				return nil, err
			}
			for k, v := range labels {
				newServiceBinding.Labels[k] = v
			}
		}
		if rawConfigDocker[key]["ports"] != nil {
			for _, port := range rawConfigDocker[key]["ports"].([]interface{}) {
				newServiceBinding.Ports = append(newServiceBinding.Ports, port)
			}
		}
		bindingsMap[key] = newServiceBinding
	}

	bindingPropertyMap["services"] = bindingsMap

	return bindingPropertyMap, nil
}

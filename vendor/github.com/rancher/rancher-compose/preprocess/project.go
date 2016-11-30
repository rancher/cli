package preprocess

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/docker/libcompose/config"
	"github.com/rancher/rancher-compose/utils"
)

type BindingProperty struct {
	Services map[string]Service `json:"services"`
}

type Service struct {
	Labels map[string]interface{} `json:"labels"`
	Ports  []interface{}          `json:"ports"`
}

func PreprocessServiceMap(bindingsBytes []byte) func(serviceMap config.RawServiceMap) (config.RawServiceMap, error) {
	return func(serviceMap config.RawServiceMap) (config.RawServiceMap, error) {
		newServiceMap := make(config.RawServiceMap)

		var binding BindingProperty
		var bindingsServices []string

		if bindingsBytes != nil {
			err := json.Unmarshal(bindingsBytes, &binding)
			if err != nil {
				return nil, err
			}

			for k := range binding.Services {
				bindingsServices = append(bindingsServices, k)
			}
		}

		for k, v := range serviceMap {
			newServiceMap[k] = make(config.RawService)
			if bindingsBytes != nil {
				if utils.Contains(bindingsServices, k) == true {
					labelMap := make(map[interface{}]interface{})
					for key, value := range binding.Services[k].Labels {
						labelMap[interface{}(key)] = value
					}
					if len(labelMap) != 0 {
						v["labels"] = labelMap
					}
					if len(binding.Services[k].Ports) > 0 {
						v["ports"] = binding.Services[k].Ports
					}
				}
			}
			for k2, v2 := range v {
				if k2 == "environment" || k2 == "labels" {
					newServiceMap[k][k2] = Preprocess(v2, true)
				} else {
					newServiceMap[k][k2] = Preprocess(v2, false)
				}

			}
		}

		return newServiceMap, nil
	}
}

func Preprocess(item interface{}, replaceTypes bool) interface{} {
	switch typedDatas := item.(type) {

	case map[interface{}]interface{}:
		newMap := make(map[interface{}]interface{})

		for key, value := range typedDatas {
			newMap[key] = Preprocess(value, replaceTypes)
		}
		return newMap

	case []interface{}:
		// newArray := make([]interface{}, 0) will cause golint to complain
		var newArray []interface{}
		newArray = make([]interface{}, 0)

		for _, value := range typedDatas {
			newArray = append(newArray, Preprocess(value, replaceTypes))
		}
		return newArray

	default:
		if replaceTypes && item != nil {
			return fmt.Sprint(item)
		}
		return item
	}
}

func TryConvertStringsToInts(serviceMap config.RawServiceMap, fields map[string]bool) (config.RawServiceMap, error) {
	newServiceMap := make(config.RawServiceMap)

	for k, v := range serviceMap {
		newServiceMap[k] = make(config.RawService)

		for k2, v2 := range v {
			if _, ok := fields[k2]; ok {
				newServiceMap[k][k2] = tryConvertStringsToInts(v2, true)
			} else {
				newServiceMap[k][k2] = tryConvertStringsToInts(v2, false)
			}

		}
	}

	return newServiceMap, nil
}

func tryConvertStringsToInts(item interface{}, replaceTypes bool) interface{} {
	switch typedDatas := item.(type) {

	case map[interface{}]interface{}:
		newMap := make(map[interface{}]interface{})

		for key, value := range typedDatas {
			newMap[key] = tryConvertStringsToInts(value, replaceTypes)
		}
		return newMap

	case []interface{}:
		// newArray := make([]interface{}, 0) will cause golint to complain
		var newArray []interface{}
		newArray = make([]interface{}, 0)

		for _, value := range typedDatas {
			newArray = append(newArray, tryConvertStringsToInts(value, replaceTypes))
		}
		return newArray

	case string:
		lineAsInteger, err := strconv.Atoi(typedDatas)

		if replaceTypes && err == nil {
			return lineAsInteger
		}

		return item
	default:
		return item
	}
}

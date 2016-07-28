package preprocess

import (
	"fmt"

	"github.com/docker/libcompose/config"
)

func PreprocessServiceMap(serviceMap config.RawServiceMap) (config.RawServiceMap, error) {
	newServiceMap := make(config.RawServiceMap)

	for k, v := range serviceMap {
		newServiceMap[k] = make(config.RawService)

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
		if replaceTypes {
			return fmt.Sprint(item)
		}
		return item
	}
}

package config

import (
	"fmt"
	"strconv"
)

func PreprocessServiceMap(serviceMap RawServiceMap) (RawServiceMap, error) {
	newServiceMap := make(RawServiceMap)

	for k, v := range serviceMap {
		newServiceMap[k] = make(RawService)
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
		if replaceTypes && item != nil {
			return fmt.Sprint(item)
		}
		return item
	}
}

func TryConvertStringsToInts(serviceMap RawServiceMap, fields map[string]bool) (RawServiceMap, error) {
	newServiceMap := make(RawServiceMap)

	for k, v := range serviceMap {
		newServiceMap[k] = make(RawService)

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

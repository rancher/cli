package utils

import (
	"fmt"
	"strings"
)

func NestedMapsToMapInterface(data map[string]interface{}) map[string]interface{} {
	newMapInterface := map[string]interface{}{}
	for k, v := range data {
		newMapInterface[k] = convertObj(v)
	}
	return newMapInterface
}

func convertObj(v interface{}) interface{} {
	switch k := v.(type) {
	case map[interface{}]interface{}:
		return mapWalk(k)
	case map[string]interface{}:
		return NestedMapsToMapInterface(k)
	case []interface{}:
		return listWalk(k)
	default:
		return v
	}
}

func listWalk(val []interface{}) []interface{} {
	for i, v := range val {
		val[i] = convertObj(v)
	}
	return val
}

func mapWalk(val map[interface{}]interface{}) map[string]interface{} {
	newMap := map[string]interface{}{}
	for k, v := range val {
		newMap[fmt.Sprintf("%v", k)] = convertObj(v)
	}
	return newMap
}

func Contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

func ToMapInterface(data map[string]string) map[string]interface{} {
	ret := map[string]interface{}{}

	for k, v := range data {
		ret[k] = v
	}

	return ret
}

func RemoveInterfaceKeys(data interface{}) interface{} {
	switch value := data.(type) {
	case map[interface{}]interface{}:
		ret := map[string]interface{}{}
		for k, v := range value {
			ret[fmt.Sprintf("%v", k)] = v
		}
		return ret
	case []interface{}:
		for i, j := range value {
			value[i] = RemoveInterfaceKeys(j)
		}
	case map[string]interface{}:
		for k, v := range value {
			value[k] = RemoveInterfaceKeys(v)
		}
	}
	return data
}

func MapUnion(left, right map[string]string) map[string]string {
	ret := map[string]string{}

	for k, v := range left {
		ret[k] = v
	}

	for k, v := range right {
		ret[k] = v
	}

	return ret
}

func TrimSplit(str, sep string, count int) []string {
	result := []string{}
	for _, i := range strings.SplitN(strings.TrimSpace(str), sep, count) {
		result = append(result, strings.TrimSpace(i))
	}

	return result
}

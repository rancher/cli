package digest

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"reflect"
	"sort"

	"github.com/docker/libcompose/utils"
	rancherClient "github.com/rancher/go-rancher/client"
	rUtils "github.com/rancher/rancher-compose/utils"
)

const (
	ServiceHashKey = "io.rancher.service.hash"
)

var (
	ignoreKeys = []string{
		ServiceHashKey,
		"links",
		"scale",
		"selector_container",
		"selector_link",
		"environment_id",
	}
)

type ServiceHash struct {
	Service                string            `json:"service,omitempty" yaml:"service,omitempty"`
	LaunchConfig           string            `json:"launch_config,omitempty" yaml:"launch_config,omitempty"`
	SecondaryLaunchConfigs map[string]string `json:"secondary_launch_configs,omitempty" yaml:"secondary_launch_configs,omitempty"`
}

func (s ServiceHash) Equals(hash ServiceHash) bool {
	return s.Service == hash.Service &&
		s.LaunchConfig == hash.LaunchConfig &&
		// The check for 0 handles when one of the maps is nil and the other is empty
		(len(s.SecondaryLaunchConfigs) == 0 && len(hash.SecondaryLaunchConfigs) == 0 ||
			reflect.DeepEqual(s.SecondaryLaunchConfigs, hash.SecondaryLaunchConfigs))
}

func toString(obj interface{}) string {
	if obj == nil {
		return ""
	}
	return fmt.Sprintf("%v", obj)
}

func LookupHash(service *rancherClient.Service) (ServiceHash, bool) {
	ret := ServiceHash{
		SecondaryLaunchConfigs: map[string]string{},
	}

	ret.Service = toString(service.Metadata[ServiceHashKey])
	ret.LaunchConfig = toString(service.LaunchConfig.Labels[ServiceHashKey])

	for _, rawSecondaryLaunchConfig := range service.SecondaryLaunchConfigs {
		var secondaryLaunchConfig rancherClient.SecondaryLaunchConfig
		if err := utils.Convert(rawSecondaryLaunchConfig, &secondaryLaunchConfig); err != nil {
			return ret, false
		}
		ret.SecondaryLaunchConfigs[secondaryLaunchConfig.Name] = toString(secondaryLaunchConfig.Labels[ServiceHashKey])
	}

	return ret, ret.Service != ""
}

func CreateServiceHash(rancherService interface{}, launchConfig *rancherClient.LaunchConfig, secondaryLaunchConfigs []rancherClient.SecondaryLaunchConfig) (ServiceHash, error) {
	var err error
	result := ServiceHash{}
	if err != nil {
		return result, err
	}

	result.Service, err = hashObj(rancherService)
	if err != nil {
		return result, err
	}

	result.LaunchConfig, err = hashObj(launchConfig)
	if err != nil {
		return result, err
	}

	for _, secondaryLaunchConfig := range secondaryLaunchConfigs {
		hash, err := hashObj(secondaryLaunchConfig)
		if err != nil {
			return result, err
		}

		if result.SecondaryLaunchConfigs == nil {
			result.SecondaryLaunchConfigs = map[string]string{}
		}

		result.SecondaryLaunchConfigs[secondaryLaunchConfig.Name] = hash
	}

	return result, nil
}

func toSortedStringMap(data map[interface{}]interface{}) ([]string, map[string]interface{}) {
	keys := []string{}
	ret := map[string]interface{}{}

	for k, v := range data {
		str := fmt.Sprintf("%v", k)
		keys = append(keys, str)
		ret[str] = v
	}

	sort.Strings(keys)
	return keys, ret
}

func writeNullTerminatedValue(writer io.Writer, value interface{}) {
	switch s := value.(type) {
	case map[interface{}]interface{}:
		writeNativeMap(writer, s)
	case []interface{}:
		for _, sliceValue := range s {
			writeNullTerminatedValue(writer, sliceValue)
		}
	default:
		io.WriteString(writer, fmt.Sprintf("%v", value))
		writer.Write([]byte{0})
	}
}

func hashObj(obj interface{}) (string, error) {
	hash := sha1.New()

	mapObj := map[interface{}]interface{}{}
	if err := utils.Convert(obj, &mapObj); err != nil {
		return "", err
	}

	writeNativeMap(hash, mapObj)

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func writeNativeMap(writer io.Writer, rawData map[interface{}]interface{}) {
	keys, data := toSortedStringMap(rawData)
	for _, key := range keys {
		if !rUtils.Contains(ignoreKeys, key) {
			writeNullTerminatedValue(writer, key)
			writeNullTerminatedValue(writer, data[key])
		}
	}
}

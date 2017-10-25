package lookup

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/rancher/rancher-compose-executor/config"
	"github.com/rancher/rancher-compose-executor/utils"
)

type FileEnvLookup struct {
	parent    config.EnvironmentLookup
	variables map[string]string
}

func parseMultiLineEnv(file string) (map[string]interface{}, error) {
	variables := map[string]interface{}{}

	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}

	if strings.HasSuffix(file, ".yml") || strings.HasSuffix(file, ".yaml") {
		yaml.Unmarshal(contents, &data)
	} else if strings.HasSuffix(file, ".json") {
		json.Unmarshal(contents, &data)
	}

	for k, v := range data {
		if stringValue, ok := v.(string); ok {
			variables[k] = stringValue
		} else if intValue, ok := v.(int); ok {
			variables[k] = fmt.Sprintf("%v", intValue)
		} else if int64Value, ok := v.(int64); ok {
			variables[k] = fmt.Sprintf("%v", int64Value)
		} else if float32Value, ok := v.(float32); ok {
			variables[k] = fmt.Sprintf("%v", float32Value)
		} else if float64Value, ok := v.(float64); ok {
			variables[k] = fmt.Sprintf("%v", float64Value)
		} else if boolValue, ok := v.(bool); ok {
			variables[k] = strconv.FormatBool(boolValue)
		} else {
			return nil, fmt.Errorf("Environment variables must be of type string, bool, or int. Key %s is of type %T", k, v)
		}
	}

	return variables, nil
}

func parseCustomEnvFile(file string) (map[string]interface{}, error) {
	variables := map[string]interface{}{}

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		t := scanner.Text()
		parts := strings.SplitN(t, "=", 2)
		if len(parts) == 1 {
			variables[parts[0]] = ""
		} else {
			variables[parts[0]] = parts[1]
		}
	}

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	return variables, nil
}

func ParseEnvFile(file string) (map[string]interface{}, error) {
	if file != "" {
		if strings.HasSuffix(file, ".yml") || strings.HasSuffix(file, ".yaml") || strings.HasSuffix(file, ".json") {

			v, err := parseMultiLineEnv(file)
			if err != nil {
				return nil, err
			}

			return v, nil
		}

		v, err := parseCustomEnvFile(file)
		if err != nil {
			return nil, err
		}

		return v, nil
	}
	return map[string]interface{}{}, nil
}

func NewFileEnvLookup(file string, qLookup *QuestionLookup) (*FileEnvLookup, error) {
	v, err := ParseEnvFile(file)
	if err != nil {
		return nil, err
	}

	variables := map[string]string{}
	for key, value := range v {
		switch value := value.(type) {
		case string:
			variables[key] = value
		}
	}

	parent := qLookup.parent
	qVariables := qLookup.variables
	questions := qLookup.questions

	combMap := map[string]string{}
	combMap = utils.MapUnion(variables, qVariables)

	for key, question := range questions {
		if _, ok := combMap[key]; ok {
			continue
		}
		answer := ask(question)
		if answer != "" {
			variables[key] = answer
		}
	}

	return &FileEnvLookup{
		parent:    parent,
		variables: variables,
	}, nil
}

func (f *FileEnvLookup) Lookup(key string, config *config.ServiceConfig) []string {
	if v, ok := f.variables[key]; ok {
		return []string{fmt.Sprintf("%s=%s", key, v)}
	}

	if f.parent == nil {
		return nil
	}

	return f.parent.Lookup(key, config)
}

func (f *FileEnvLookup) Variables() map[string]string {
	return utils.MapUnion(f.variables, f.parent.Variables())
}

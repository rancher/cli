package config

import (
	"encoding/json"
	"strings"

	"github.com/docker/go-connections/nat"
	"github.com/xeipuuv/gojsonschema"
)

var (
	schemaLoaderV1           gojsonschema.JSONLoader
	constraintSchemaLoaderV1 gojsonschema.JSONLoader
	schemaLoaderV2           gojsonschema.JSONLoader
	constraintSchemaLoaderV2 gojsonschema.JSONLoader
	schemaV1                 map[string]interface{}
	schemaV2                 map[string]interface{}
)

type (
	environmentFormatChecker struct{}
	portsFormatChecker       struct{}
)

func (checker environmentFormatChecker) IsFormat(input string) bool {
	// If the value is a boolean, a warning should be given
	// However, we can't determine type since gojsonschema converts the value to a string
	// Adding a function with an interface{} parameter to gojsonschema is probably the best way to handle this
	return true
}

func (checker portsFormatChecker) IsFormat(input string) bool {
	_, _, err := nat.ParsePortSpecs([]string{input})
	return err == nil
}

func setupSchemaLoaders(schemaData string, schema *map[string]interface{}, schemaLoader, constraintSchemaLoader *gojsonschema.JSONLoader) error {
	if *schema != nil {
		return nil
	}

	var schemaRaw interface{}
	err := json.Unmarshal([]byte(schemaData), &schemaRaw)
	if err != nil {
		return err
	}

	*schema = schemaRaw.(map[string]interface{})

	gojsonschema.FormatCheckers.Add("environment", environmentFormatChecker{})
	gojsonschema.FormatCheckers.Add("ports", portsFormatChecker{})
	gojsonschema.FormatCheckers.Add("expose", portsFormatChecker{})
	*schemaLoader = gojsonschema.NewGoLoader(schemaRaw)

	definitions := (*schema)["definitions"].(map[string]interface{})
	constraints := definitions["constraints"].(map[string]interface{})
	service := constraints["service"].(map[string]interface{})
	*constraintSchemaLoader = gojsonschema.NewGoLoader(service)

	return nil
}

func appendValidTypes(oneOf interface{}, validTypes []string) []string {
	validConditions, ok := oneOf.([]interface{})

	if ok {
		for _, validCondition := range validConditions {
			condition := validCondition.(map[string]interface{})
			if _, ok := condition["type"]; ok {
				validTypes = append(validTypes, condition["type"].(string))
			}
		}

		return validTypes
	}

	return []string{}
}

// gojsonschema doesn't provide a list of valid types for a property
// This parses the schema manually to find all valid types
func parseValidTypesFromSchema(schema map[string]interface{}, context string) []string {
	contextSplit := strings.Split(context, ".")
	key := contextSplit[len(contextSplit)-1]

	definitions := make(map[string]interface{})
	if _, ok := schema["definitions"]; ok {
		definitions = schema["definitions"].(map[string]interface{})
	}

	service := make(map[string]interface{})
	if _, ok := definitions["service"]; ok {
		service = definitions["service"].(map[string]interface{})
	}

	properties := make(map[string]interface{})
	if _, ok := service["properties"]; ok {
		properties = service["properties"].(map[string]interface{})
	}

	property := make(map[string]interface{})
	if _, ok := properties[key]; ok {
		property = properties[key].(map[string]interface{})
	}

	var validTypes []string

	if val, ok := property["oneOf"]; ok {
		validTypes = appendValidTypes(val, validTypes)
	} else if val, ok := property["$ref"]; ok {
		reference := val.(string)
		if reference == "#/definitions/string_or_list" {
			return []string{"string", "array"}
		} else if reference == "#/definitions/list_of_strings" {
			return []string{"array"}
		} else if reference == "#/definitions/list_or_dict" {
			return []string{"array", "object"}
		}
	}

	if _, ok := property["oneOf"]; !ok {
		key = contextSplit[len(contextSplit)-2]

		property := make(map[string]interface{})
		if _, ok := properties[key]; ok {
			property = properties[key].(map[string]interface{})
		}

		if _, ok := property["patternProperties"]; ok {
			// address schema for 'ulimits'
			patternProperties := property["patternProperties"].(map[string]interface{})

			for _, value := range patternProperties {
				patternProperty := value.(map[string]interface{})

				if _, ok := patternProperty["oneOf"]; ok {
					validTypes = appendValidTypes(patternProperty["oneOf"], validTypes)
				}
			}
		} else if _, ok := property["items"]; ok {
			// address schema for 'secrets'
			items := property["items"].(map[string]interface{})

			if val, ok := items["oneOf"]; ok {
				validTypes = appendValidTypes(val, validTypes)
			}
		}
	}

	return validTypes
}

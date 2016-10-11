package main

import (
	"io/ioutil"
	"os"
	"text/template"
)

func main() {
	t, err := template.New("schema_template").ParseFiles("./scripts/schema_template")
	if err != nil {
		panic(err)
	}

	schemaV1, err := ioutil.ReadFile("./scripts/config_schema_v1.json")
	if err != nil {
		panic(err)
	}
	schemaV2, err := ioutil.ReadFile("./scripts/config_schema_v2.0.json")
	if err != nil {
		panic(err)
	}

	inlinedFile, err := os.Create("vendor/github.com/docker/libcompose/config/schema.go")
	if err != nil {
		panic(err)
	}

	err = t.Execute(inlinedFile, map[string]string{
		"schemaV1": string(schemaV1),
		"schemaV2": string(schemaV2),
	})

	if err != nil {
		panic(err)
	}
}

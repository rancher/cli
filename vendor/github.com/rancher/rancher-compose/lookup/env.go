package lookup

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/config"
)

type FileEnvLookup struct {
	parent    config.EnvironmentLookup
	variables map[string]string
}

func NewFileEnvLookup(file string, parent config.EnvironmentLookup) (*FileEnvLookup, error) {
	variables := map[string]string{}

	if file != "" {
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
	}

	logrus.Debugf("Environment Context from file %s: %v", file, variables)
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

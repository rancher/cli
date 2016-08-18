package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rancher/go-rancher/client"
)

func pickAction(resource *client.Resource, actions ...string) (string, error) {
	for _, action := range actions {
		if _, ok := resource.Actions[action]; ok {
			return action, nil
		}
	}
	msg := fmt.Sprintf("%s not currently available on %s %s", strings.Join(actions, "/"), resource.Type, resource.Id)
	return "", errors.New(replaceTypeNames(msg))
}

func replaceTypeNames(msg string) string {
	msg = strings.Replace(msg, "environment", "stack", -1)
	return strings.Replace(msg, "project", "enviroment", -1)
}

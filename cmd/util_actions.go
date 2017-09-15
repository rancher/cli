package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rancher/go-rancher/v3"
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
	return strings.Replace(msg, "project", "environment", -1)
}

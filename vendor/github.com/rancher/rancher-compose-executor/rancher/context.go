package rancher

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher/v2"
	"github.com/rancher/rancher-compose-executor/project"
	"github.com/rancher/rancher-compose-executor/utils"
)

var projectRegexp = regexp.MustCompile("[^a-zA-Z0-9-]")

type Context struct {
	project.Context

	Url          string
	AccessKey    string
	SecretKey    string
	Client       *client.RancherClient
	Stack        *client.Stack
	isOpen       bool
	SidekickInfo *SidekickInfo
	Uploader     Uploader
	PullCached   bool
	Pull         bool
	Prune        bool
	Args         []string

	Upgrade        bool
	ForceUpgrade   bool
	Rollback       bool
	Interval       int64
	BatchSize      int64
	ConfirmUpgrade bool
}

func (c *Context) sanitizedProjectName() string {
	projectName := projectRegexp.ReplaceAllString(strings.ToLower(c.ProjectName), "-")

	if len(projectName) > 0 && strings.ContainsAny(projectName[0:1], "_.-") {
		projectName = "x" + projectName
	}

	return projectName
}

func (c *Context) loadClient() (*client.RancherClient, error) {
	if c.Client == nil {
		if c.Url == "" {
			return nil, fmt.Errorf("RANCHER_URL is not set")
		}

		if client, err := client.NewRancherClient(&client.ClientOpts{
			Url:       c.Url,
			AccessKey: c.AccessKey,
			SecretKey: c.SecretKey,
		}); err != nil {
			return nil, err
		} else {
			c.Client = client
		}
	}

	return c.Client, nil
}

func (c *Context) open() error {
	if c.isOpen {
		return nil
	}

	c.ProjectName = c.sanitizedProjectName()

	if _, err := c.loadClient(); err != nil {
		return err
	}

	if stackSchema, ok := c.Client.GetTypes()["stack"]; !ok || !utils.Contains(stackSchema.CollectionMethods, "POST") {
		return fmt.Errorf("Can not create a stack, check API key [%s] for [%s]", c.AccessKey, c.Url)
	}

	stack, err := c.LoadStack()
	if err != nil {
		return err
	}
	proj, err := c.Client.Project.ById(stack.AccountId)
	if err != nil {
		return err
	}
	c.EnvironmentName = proj.Name

	c.isOpen = true
	return nil
}

func (c *Context) LoadStack() (*client.Stack, error) {
	if c.Stack != nil {
		return c.Stack, nil
	}

	projectName := c.sanitizedProjectName()
	if _, err := c.loadClient(); err != nil {
		return nil, err
	}

	logrus.Debugf("Looking for stack %s", projectName)
	// First try by name
	stacks, err := c.Client.Stack.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"name":         projectName,
			"removed_null": nil,
		},
	})
	if err != nil {
		return nil, err
	}

	for _, stack := range stacks.Data {
		if strings.EqualFold(projectName, stack.Name) {
			logrus.Debugf("Found stack: %s(%s)", stack.Name, stack.Id)
			c.Stack = &stack
			return c.Stack, nil
		}
	}

	// Now try not by name for case sensitive databases
	stacks, err = c.Client.Stack.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"removed_null": nil,
		},
	})
	if err != nil {
		return nil, err
	}

	for _, stack := range stacks.Data {
		if strings.EqualFold(projectName, stack.Name) {
			logrus.Debugf("Found stack: %s(%s)", stack.Name, stack.Id)
			c.Stack = &stack
			return c.Stack, nil
		}
	}

	logrus.Infof("Creating stack %s", projectName)
	stack, err := c.Client.Stack.Create(&client.Stack{
		Name: projectName,
	})
	if err != nil {
		return nil, err
	}

	c.Stack = stack

	return c.Stack, nil
}

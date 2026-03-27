package cmd

import (
	"bufio"
	"io"
	"strings"

	"github.com/rancher/norman/types"
	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
)

type fakePrincipalGetter struct {
	ByIDFunc func(id string) (*managementClient.Principal, error)
}

func (g *fakePrincipalGetter) ByID(id string) (*managementClient.Principal, error) {
	if g.ByIDFunc != nil {
		return g.ByIDFunc(id)
	}
	return nil, nil
}

type fakeUserConfig struct {
	GetCurrentClusterFunc func() string
	GetCurrentProjectFunc func() string
}

func (c *fakeUserConfig) GetCurrentCluster() string {
	if c.GetCurrentClusterFunc != nil {
		return c.GetCurrentClusterFunc()
	}
	return ""
}

func (c *fakeUserConfig) GetCurrentProject() string {
	if c.GetCurrentProjectFunc != nil {
		return c.GetCurrentProjectFunc()
	}
	return ""
}

type fakeCRTBLister struct {
	ListFunc func(opts *types.ListOpts) (*managementClient.ClusterRoleTemplateBindingCollection, error)
}

func (f *fakeCRTBLister) List(opts *types.ListOpts) (*managementClient.ClusterRoleTemplateBindingCollection, error) {
	if f.ListFunc != nil {
		return f.ListFunc(opts)
	}
	return nil, nil
}

type fakePRTBLister struct {
	ListFunc func(opts *types.ListOpts) (*managementClient.ProjectRoleTemplateBindingCollection, error)
}

func (f *fakePRTBLister) List(opts *types.ListOpts) (*managementClient.ProjectRoleTemplateBindingCollection, error) {
	if f.ListFunc != nil {
		return f.ListFunc(opts)
	}
	return nil, nil
}

func parseTabWriterOutput(r io.Reader) [][]string {
	var parsed [][]string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		var fields []string
		for _, field := range strings.Split(scanner.Text(), "  ") {
			if field == "" {
				continue
			}
			fields = append(fields, strings.TrimSpace(field))
		}
		parsed = append(parsed, fields)
	}
	return parsed
}

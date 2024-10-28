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
	FocusedClusterFunc func() string
	FocusedProjectFunc func() string
}

func (c *fakeUserConfig) FocusedCluster() string {
	if c.FocusedClusterFunc != nil {
		return c.FocusedClusterFunc()
	}
	return ""
}

func (c *fakeUserConfig) FocusedProject() string {
	if c.FocusedProjectFunc != nil {
		return c.FocusedProjectFunc()
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

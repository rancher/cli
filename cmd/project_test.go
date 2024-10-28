package cmd

import (
	"bytes"
	"flag"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/rancher/norman/types"
	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

func TestListProjectMembers(t *testing.T) {
	now := time.Now()

	userConfig := &fakeUserConfig{
		FocusedProjectFunc: func() string {
			return "c-fn7lc:p-9mdxl"
		},
	}

	prtbs := &fakePRTBLister{
		ListFunc: func(opts *types.ListOpts) (*managementClient.ProjectRoleTemplateBindingCollection, error) {
			return &managementClient.ProjectRoleTemplateBindingCollection{
				Data: []managementClient.ProjectRoleTemplateBinding{
					{
						Resource: types.Resource{
							ID: "p-9mdxl:creator-project-owner",
						},
						Created:         now.Format(time.RFC3339),
						RoleTemplateID:  "project-owner",
						UserPrincipalID: "local://user-2p7w6",
					},
					{
						Resource: types.Resource{
							ID: "p-9mdxl:prtb-qd49d",
						},
						Created:          now.Format(time.RFC3339),
						RoleTemplateID:   "project-member",
						GroupPrincipalID: "okta_group://b4qkhsnliz",
					},
				},
			}, nil
		},
	}

	principals := &fakePrincipalGetter{
		ByIDFunc: func(id string) (*managementClient.Principal, error) {
			id, err := url.PathUnescape(id)
			require.NoError(t, err)

			switch id {
			case "local://user-2p7w6":
				return &managementClient.Principal{
					Name:          "Default Admin",
					LoginName:     "admin",
					Provider:      "local",
					PrincipalType: "user",
				}, nil
			case "okta_group://b4qkhsnliz":
				return &managementClient.Principal{
					Name:          "DevOps",
					LoginName:     "devops",
					Provider:      "okta",
					PrincipalType: "group",
				}, nil
			default:
				return nil, fmt.Errorf("not found")
			}
		},
	}

	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	cctx := cli.NewContext(nil, flagSet, nil)

	var out bytes.Buffer

	err := listProjectMembers(cctx, &out, userConfig, prtbs, principals)
	require.NoError(t, err)
	require.NotEmpty(t, out)

	output := out.String()
	assert.Contains(t, output, "p-9mdxl:creator-project-owner")
	assert.Contains(t, output, "Default Admin (Local User)")
	assert.Contains(t, output, "DevOps (Okta Group)")
}

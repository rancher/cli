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

func TestListClusterMembers(t *testing.T) {
	now := time.Now()

	userConfig := &fakeUserConfig{
		FocusedClusterFunc: func() string {
			return "c-fn7lc"
		},
	}

	created := now.Format(time.RFC3339)
	crtbs := &fakeCRTBLister{
		ListFunc: func(opts *types.ListOpts) (*managementClient.ClusterRoleTemplateBindingCollection, error) {
			return &managementClient.ClusterRoleTemplateBindingCollection{
				Data: []managementClient.ClusterRoleTemplateBinding{
					{
						Resource: types.Resource{
							ID: "c-fn7lc:creator-cluster-owner",
						},
						Created:         created,
						RoleTemplateID:  "cluster-owner",
						UserPrincipalID: "local://user-2p7w6",
					},
					{
						Resource: types.Resource{
							ID: "c-fn7lc:crtb-qd49d",
						},
						Created:          created,
						RoleTemplateID:   "cluster-member",
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

	err := listClusterMembers(cctx, &out, userConfig, crtbs, principals)
	require.NoError(t, err)
	require.NotEmpty(t, out)

	humanCreated := now.Format(humanTimeFormat)
	want := [][]string{
		{"BINDING-ID", "MEMBER", "ROLE", "CREATED"},
		{"c-fn7lc:creator-cluster-owner", "Default Admin (Local User)", "cluster-owner", humanCreated},
		{"c-fn7lc:crtb-qd49d", "DevOps (Okta Group)", "cluster-member", humanCreated},
	}

	got := parseTabWriterOutput(&out)
	assert.Equal(t, want, got)
}

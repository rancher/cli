package cmd

import (
	"fmt"
	"net/url"
	"strconv"
	"testing"

	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseClusterAndProjectID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id,
		cluster string
		project   string
		shouldErr bool
	}{
		{
			id:      "local:p-12345",
			cluster: "local",
			project: "p-12345",
		},
		{
			id:      "c-12345:p-12345",
			cluster: "c-12345",
			project: "p-12345",
		},
		{
			id:        "cocal:p-12345",
			shouldErr: true,
		},
		{
			id:        "c-123:p-123",
			shouldErr: true,
		},
		{
			shouldErr: true,
		},
		{
			id:      "c-m-12345678:p-12345",
			cluster: "c-m-12345678",
			project: "p-12345",
		},
		{
			id:        "c-m-123:p-12345",
			shouldErr: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.id, func(t *testing.T) {
			t.Parallel()

			cluster, project, err := parseClusterAndProjectID(test.id)
			if test.shouldErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.cluster, cluster)
			assert.Equal(t, test.project, project)
		})
	}
}

func TestConvertSnakeCaseKeysToCamelCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input map[string]any
		want  map[string]any
	}{
		{
			map[string]any{"foo_bar": "hello"},
			map[string]any{"fooBar": "hello"},
		},
		{
			map[string]any{"fooBar": "hello"},
			map[string]any{"fooBar": "hello"},
		},
		{
			map[string]any{"foobar": "hello", "some_key": "valueUnmodified", "bar-baz": "bar-baz"},
			map[string]any{"foobar": "hello", "someKey": "valueUnmodified", "bar-baz": "bar-baz"},
		},
		{
			map[string]any{"foo_bar": "hello", "backup_config": map[string]any{"hello_world": true}, "config_id": 123},
			map[string]any{"fooBar": "hello", "backupConfig": map[string]any{"helloWorld": true}, "configId": 123},
		},
	}

	for i, test := range tests {
		test := test
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			convertSnakeCaseKeysToCamelCase(test.input)
			assert.Equal(t, test.input, test.want)
		})
	}
}

func TestParsePrincipalID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id   string
		want *managementClient.Principal
	}{
		{
			id: "local://user-2p7w6",
			want: &managementClient.Principal{
				Name:          "user-2p7w6",
				LoginName:     "user-2p7w6",
				Provider:      "local",
				PrincipalType: "user",
			},
		},
		{
			id: "okta_group://b4qkhsnliz",
			want: &managementClient.Principal{
				Name:          "b4qkhsnliz",
				LoginName:     "b4qkhsnliz",
				Provider:      "okta",
				PrincipalType: "group",
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.id, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.want, parsePrincipalID(test.id))
		})
	}
}

func TestGetMemberNameFromPrincipal(t *testing.T) {
	t.Parallel()

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

	tests := []struct {
		id   string
		want string
	}{
		{
			id:   "local://user-2p7w6",
			want: "Default Admin (Local User)",
		},
		{
			id:   "okta_group://b4qkhsnliz",
			want: "DevOps (Okta Group)",
		},
		{
			id:   "okta_user://lfql6h5tmh",
			want: "lfql6h5tmh (Okta User)",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.id, func(t *testing.T) {
			t.Parallel()

			got := getMemberNameFromPrincipal(principals, test.id)
			assert.Equal(t, test.want, got)
		})
	}
}

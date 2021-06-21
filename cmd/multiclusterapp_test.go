package cmd

import (
	"testing"

	client "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"github.com/stretchr/testify/assert"
)

func TestFromMultiClusterAppAnswers(t *testing.T) {
	assert := assert.New(t)
	answerSlice := []client.Answer{
		{
			ProjectID: "c-1:p-1",
			Values: map[string]string{
				"var-1": "val1",
				"var-2": "val2",
			},
			ValuesSetString: map[string]string{
				"str-var-1": "str-val1",
				"str-var-2": "str-val2",
			},
		}, {
			ProjectID: "c-1:p-2",
			Values: map[string]string{
				"var-3": "val3",
			},
			ValuesSetString: map[string]string{
				"str-var-3": "str-val3",
			},
		}, {
			ClusterID: "c-1",
			Values: map[string]string{
				"var-4": "val4",
			},
			ValuesSetString: map[string]string{
				"str-var-4": "str-val4",
			},
		}, {
			ClusterID: "c-2",
			Values: map[string]string{
				"var-5": "val5",
			},
			ValuesSetString: map[string]string{
				"str-var-5": "str-val5",
			},
		}, {
			Values: map[string]string{
				"var-6": "val6",
			},
			ValuesSetString: map[string]string{
				"str-var-6": "str-val6",
			},
		},
	}

	answers, answersSetString := fromMultiClusterAppAnswers(answerSlice)
	assert.Equal(len(answers), 6)
	assert.Equal(answers["c-1:p-1:var-1"], "val1")
	assert.Equal(answers["c-1:p-1:var-2"], "val2")
	assert.Equal(answers["c-1:p-2:var-3"], "val3")
	assert.Equal(answers["c-1:var-4"], "val4")
	assert.Equal(answers["c-2:var-5"], "val5")
	assert.Equal(answers["var-6"], "val6")

	assert.Equal(len(answersSetString), 6)
	assert.Equal(answersSetString["c-1:p-1:str-var-1"], "str-val1")
	assert.Equal(answersSetString["c-1:p-1:str-var-2"], "str-val2")
	assert.Equal(answersSetString["c-1:p-2:str-var-3"], "str-val3")
	assert.Equal(answersSetString["c-1:str-var-4"], "str-val4")
	assert.Equal(answersSetString["c-2:str-var-5"], "str-val5")
	assert.Equal(answersSetString["str-var-6"], "str-val6")
}

func TestGetReadableTargetNames(t *testing.T) {
	assert := assert.New(t)
	clusters := map[string]client.Cluster{
		"c-1": {
			Name: "cn-1",
		},
		"c-2": {
			Name: "cn-2",
		},
	}
	projects := map[string]client.Project{
		"c-1:p-1": {
			Name: "pn-1",
		},
		"c-1:p-2": {
			Name: "pn-2",
		},
		"c-2:p-3": {
			Name: "pn-3",
		},
		"c-2:p-4": {
			Name: "pn-4",
		},
	}
	targets := []client.Target{
		{
			ProjectID: "c-1:p-1",
		},
		{
			ProjectID: "c-1:p-2",
		},
		{
			ProjectID: "c-2:p-3",
		},
	}
	result := getReadableTargetNames(clusters, projects, targets)
	assert.Contains(result, "cn-1:pn-1")
	assert.Contains(result, "cn-1:pn-2")
	assert.Contains(result, "cn-2:pn-3")

	targets = []client.Target{
		{
			ProjectID: "c-0:p-0",
		},
	}
	result = getReadableTargetNames(clusters, projects, targets)
	assert.Contains(result, "c-0:p-0")
}

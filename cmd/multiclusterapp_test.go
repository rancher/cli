package cmd

import (
	"testing"

	"github.com/rancher/types/client/management/v3"
	"github.com/stretchr/testify/assert"
)

func TestFromMultiClusterAppAnswers(t *testing.T) {
	assert := assert.New(t)
	answers := []client.Answer{
		{
			ProjectID: "c-1:p-1",
			Values: map[string]string{
				"k1": "v1",
				"k2": "v2",
			},
		}, {
			ProjectID: "c-1:p-2",
			Values: map[string]string{
				"k3": "v3",
				"k4": "v4",
			},
		}, {
			ClusterID: "c-1",
			Values: map[string]string{
				"k5": "v5",
				"k6": "v6",
			},
		}, {
			ClusterID: "c-2",
			Values: map[string]string{
				"k7": "v7",
				"k8": "v8",
			},
		}, {
			Values: map[string]string{
				"k9":  "v9",
				"k10": "v10",
			},
		},
	}

	answerMap := fromMultiClusterAppAnswers(answers)
	assert.Equal(len(answerMap), 10)
	assert.Equal(answerMap["c-1:p-1:k1"], "v1")
	assert.Equal(answerMap["c-1:p-1:k2"], "v2")
	assert.Equal(answerMap["c-1:p-2:k3"], "v3")
	assert.Equal(answerMap["c-1:k5"], "v5")
	assert.Equal(answerMap["c-2:k7"], "v7")
	assert.Equal(answerMap["k9"], "v9")
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

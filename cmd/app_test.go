package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

const redisSample = `## Redis values examples
##
image:
  registry: docker.io
  repository: bitnami/redis
  tag: 4.0.11
cluster:
  enabled: true
  slaveCount: 1
rbac:
  create: false
  role:
    rules: []
persistence: {}
master:
  port: 6379
  args: ["redis-server","--maxmemory-policy volatile-ttl"]
  disableCommands: "FLUSHDB,FLUSHALL"
  livenessProbe:
    enabled: true
    initialDelaySeconds: 30
    periodSeconds: 10
    timeoutSeconds: 5
    successThreshold: 1
    failureThreshold: 5
  service:
    type: ClusterIP
    port: 6379
    annotations: {}
    loadBalancerIP:
  securityContext:
    enabled: true
    fsGroup: 1001
    runAsUser: 1001
  persistence:
    enabled: true
    path: /bitnami/redis/data
    subPath: ""
    accessModes:
    - ReadWriteOnce
    size: 8Gi
  statefulset:
    updateStrategy: RollingUpdate
configmap: |-
  # Redis configuration file
  bind 127.0.0.1
  port 6379
`

func TestValuesToAnswers(t *testing.T) {
	assert := assert.New(t)

	answers := map[string]interface{}{}
	values := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(redisSample), &values); err != nil {
		t.Error(err)
	}
	valuesToAnswers(values, answers)

	assert.Equal("docker.io", answers["image.registry"], "unexpected image.registry")
	assert.Equal(true, answers["cluster.enabled"], "unexpected cluster.enabled")
	assert.Equal("1", answers["cluster.slaveCount"], "unexpected cluster.slaveCount")
	assert.Equal(nil, answers["rbac.role.rules"], "unexpected rbac.role.rules")
	assert.Equal(nil, answers["persistence"], "unexpected persistence")
	assert.Equal("redis-server", answers["master.args[0]"], "unexpected master.args[0]")
	assert.Equal("--maxmemory-policy volatile-ttl", answers["master.args[1]"], "unexpected master.args[1]")
	assert.Equal("FLUSHDB,FLUSHALL", answers["master.disableCommands"], "unexpected master.disableCommands")
	assert.Equal(nil, answers["master.service.loadBalancerIP"], "unexpected master.service.loadBalancerIP")
	assert.Equal("ReadWriteOnce", answers["master.persistence.accessModes[0]"], "unexpected master.persistence.accessModes[0]")
	assert.Equal("# Redis configuration file\nbind 127.0.0.1\nport 6379", answers["configmap"], "unexpected configmap")
}

func TestGetExternalIDInVersion(t *testing.T) {
	assert := assert.New(t)

	got, err := updateExternalIDVersion("catalog://?catalog=library&template=cert-manager&version=v0.5.2", "v1.2.3")
	assert.Nil(err)
	assert.Equal("catalog://?catalog=library&template=cert-manager&version=v1.2.3", got)

	got, err = updateExternalIDVersion("catalog://?catalog=c-29wkq/clusterscope&type=clusterCatalog&template=mysql&version=0.3.8", "0.3.9")
	assert.Nil(err)
	assert.Equal("catalog://?catalog=c-29wkq/clusterscope&type=clusterCatalog&template=mysql&version=0.3.9", got)

	got, err = updateExternalIDVersion("catalog://?catalog=p-j9gfw/projectscope&type=projectCatalog&template=grafana&version=0.0.31", "0.0.30")
	assert.Nil(err)
	assert.Equal("catalog://?catalog=p-j9gfw/projectscope&type=projectCatalog&template=grafana&version=0.0.30", got)
}

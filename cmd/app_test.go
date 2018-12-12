package cmd

import (
	"gopkg.in/yaml.v2"
	"testing"
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
	answers := map[string]string{}
	values := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(redisSample), &values); err != nil {
		t.Error(err)
	}
	valuesToAnswers(values, answers)
	if got, want := answers["image.registry"], "docker.io"; got != want {
		t.Errorf("Want image.registry %q, got %q", want, got)
	}
	if got, want := answers["cluster.enabled"], "true"; got != want {
		t.Errorf("Want cluster.enabled %q, got %q", want, got)
	}
	if got, want := answers["cluster.slaveCount"], "1"; got != want {
		t.Errorf("Want cluster.slaveCount %q, got %q", want, got)
	}
	if got, want := answers["rbac.role.rules"], ""; got != want {
		t.Errorf("Want rbac.role.rules %q, got %q", want, got)
	}
	if got, want := answers["persistence"], ""; got != want {
		t.Errorf("Want persistence %q, got %q", want, got)
	}
	if got, want := answers["master.args[0]"], "redis-server"; got != want {
		t.Errorf("Want master.args[0] %q, got %q", want, got)
	}
	if got, want := answers["master.args[1]"], "--maxmemory-policy volatile-ttl"; got != want {
		t.Errorf("Want master.args[1] %q, got %q", want, got)
	}
	if got, want := answers["master.disableCommands"], "FLUSHDB,FLUSHALL"; got != want {
		t.Errorf("Want master.disableCommands %q, got %q", want, got)
	}
	if got, want := answers["master.service.loadBalancerIP"], ""; got != want {
		t.Errorf("Want master.service.loadBalancerIP %q, got %q", want, got)
	}
	if got, want := answers["master.persistence.accessModes[0]"], "ReadWriteOnce"; got != want {
		t.Errorf("Want master.persistence.accessModes[0] %q, got %q", want, got)
	}
	if got, want := answers["configmap"], "# Redis configuration file\nbind 127.0.0.1\nport 6379"; got != want {
		t.Errorf("Want configmap %q, got %q", want, got)
	}
}

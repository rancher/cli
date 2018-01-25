package cmd

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Sirupsen/logrus"
	rancher "github.com/rancher/go-rancher/v2"
	apiproxy "github.com/rancher/rancher-docker-api-proxy"
)

const (
	DefaultNamedPipeline = "//./pipe/docker-sock"
	PipelineNamePrefix   = "npipe://" + DefaultNamedPipeline
)

func getDockerHost(client *rancher.RancherClient, host string) (string, *apiproxy.Proxy, error) {
	dockerHost := getRandNPipe()
	logrus.Info(dockerHost)
	proxy := apiproxy.NewProxy(client, host, dockerHost)
	return dockerHost, proxy, nil
}

func getRandNPipe() string {
	rand.Seed(time.Now().Unix())
	return fmt.Sprintf(PipelineNamePrefix+"%09d", rand.Intn(1000000000))
}

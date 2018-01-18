//+build !windows

package cmd

import (
	"io/ioutil"
	"os"

	rancher "github.com/rancher/go-rancher/v2"
	apiproxy "github.com/rancher/rancher-docker-api-proxy"
)

func getDockerHost(client *rancher.RancherClient, host string) (string, *apiproxy.Proxy, error) {
	tempfile, err := ioutil.TempFile("", "docker-sock")
	if err != nil {
		return "", nil, err
	}
	defer os.Remove(tempfile.Name())

	if err := tempfile.Close(); err != nil {
		return "", nil, err
	}

	dockerHost := "unix://" + tempfile.Name()
	proxy := apiproxy.NewProxy(client, host, dockerHost)
	return dockerHost, proxy, nil
}

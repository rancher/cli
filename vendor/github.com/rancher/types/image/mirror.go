package image

import "strings"

var Mirrors = map[string]string{}

func Mirror(image string) string {
	orig := image
	if strings.HasPrefix(image, "weaveworks") {
		return image
	}

	image = strings.Replace(image, "gcr.io/google_containers", "rancher", 1)
	image = strings.Replace(image, "quay.io/coreos/", "rancher/coreos-", 1)
	image = strings.Replace(image, "quay.io/calico/", "rancher/calico-", 1)
	image = strings.Replace(image, "k8s.gcr.io/", "rancher/nginx-ingress-controller-", 1)
	image = strings.Replace(image, "plugins/docker", "rancher/jenkins-plugins-docker", 1)
	image = strings.Replace(image, "kibana", "rancher/kibana", 1)
	image = strings.Replace(image, "jenkins/", "rancher/jenkins-", 1)
	image = strings.Replace(image, "alpine/git", "rancher/alpine-git", 1)
	image = strings.Replace(image, "prom/", "rancher/prom-", 1)
	image = strings.Replace(image, "quay.io/pires", "rancher", 1)

	Mirrors[image] = orig
	return image
}

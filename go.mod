module github.com/rancher/cli

go 1.13

replace (
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.2
	k8s.io/client-go => k8s.io/client-go v0.17.2
)

require (
	github.com/c-bata/go-prompt v0.0.0-20180219161504-f329ebd2409d
	github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0
	github.com/ghodss/yaml v1.0.0
	github.com/gorilla/websocket v1.4.0
	github.com/grantae/certinfo v0.0.0-20170412194111-59d56a35515b
	github.com/hashicorp/go-version v1.1.0
	github.com/mattn/go-tty v0.0.0-20180219170247-931426f7535a // indirect
	github.com/patrickmn/go-cache v2.0.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/pkg/term v0.0.0-20160705081919-b1f72af2d630 // indirect
	github.com/rancher/norman v0.0.0-20200520181341-ab75acb55410
	github.com/rancher/types v0.0.0-20200528213132-b5fb46b1825d
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/urfave/cli v1.22.4
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/client-go v12.0.0+incompatible
)

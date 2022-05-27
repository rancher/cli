module github.com/rancher/cli

go 1.16

replace k8s.io/client-go => k8s.io/client-go v0.20.1

require (
	github.com/c-bata/go-prompt v0.2.6
	github.com/docker/docker v1.6.1
	github.com/ghodss/yaml v1.0.0
	github.com/gorilla/websocket v1.4.2
	github.com/grantae/certinfo v0.0.0-20170412194111-59d56a35515b
	github.com/hashicorp/go-version v1.2.1
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/rancher/norman v0.0.0-20200820172041-261460ee9088
	github.com/rancher/rancher/pkg/client v0.0.0-20211110212758-cc2b8beb1473
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli v1.22.5
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0 // indirect
	k8s.io/client-go v12.0.0+incompatible
)

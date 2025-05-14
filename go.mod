module github.com/rancher/cli

go 1.23.0

toolchain go1.23.7

replace (
	golang.org/x/net => golang.org/x/net v0.38.0 // CVE-2025-22872
	k8s.io/api => k8s.io/api v0.30.10
	k8s.io/apimachinery => k8s.io/apimachinery v0.30.10
	k8s.io/apiserver => k8s.io/apiserver v0.30.10
	k8s.io/client-go => k8s.io/client-go v0.30.10
	k8s.io/component-base => k8s.io/component-base v0.30.10
	k8s.io/kubernetes => k8s.io/kubernetes v1.30.10
	sigs.k8s.io/cli-utils => sigs.k8s.io/cli-utils v0.30.0
	sigs.k8s.io/structured-merge-diff/v4 => sigs.k8s.io/structured-merge-diff/v4 v4.4.3
)

require (
	github.com/ghodss/yaml v1.0.0
	github.com/grantae/certinfo v0.0.0-20170412194111-59d56a35515b
	github.com/hashicorp/go-version v1.2.1
	github.com/pkg/errors v0.9.1
	github.com/rancher/norman v0.3.1
	github.com/rancher/rancher/pkg/apis v0.0.0-20250514154247-fc47e69e4e3f
	github.com/rancher/rancher/pkg/client v0.0.0-20250514154247-fc47e69e4e3f
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.10.0
	github.com/tidwall/gjson v1.17.0
	github.com/urfave/cli v1.22.5
	golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56
	golang.org/x/oauth2 v0.29.0
	golang.org/x/sync v0.14.0
	golang.org/x/term v0.32.0
	golang.org/x/text v0.25.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/client-go v12.0.0+incompatible
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.20.5 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.55.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rancher/aks-operator v1.9.9-rc.1 // indirect
	github.com/rancher/eks-operator v1.9.9-rc.1 // indirect
	github.com/rancher/fleet/pkg/apis v0.10.11 // indirect
	github.com/rancher/gke-operator v1.9.9-rc.1 // indirect
	github.com/rancher/lasso v0.2.1 // indirect
	github.com/rancher/rke v1.6.10-rc.2 // indirect
	github.com/rancher/wrangler/v3 v3.1.0 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/time v0.7.0 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.32.2 // indirect
	k8s.io/apimachinery v0.32.2 // indirect
	k8s.io/apiserver v0.31.1 // indirect
	k8s.io/component-base v0.31.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20241105132330-32ad38e42d3f // indirect
	k8s.io/kubernetes v1.30.10 // indirect
	k8s.io/utils v0.0.0-20241104100929-3ea5e8cea738 // indirect
	sigs.k8s.io/cli-utils v0.37.2 // indirect
	sigs.k8s.io/json v0.0.0-20241010143419-9aa6b5e7a4b3 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.2 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

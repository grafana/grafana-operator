module github.com/integr8ly/grafana-operator/v3

go 1.13

require (
	github.com/Microsoft/hcsshim v0.8.9 // indirect
	github.com/blang/semver v3.5.1+incompatible
	github.com/brancz/gojsontoyaml v0.0.0-20191212081931-bf2969bbd742 // indirect
	github.com/brancz/kube-rbac-proxy v0.5.0 // indirect
	github.com/containerd/continuity v0.0.0-20200413184840-d3ef23f19fbb // indirect
	github.com/containerd/ttrpc v1.0.1 // indirect
	github.com/coreos/prometheus-operator v0.29.0 // indirect
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-bindata/go-bindata/v3 v3.1.3 // indirect
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/go-openapi/spec v0.19.9
	github.com/google/go-jsonnet v0.16.0
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/hashicorp/go-version v1.1.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/iancoleman/strcase v0.0.0-20190422225806-e506e3ef7365 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/jsonnet-bundler/jsonnet-bundler v0.3.1 // indirect
	github.com/kylelemons/godebug v0.0.0-20170820004349-d65d576e9348 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mikefarah/yq/v2 v2.4.1 // indirect
	github.com/mitchellh/hashstructure v0.0.0-20170609045927-2bca23e0e452 // indirect
	github.com/onsi/gomega v1.10.1 // indirect
	github.com/opencontainers/image-spec v1.0.2-0.20190823105129-775207bd45b6 // indirect
	github.com/openshift/api v3.9.0+incompatible
	github.com/openshift/prom-label-proxy v0.1.1-0.20191016113035-b8153a7f39f1 // indirect
	github.com/operator-framework/operator-registry v1.5.5 // indirect
	github.com/operator-framework/operator-sdk v0.8.2
	github.com/otiai10/copy v1.2.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.5.1 // indirect
	github.com/prometheus/procfs v0.0.11 // indirect
	github.com/rogpeppe/go-internal v1.5.0 // indirect
	github.com/sirupsen/logrus v1.5.0 // indirect
	github.com/thanos-io/thanos v0.11.0 // indirect
	go.etcd.io/bbolt v1.3.4 // indirect
	go.uber.org/zap v1.14.1 // indirect
	gomodules.xyz/jsonpatch/v3 v3.0.1 // indirect
	gopkg.in/yaml.v3 v3.0.0-20190905181640-827449938966 // indirect
	gotest.tools/v3 v3.0.2 // indirect
	helm.sh/helm/v3 v3.2.0 // indirect
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v0.18.2
	k8s.io/code-generator v0.18.6 // indirect
	k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	k8s.io/kube-state-metrics v1.7.2 // indirect
	k8s.io/kubectl v0.18.2 // indirect
	k8s.io/utils v0.0.0-20200603063816-c1c6865ac451 // indirect
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/controller-runtime v0.6.0
	sigs.k8s.io/controller-tools v0.1.11-0.20190411181648-9d55346c2bde // indirect
	sigs.k8s.io/kubebuilder v1.0.9-0.20200513134826-f07a0146a40b // indirect
)

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM

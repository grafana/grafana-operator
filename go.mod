module github.com/integr8ly/grafana-operator/v3

go 1.13

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.2.0
	github.com/google/go-jsonnet v0.16.0
	github.com/openshift/api v3.9.0+incompatible
	github.com/operator-framework/operator-sdk v0.18.2
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.20.6
	k8s.io/apimachinery v0.20.6
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v14.2.0+incompatible
	github.com/go-logr/logr => github.com/go-logr/logr v0.1.0
	k8s.io/client-go => k8s.io/client-go v0.20.6
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.1.0
)

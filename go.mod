module github.com/integr8ly/grafana-operator/v3

go 1.14

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/google/go-jsonnet v0.16.0
	github.com/openshift/api v3.9.0+incompatible
	github.com/operator-framework/operator-sdk v0.18.2
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.5.1
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.18.2
)

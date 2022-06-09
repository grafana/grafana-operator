module github.com/grafana-operator/grafana-operator/v4

go 1.16

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.4.0
	github.com/google/go-jsonnet v0.17.0
	github.com/openshift/api v3.9.0+incompatible
	github.com/operator-framework/operator-lib v0.9.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e // indirect
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	sigs.k8s.io/controller-runtime v0.10.0
)

replace github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2

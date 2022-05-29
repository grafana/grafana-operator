module github.com/grafana-operator/grafana-operator/v4

go 1.18

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.4.0
	github.com/google/go-jsonnet v0.17.0
	github.com/openshift/api v0.0.0-20180801171038-322a19404e37
	github.com/operator-framework/operator-lib v0.4.1
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
)

replace github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2

// Handle CVE-2022-27191
replace golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0 => golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4

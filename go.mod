module github.com/integr8ly/grafana-operator

go 1.15

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/go-logr/logr v0.3.0
	github.com/integr8ly/grafana-operator/v3 v3.7.0
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.4
	github.com/openshift/api v3.9.0+incompatible
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.8.0

)

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20210116050105-fc034b4b7616

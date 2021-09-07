# OpenShift CI

## Dockerfile.tools

Base image used on CI for all builds and test jobs.

### Build and Test

```shell
$ docker build -t registry.svc.ci.openshift.org/integr8ly/grafana-operator-base-image:latest - < Dockerfile.tools
$ IMAGE_NAME=registry.svc.ci.openshift.org/integr8ly/grafana-operator-base-image:latest test/run
operator-sdk version: "v0.12.0", commit: "2445fcda834ca4b7cf0d6c38fba6317fb219b469", go version: "go1.13.5 linux/amd64"
go version go1.13.5 linux/amd64
go mod tidy
...
SUCCESS!
```

# Contributing to grafana-operator

Thank you for investing your time in contributing to our project.

## Development

The operator uses unit tests and [Kuttl](https://kuttl.dev/) for e2e tests to make sure that the operator is working as intended, we use make to generate a number of docs and scripts for us.

### E2e tests using Kuttl

As mentioned above we use Kuttl to run e2e tests for the operator, we normally run Kuttl on [Kind](https://kind.sigs.k8s.io/)

The `make e2e` command will

```shell
# Build the container
VERSION=latest make docker-build
# Using kind load the locally built image to the kind cluster
kind load docker-image quay.io/grafana-operator/grafana-operator:v5.0.0
# Create grafana-operator-system namespace
kubectl create ns grafana-operator-system
# Run the Kuttl tests
VERSION=latest make e2e
```

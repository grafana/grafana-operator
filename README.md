# Grafana Operator

A Kubernetes Operator based on the Operator SDK for creating and managing Grafana instances.

# Current status

The Operator is functional and can deploy and manage a Grafana instance on Kubernetes and OpenShift. The following features are supported:

* Install Grafana to a namespace
* Import Grafana dashboards from the same or other namespaces
* Import Grafana datasources from the same namespace
* Install Plugins (panels) defined as dependencies of dashboards 

# Operator flags

The operator supports the following flags on startup:

* *--grafana-image*: overrides the Grafana image, defaults to `docker.io/grafana/grafana`.
* *--grafana-image-tag*: overrides the Grafana tag, defaults to `5.4.2`.
* *--scan-all*: watch for dashboards in all namespaces. This requires the the operator service account to have cluster wide permissions to `get`, `list`, `update` and `watch` dashboards. See `deploy/examples/cluster_roles`.
* *--openshift*: force the operator to use a [route](https://docs.openshift.com/container-platform/3.11/architecture/networking/routes.html) instead of an [ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/). Note that routes are only supported on OpenShift.

Flags can be passed as `args` to the container.

# Supported Custom Resources

The following Grafana resources are supported:

* Grafana
* GrafanaDashboard
* GrafanaDatasource

all custom resources use the api group `integreatly.org` and version `v1alpha1`.

## Grafana

Represents a Grafana instance. See [the documentation](./documentation/deploy_grafana.md) for a description of properties supported in the spec.

## GrafanaDashboard

Represents a Grafana dashboard and allows to specify required plugins. See [the documentation](./documentation/dashboards.md) for a description of properties supported in the spec.

## GrafanaDatasource

# Running locally

You can run the Operator locally against a remote namespace using the operator-sdk:

```sh
$ operator-sdk up local --namespace=<namespace> --operator-flags="<flags to pass>"
```
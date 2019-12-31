# Grafana Operator

A Kubernetes Operator based on the Operator SDK for creating and managing Grafana instances.

# Current status

The Operator is available on [Operator Hub](https://operatorhub.io/operator/grafana-operator).

It can deploy and manage a Grafana instance on Kubernetes and OpenShift. The following features are supported:

* Install Grafana to a namespace
* Import Grafana dashboards from the same or other namespaces
* Import Grafana datasources from the same namespace
* Install Plugins (panels) defined as dependencies of dashboards 

# Operator flags

The operator supports the following flags on startup.
See [the documentation](./documentation/deploy_grafana.md) for a full list.
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

Represents a Grafana datasource. See [the documentation](./documentation/datasources.md) for a description of properties supported in the spec.

# Running locally

You can run the Operator locally against a remote namespace using the operator-sdk:

Prerequisites:

* [operator-sdk](https://github.com/operator-framework/operator-sdk) installed
* kubectl pointing to the local context. [minikube](https://github.com/kubernetes/minikube) automatically sets the context to the local VM. If not you can use `kubectl config use <context>` or (if using the OpenShift CLI) `oc login -u <user> <url>`
* make sure to deploy the custom resource defination using the command ```kubectl create -f deploy/crds```

```sh
$ operator-sdk up local --namespace=<namespace> --operator-flags="<flags to pass>"
```

# Grafana features not yet supported in the operator

## Notifier provisioning

Grafana has provisioning support for multiple channels (notifiers) of alerts. The operator does currently not support this type of provisioning. An empty directory is mounted at the expected location to prevent a warning in the grafana log. This feature might be supported in the future. 

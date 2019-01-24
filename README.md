# Grafana Operator

A Kubernetes Operator based on the Operator SDK for creating Grafana instances and importing dashboards from namespaces.

# Current status

This is a PoC / alpha version. Most functionality is there but it is higly likely there are bugs and improvements needed.

# Supported Custom Resources

The following Grafana resources are supported:

* Grafana
* GrafanaDashboard

## Grafana

Represents a Grafana installation.

## GrafanaDashboard

Represents a Grafana dashboard. To be created in the namespace of the service the dashboard is associated with.

# Running locally

You can run the Operator locally against a remote namespace. The name of the namespace should be `application-monitoring`. To run the operator execute:

```sh
$ make setup/dep
$ make code/run
```
# Working with dashboards

This document describes how to create dashboards and manage plugins (panels).

## Dashboard properties

Dashboards are represented by the `GrafanaDashboard` custom resource. Examples can be found in `deploy/examples/dashboards`.

The following properties are accepted in the `spec`:

* *name*: The filename of the dashboard that gets mounted into a volume in the grafana instance. Not to be confused with `metadata.name`.
* *json*: Raw json string with the dashboard contents. Check the [official documentation](https://grafana.com/docs/reference/dashboard/#dashboard-json).
* *plugins*: A list of plugins required by the dashboard. They will be installed by the operator if not already present.

## Creating a new dashboard

By default the operator only watches for dashboards in it's own namespace. To watch for dashboards in other namespaces, the `--scan-all` flag must be passed.

 
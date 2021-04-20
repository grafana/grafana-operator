# Working with data sources

This document describes how to create data sources.

## Data source properties

Data sources are represented by the `GrafanaDataSource` custom resource. Examples can be found
in `deploy/examples/datasources`.

A single `GrafanaDataSource` resource can contain a list of data sources.

The following properties are accepted in the `spec`:

* *name*: The filename of the data source that gets mounted into a volume in the grafana instance. Not to be confused
  with `metadata.name`.
* *datasources*: A list of data source definitions. Check
  the [official documentation](https://grafana.com/docs/features/datasources/).

A data source accepts all properties
listed [here](https://grafana.com/docs/administration/provisioning/#example-datasource-config-file), but does not
support `apiVersion` and `deleteDatasources`.

To see how to install datasource plugins, see [Plugins](./plugins.md)
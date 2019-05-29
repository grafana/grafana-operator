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

To create a dashboard in the `grafana` namespace run:

```sh
$ kubectl create -f deploy/examples/dashboards/SimpleDashboard.yaml -n grafana
```

*NOTE*: it can take up to a minute until new dashboards are discovered by Grafana.

## Plugins

Dashboards can specify plugins (panels) they depend on. The operator will automatically install them.

You need to provide a name and a version for every plugins, e.g.:

```yaml
spec:
  name: "dummy"
  json: "{}"
  plugins:
    - name: "grafana-piechart-panel"
      version: "1.3.6"
    - name: "grafana-clock-panel"
      version: "1.0.2"
```

Plugins are installed from the [Grafana plugin registry](https://grafana.com/plugins).

## Dashboard discovery

The operator uses a list of [set based selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#resources-that-support-set-based-requirements) to discover dashboards by their [labels](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/). The `dashboardLabelSelector` property of the `Grafana` resource allows you to add selectors by which the dashboards will be filtered.

*NOTE*: If no `dashboardLabelSelector` is present, the operator will not discover any dashboards. The same goes for dashboards without labels, they will not be discovered by the operator. 

Every selector can have a list of `matchLabels` and `matchExpressions`. The rules inside a single selector will be **AND**ed, while the list of selectors is evaluated with **OR**. 

For example, the following selector:

```yaml
dashboardLabelSelector:
  - matchExpressions:
      - {key: app, operator: In, values: [grafana]}
      - {key: group, operator: In, values: [grafana]}
```

requires the dashboard to have two labels, `app` and `group` and each label is required to have a value of `grafana`.

To accept either, the `app` or the `group` label, you can write the selector in the following way:

```yaml
dashboardLabelSelector:
  - matchExpressions:
      - {key: app, operator: In, values: [grafana]}
  - matchExpressions:
      - {key: group, operator: In, values: [grafana]}          
```
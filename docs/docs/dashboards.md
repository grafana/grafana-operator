---
title: Dashboards
weight: 13
---

[Dashboards](https://grafana.com/docs/grafana/latest/dashboards/) is the core feature of Grafana and of course something that you can manage through the operator.

To view the entire configuration that you can do within dashboards, look at our [API documentation](../api/#grafanadashboardspec).

## Dashboard managment

You can configure dashboards as code many different ways.

- json
- gzipJson
- URL
- Jsonnet

### Json

A pure JSON representation of your Grafana dashboard.
Normally you would create your dashboard manually within Grafana, when you have come up with how you want the dashboard to look like, you export it as JSON,
grab the JSON using the export function in grafana and put inside the GrafanaDashboard CR.

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafanadashboard-sample
spec:
  resyncPeriod: 30s
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  json: >
    {
      "id": null,
      "title": "Simple Dashboard",
      "tags": [],
      "style": "dark",
      "timezone": "browser",
      "editable": true,
      "hideControls": false,
      "graphTooltip": 1,
      "panels": [],
      "time": {
        "from": "now-6h",
        "to": "now"
      },
      "timepicker": {
        "time_options": [],
        "refresh_intervals": []
      },
      "templating": {
        "list": []
      },
      "annotations": {
        "list": []
      },
      "refresh": "5s",
      "schemaVersion": 17,
      "version": 0,
      "links": []
    }
```

### gzipJson

It's just like JSON but instead of adding pure JSON to the dashboard CR you add a gzipped representation.
This allows you to do really **big** dashboards that which workaround the etcd maximum request size of 1,5 MiB.

To create a gzipped representation of your dashboards, assuming that you have saved it to disk, can be done through.

```shell
cat dashboard.json | gzip | base64 -w0
```

Take the output and put it in your GrafanaDashboard CR, for example:

```yaml
---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafanadashboard-gzipped
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  gzipJson: |-
    H4sIAAAAAAAAA4WQQU/DMAyF7/0VVc9MggMgcYV/AOKC0OQubmM1jSPH28Sm/XfSNJ1WcaA3f+/l+dXnqk5fQ6Z5qf3eubt5VlKHCTXvNAaH9RtE2zKI2fQnCgFNsxihj8n39V3mqD/zQwMyXE004ol95q3wMaIsEhpSaPMTlT0WasngK3sVdlN6By4uUi8Q7AezUwpJeig4gEe3ajItTfM5T5l0wuNUwfNx82RLg9nLhTeZXW4iAu2GVHcVNPEtByX2tyuzJtgJRrslrygHKJ3WsZhuCkq+X8c6ivrXDd6zwrLrX3vZP/3PY1yuHHcWR/hEiSlmutpzEQ5XdF+IIz+Uzpeq+gWtMMT1HwIAAA==
```

[Example documentation](../examples/dashboard_gzipped/readme).

### URL

Probably the easiest way to get started to add dashboards to your Grafana instances.
You can use a GitHub URL.

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafanadashboard-from-url
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  url: "https://raw.githubusercontent.com/grafana-operator/grafana-operator/master/examples/dashboard_from_url/dashboard.json"
```

Or why not use a finished dashboard from grafana.

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafanadashboard-from-grafana
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  url: "https://grafana.com/api/dashboards/1860/revisions/30/download"
```

[Example documentation](../examples/dashboard_from_url/readme).

### Jsonnet

TODO

## Plugins

[Plugins](https://grafana.com/grafana/plugins/) is a way to extend the grafana functionality in dashboards and datasources.

Plugins can be installed to grafana instances managed by the operator and be defined in both datasources and dashboards.

They **cannot** be installed using external grafana instances due to how the installation of plugins is done at the start of the instance using environment variables which is a built in feature in grafana.

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: keycloak-dashboard
spec:
  instanceSelector:
    matchLabels:
      dashboards: grafana
  plugins:
    - name: grafana-piechart-panel
      version: 1.3.9
  json: >
    {
      "__inputs": [
        {
          "name": "DS_PROMETHEUS",
          "label": "Prometheus",
          "description": "",
          "type": "datasource",
          "pluginId": "prometheus",
          "pluginName": "Prometheus"
        }
      ],
      . # and much more
      .
      .
    }
```

Look here for more examples on how to install [plugins](../examples/plugins/readme)

## Dashboard UID management

TODO

## Custom folders

If you want to add a dashboard to a specific folder you can do that through.

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafanadashboard-with-custom-folder
spec:
  folder: "Custom Folder"
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  url: "https://raw.githubusercontent.com/grafana-operator/grafana-operator/master/examples/dashboard_from_url/dashboard.json"
```

[Example documentation](../examples/dashboard_with_custom_folder/readme).

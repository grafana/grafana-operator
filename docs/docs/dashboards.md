---
title: Dashboards
weight: 13
---

[Dashboards](https://grafana.com/docs/grafana/latest/dashboards/) is the core feature of Grafana and of course something that you can manage through the operator.

To view all configuration you can do within dashboards look at our [API documentation](../api/#grafanadashboardspec).

## Dashboard managment

You can configure dashboards as code many different ways.

- json
- gzipJson
- URL
- Jsonnet

### Json

### gzipJson

It's just like json but instead of adding pure json to the dashboard CR you add a gziped representation.
This to be able to do really **big** dashboards that allows you to workaround etcd maximum request size of 1,5 MiB.

To create a gziped representation of your dashboards assuming that you have saved it to disk can be done through.

```shell
cat dashboard.json | gzip | base64 -w0
```

### URL

Probably the easiest way to get started to add dashboards to your Grafana instances.

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

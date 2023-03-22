---
title: Datasources
weight: 12
---

[Datasources](https://grafana.com/docs/grafana/latest/datasources/) is a basic part of grafana and of course you can manage your datasources through the grafana operator.

To view all configuration you can do within datasources look at our [API documentation](../api/#grafanadatasourcespec).

## Secret management

Since you don't want to save secrets in your git repository grafana have added a possibility to overwrite secrets using environment variables.

We have enabled this in the operator through `grafanadatasources.spec.secrets` where you define a secret name in the same namespace as the datasource and you can use the secret keys to get the value you need.

For example:

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: credentials
  namespace: grafana
stringData:
  PROMETHEUS_USERNAME: root
  PROMETHEUS_PASSWORD: secret
type: Opaque
---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDatasource
metadata:
  name: grafanadatasource-sample
spec:
  secrets:
    - credentials
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  datasource:
    name: prometheus
    type: prometheus
    access: proxy
    basicAuth: true
    url: http://prometheus-service:9090
    isDefault: true
    user: ${PROMETHEUS_USERNAME}
    jsonData:
      "tlsSkipVerify": true
      "timeInterval": "5s"
    secureJsonData:
      "password": ${PROMETHEUS_PASSWORD} # Notice the brakes around PROMETHEUS_PASSWORD
    editable: true
```

[Here](../examples/datasource_variables/readme) you can find a bigger example on how to use datasources with environment variables.

## Plugins

[Plugins](https://grafana.com/grafana/plugins/) is a way to extend the grafana functionality in dashboards and datasources.

Plugins can be installed to grafana instances managed by the operator and be defined in both datasources and dashboards.

They **cannot** be installed using external grafana instances due to how the installation of plugins is done at the start of the instance using environment variables which is a built in feature in grafana.

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDatasource
metadata:
  name: example-grafanadatasource
spec:
  datasource:
    access: proxy
    type: prometheus
    jsonData:
      timeInterval: 5s
      tlsSkipVerify: true
    name: Prometheus
    url: http://prometheus-service:9090
  instanceSelector:
    matchLabels:
      dashboards: grafana
  plugins:
    - name: grafana-clock-panel
      version: 1.3.0
```

Look here for more examples on how to install [plugins](../examples/plugins/readme)

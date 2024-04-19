---
title: Datasources
weight: 12
---

[Datasources](https://grafana.com/docs/grafana/latest/datasources/) is a basic part of grafana and of course you can manage your datasources through the grafana operator.

To view all configuration you can do within datasources look at our [API documentation](../api/#grafanadatasourcespec).

## Secret management

In case a datasource requires authentication, it is advised not to include credentials directly in `url`. Instead, it's better to rely on value substitution like in the example below.

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
  valuesFrom:
    - targetPath: "user"
      valueFrom:
        secretKeyRef:
          name: "credentials"
          key: "PROMETHEUS_USERNAME"
    - targetPath: "secureJsonData.password"
      valueFrom:
        secretKeyRef:
          name: "credentials"
          key: "PROMETHEUS_PASSWORD"
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
```

**NOTE:** The secret must exist in the same namespace as the datasource.

[Here](../examples/datasource_variables/readme) you can find a bigger example on how to use datasources with environment variables.

## Plugins

[Plugins](https://grafana.com/grafana/plugins/) is a way to extend the grafana functionality in dashboards and datasources.

Plugins can be installed to grafana instances managed by the operator and be defined in both datasources and dashboards.

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

**NOTE:** To make grafana install a plugin, the operator bootstraps a grafana instance with a custom value passed in `GF_INSTALL_PLUGINS` environment variable ([Install plugins in the Docker container](https://grafana.com/docs/grafana/latest/setup-grafana/installation/docker/#install-official-and-community-grafana-plugins)). Thus, there is no way for the operator to install a plugin in an external grafana instance.

Look here for more examples on how to install [plugins](../examples/plugins/readme)

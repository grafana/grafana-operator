---
title: Datasources
weight: 50
---

[Datasources](https://grafana.com/docs/grafana/latest/datasources/) is a basic part of grafana and of course you can manage your datasources through the grafana operator.

To view all configuration you can do within Datasources look at our [API documentation](/docs/api/#grafanadatasourcespec).

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
    - targetPath: "basicAuthUser"
      valueFrom:
        secretKeyRef:
          name: "credentials"
          key: "PROMETHEUS_USERNAME"
    - targetPath: "secureJsonData.basicAuthPassword"
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
    basicAuthUser: ${PROMETHEUS_USERNAME}
    jsonData:
      "tlsSkipVerify": true
      "timeInterval": "5s"
    secureJsonData:
      "basicAuthPassword": ${PROMETHEUS_PASSWORD} # Notice the braces around PROMETHEUS_PASSWORD
```

{{% alert title="Note" color="primary" %}}
The secret must exist in the same namespace as the datasource.
{{% /alert %}}

[Here](./datasource_variables/readme) you can find a bigger example on how to use datasources with environment variables.

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

{{% alert title="Note" color="primary" %}}
To make grafana install a plugin, the operator bootstraps a grafana instance with a custom value passed in `GF_INSTALL_PLUGINS` environment variable ([Install plugins in the Docker container](https://grafana.com/docs/grafana/latest/setup-grafana/installation/docker/#install-official-and-community-grafana-plugins)). Thus, there is no way for the operator to install a plugin in an external grafana instance.
{{% /alert %}}

Look here for more examples on how to install [plugins](./plugins/readme)

## Private data source connect (PDC)

[Private data source connect](https://grafana.com/docs/grafana-cloud/connect-externally-hosted/private-data-source-connect/), or PDC, is a way for you to establish a private, secured connection between a Grafana Cloud instance, or stack, and data sources secured within a private network.

To enable PDC on data sources configured through the Grafana operator, set the `enableSecureSocksProxy` and `secureSocksProxyUsername` fields in the `jsonData` field of the resource like this:
```.
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDatasource
metadata:
  name: grafanadatasource-sample
spec:
  instanceSelector:
    matchLabels:
      instance: grafanacloud-instance
  datasource:
    name: prometheus-pdc-operator
    type: prometheus
    access: proxy
    url: http://localhost:9090
    jsonData:
      "enableSecureSocksProxy": true
      "secureSocksProxyUsername": "<your-pdc-network-id>"
```

To find the PDC network ID, go to the *Connections / Private data source connect* page in your Grafana Cloud instance and select the network you want to connect to.

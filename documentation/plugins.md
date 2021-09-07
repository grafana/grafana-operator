# Plugins

## Install plugins as dependencies of datasources or dashboards

If a datasource or dashboard requires a plugin, it can be defined within a dashboard that it is used in, this will
install the plugin, the datasource can be parsed regardless of the fact if the plugin exist (yet). Once defined in a
dashboard, the operator should install the plugin and make it accessible to the datasource.
If a plugin already exists then the operator will not attempt another installation.

## Install plugins using Grafana env vars via the operator

The operator allows you to pass custom env vars to the grafana deployment. This means that you can set
the `GF_INSTALL_PLUGINS` flag, as described
in [install-official-and-community-grafana-plugins](https://grafana.com/docs/grafana/latest/installation/docker/#install-official-and-community-grafana-plugins)

These can be added to the `spec.deployment.envFrom` section of the Grafana CR. EG:

```yaml
spec:
  ...
  deployment:
    envFrom:
        - secretRef:
          name: grafana-env

```

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: grafana-env
  namespace: default
stringData:
  GF_INSTALL_PLUGINS: doitintl-bigquery-datasource 1.0.8
```

If the plugin doesn't install, try restarting the grafana deployment.

# Changes in version 3.0.0

This version includes the following changes:

* Installs Grafana 6.5.1 by default
* Dashboards are no longer stored in a ConfigMap. Instead they are now directly imported using the Grafana API.
* Dashboard and Datasource custom resources no longer need finalizers. This means they can be deleted at any time, even when the operator itself is no longer running.
* Updated reconciliation strategy that keeps all resources up to date at all times and allows for better configuration through the Grafana CR.
* Updated to [operator-sdk v0.12.0](https://github.com/operator-framework/operator-sdk/releases/tag/v0.12.0)
* Using Go modules instead of dep now 

# Upgrade from 1.x.x or 2.x.x

The CRDs need to be reapplied:

```sh
$ kubectl apply -f deploy/crds
```

# Caveats

Dashboards are now imported using the Grafana API. This requires basic auth to be enabled (which is the default). If turned off through the config, dashboards can no longer be imported.
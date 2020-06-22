# Support for multiple namespaces

The operator can import dashboards from either one, some or all namespaces. By default it will look for dashboards in all namespaces.
This is done by setting the `--scan-all` flag. The default for `--scan-all` is `True` meaning that dashboards from all namespaces will be reconciled in the Grafana CR, if this isn't desired behaviour for project specific purposes this flag needs to be set to `False` or a namespace/namespaces have to be specified with `--namespaces` for the desired namespaces to be watched for Grafana CRs.

The `Grafana` and `GrafanaDataSource` resources do not support multiple namespaces and are only reconciled if created in the operators namespace.

## Watching for dashboards in all namespaces

Set the `--scan-all` flag to watch for dashboards in all namespaces. Cluster wide permissions for the `grafana-operator` service account are required (see `deploy/cluster_roles`).

## Watching for dashboards in some namespaces

### 1. **namespace operator flag**
You can provide a comma separated list of watch namespaces using the `--namespaces` flag. The format is `--namespaces=<NS_1,NS_2,...,NS_N>`, for example: `--namespaces=grafana,dashboards,example_namespace`.
The same cluster wide permissions as for watching all namespaces are required.

***NOTE***: `--scan-all` and `--namespaces` are mutually exclusive. You can only use one at a time.

### 2. **dashboardNamespaceSelector**
You can also watch for dashboards in certain namespaces by using the dashboardNamespaceSelector in the Grafana CR. This watches for dashboards only in the Namespaces that have the specified namespace label. The format to specify labels is 
```
dashboardNamespaceSelector:
    matchLabels:
      key: value
```
***NOTE***: `--namespaces` and the `dashboardNamespaceSelector` are mutually exclusive and shoudlnt be used together
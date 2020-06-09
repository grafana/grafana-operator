# Changes in version 3.0.0

This version includes the following changes:

* Installs Grafana 6.5.1 by default
* Dashboards are no longer stored in a ConfigMap. Instead they are now directly imported using the Grafana API.
* Dashboard and Datasource custom resources no longer need finalizers. This means they can be deleted at any time, even when the operator itself is no longer running.
* Updated reconciliation strategy that keeps all resources up to date at all times and allows for better configuration through the Grafana CR.
* Updated to [operator-sdk v0.12.0](https://github.com/operator-framework/operator-sdk/releases/tag/v0.12.0)
* Using Go modules instead of dep now 

## Upgrade from 1.x.x or 2.x.x

There is no direct upgrade path from previous versions to 3.0.0. This is due to a change in the resource status fields: status is now a subresource and the format has changed.

To upgrade, the following steps need to be performed:

### 1. Create backups of your resources

Create a directory for your backups:

```shell script
$ mkdir grafana-restore
```

Create backups of all dashboards:

```shell script
$ kubectl get grafanadashboards -n <namespace> --selector='<label selector>' -oyaml > ./grafana-restore/dashboards.yaml 
```

Or, if you want to grab all dashboards in all namespaces, use:

```shell script
$ kubectl get grafanadashboards --all-namespaces --selector='<label selector>' -oyaml > ./grafana-restore/dashboards.yaml
```

Repeat those steps for `grafanadatasources` and `grafanas`.

### 2. Remove finalizers and status

Edit the backed up resources and remove the finalizers from dashboards and datasources as well as the status from all resources.

### 3. Uninstall Grafana and all resources

Remove the existing CRs:

```shell script
$ kubectl delete grafanadashboards --all -n <namespace>
$ kubectl delete grafanadatasources --all -n <namespace>
$ kubectl delete grafanas --all -n <namespace>
```

### 4. Update the operator deployment

Update the image in the deployment of the existing grafana operator to `quay.io/integreatly/grafana-operator:v3.0.0`.

You might have to remove the `grafana-operator-lock` configmap if the new operator doesn't reach ready state because of leader election.

### 5. Update CRDs and roles

```shell script
$ kubectl apply -f deploy/crds
$ kubectl apply -f deploy/roles -n <namespace>
```

If you were using the multi namespace support for dashboards, also reapply the cluster role and role binding:

```shell script
$ kubectl apply -f deploy/cluster_roles
```

### 6. Reinstall Grafana and all resources

```shell script
$ kubectl apply -f -/grafana-resore
```

## Caveats

Dashboards are now imported using the Grafana API. This requires basic auth to be enabled (which is the default). If turned off through the config, dashboards can no longer be imported.
# Upgrade

## Changes in version 4.0.0

* Operator-sdk updated to v1.3.0
* Installs Grafana 7.1.1 by default
* Dashboard deleted in the Grafana console will be automatically restored
* New `kustomize` based installation method, installs the operator in the namespace `grafana-operator-system`

## Changes in version 3.0.0

This version includes the following changes:

* Installs Grafana 6.5.1 by default
* Dashboards are no longer stored in a ConfigMap. Instead they are now directly imported using the Grafana API.
* Dashboard and Datasource custom resources no longer need finalizers. This means they can be deleted at any time, even when the operator itself is no longer running.
* Updated reconciliation strategy that keeps all resources up to date at all times and allows for better configuration through the Grafana CR.
* Updated to [operator-sdk v0.12.0](https://github.com/operator-framework/operator-sdk/releases/tag/v0.12.0)
* Using Go modules instead of dep now

## Upgrade from 3.x.x to 4.x.x

There is no direct upgrade path from previous versions to 4.0.0. This is due to an upgraded operator-sdk version and an update of the CRD definitions from v1beta1 to v1.

To upgrade, we recommend the following steps:

### Uninstall the previous version

To uninstall the Grafana Operator, either remove the deployment with the name `grafana-operator`, or, if installed through OLM, follow the appropriate steps to remove the subscription:

1) On OpenShift 4.x, you can uninstall the Operator via `Operators -> Installed Operators`
2) On Kubernetes, use kubectl to identify the subscription:

```shell
kubectl get subscriptions -n<operator namespace>
```

Then delete the subscription:

```shell
kubectl delete subscription <subscription name> -n<operator namespace>
```

__NOTE__: uninstalling the Grafana Operator will not remove your Grafana instance or your dashboards.

### Install 4.0.0

Install the Grafana Operator v4.0.0 either using OLM or by using the [kustomize based installer](./deploy_grafana.md#Kustomize).

The new Operator should discover the existing Grafana and Dashboard CRs and take over.

## Upgrade from 1.x.x or 2.x.x

There is no direct upgrade path from previous versions to 3.0.0. This is due to a change in the resource status fields: status is now a subresource and the format has changed.

To upgrade, the following steps need to be performed:

### 1. Create backups of your resources

Create a directory for your backups:

```shell script
mkdir grafana-restore
```

Create backups of all dashboards:

```shell script
kubectl get grafanadashboards -n <namespace> --selector='<label selector>' -oyaml > ./grafana-restore/dashboards.yaml
```

Or, if you want to grab all dashboards in all namespaces, use:

```shell script
kubectl get grafanadashboards --all-namespaces --selector='<label selector>' -oyaml > ./grafana-restore/dashboards.yaml
```

Repeat those steps for `grafanadatasources` and `grafanas`.

### 2. Remove finalizers and status

Edit the backed up resources and remove the finalizers from dashboards and datasources as well as the status from all resources.

### 3. Uninstall Grafana and all resources

Remove the existing CRs:

```shell script
kubectl delete grafanadashboards --all -n <namespace>
kubectl delete grafanadatasources --all -n <namespace>
kubectl delete grafanas --all -n <namespace>
```

### 4. Update the operator deployment

Update the image in the deployment of the existing grafana operator to `quay.io/integreatly/grafana-operator:v3.0.0`.

You might have to remove the `grafana-operator-lock` configmap if the new operator doesn't reach ready state because of leader election.

### 5. Update CRDs and roles

```shell script
kubectl apply -f deploy/crds
kubectl apply -f deploy/roles -n <namespace>
```

If you were using the multi namespace support for dashboards, also reapply the cluster role and role binding:

```shell script
kubectl apply -f deploy/cluster_roles
```

### 6. Reinstall Grafana and all resources

```shell script
kubectl apply -f -/grafana-resore
```

## Caveats

Dashboards are now imported using the Grafana API. This requires basic auth to be enabled (which is the default). If turned off through the config, dashboards can no longer be imported.

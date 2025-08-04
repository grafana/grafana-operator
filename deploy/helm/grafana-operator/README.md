---
title: "Helm installation"
linkTitle: "Helm installation"
---

# grafana-operator

[grafana-operator](https://github.com/grafana/grafana-operator) for Kubernetes to manage Grafana instances and grafana resources.

![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v5.19.0](https://img.shields.io/badge/AppVersion-v5.19.0-informational?style=flat-square)

## Installation

This is a OCI helm chart, helm started support OCI in version 3.8.0.

```shell
helm upgrade -i grafana-operator oci://ghcr.io/grafana/helm-charts/grafana-operator --version v5.19.0
```

Sadly helm OCI charts currently don't support searching for available versions of a helm [oci registry](https://github.com/helm/helm/issues/11000).

### Using Terraform

To install the helm chart using terraform, make sure you use the right values for `repository` and `name` as shown below:

```hcl
resource "helm_release" "grafana_kubernetes_operator" {
  name       = "grafana-operator"
  namespace  = "default"
  repository = "oci://ghcr.io/grafana/helm-charts"
  chart      = "grafana-operator"
  verify     = false
  version    = "v5.19.0"
}
```

## Upgrading

Helm does not provide functionality to update custom resource definitions. This can result in the operator misbehaving when a release contains updates to the custom resource definitions.
To avoid issues due to outdated or missing definitions, run the following command before updating an existing installation:

```shell
kubectl apply --server-side --force-conflicts -f https://github.com/grafana/grafana-operator/releases/download/v5.19.0/crds.yaml
```

The `--server-side` and `--force-conflict` flags are required to avoid running into issues with the `kubectl.kubernetes.io/last-applied-configuration` annotation.
By using server side apply, this annotation is not considered. `--force-conflict` allows kubectl to modify fields previously managed by helm.

## Development

For general and helm specific development instructions please read the [CONTRIBUTING.md](https://github.com/grafana/grafana-operator/blob/master/CONTRIBUTING.md)

## Out of scope

The chart won't support any configuration of grafana instances or similar. It's only meant to be used to install the grafana-operator.
Deployments of grafana instances using the CRs is supposed to be done outside of the chart.

Currently the plan is not to support networkpolicy. The operators support os diverse configuration that you have to support all options.
It's easier to just manage this configuration outside of the operator.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| additionalLabels | object | `{}` | additional labels to add to all resources |
| affinity | object | `{}` | pod affinity |
| clusterDomain | string | `""` | Sets the `CLUSTER_DOMAIN` environment variable, it defines how internal Kubernetes services managed by the operator are addressed. By default, this is empty, and internal services are addressed without a cluster domain specified, i.e., a relative domain name that will resolve regardless of if a custom domain is configured for the cluster. If you wish to have services addressed using their FQDNs, you can specify the cluster domain explicitly, e.g., "cluster.local" for the default Kubernetes configuration. |
| dashboard.annotations | object | `{}` | Annotations to add to the Grafana dashboard ConfigMap |
| dashboard.enabled | bool | `false` | Whether to create a ConfigMap containing a dashboard monitoring the operator metrics. Consider enabling this if you are enabling the ServiceMonitor. Optionally, a GrafanaDashboard CR can be manually created pointing to the Grafana.com dashboard ID 22785 https://grafana.com/grafana/dashboards/22785-grafana-operator/ The Grafana.com dashboard is maintained by the community and does not necessarily match the JSON definition in this repository. |
| dashboard.labels | object | `{}` | Labels to add to the Grafana dashboard ConfigMap |
| enforceCacheLabels | string | `"safe"` | Sets the `ENFORCE_CACHE_LABELS` environment variable, Allows to tweak how caching of various Kubernetes resources works inside the operator. Valid values are "off", "safe", and "all". When set to "off", all resources are cached (including Deployments, Services, Ingresses, and any other native resources that the operator interacts with), which results in much higher memory usage (essentially, grows with cluster size). When set to `safe`, ConfigMaps and Secrets are not cached, all other native resources are cached only when they have `app.kubernetes.io/managed-by: grafana-operator` label. The label is automatically set on all resources that are created/owned by the operator (applicable to any mode). When set to `all`, only resources that have `app.kubernetes.io/managed-by: grafana-operator` are cached. The caveat is that ConfigMaps and Secrets can be seen by the operator only if they have the label. Thus, usage of this mode requires more careful planning. |
| env | list | `[]` | Additional environment variables |
| extraObjects | list | `[]` | Array of extra K8s objects to deploy |
| fullnameOverride | string | `""` | Overrides the fully qualified app name. |
| image.pullPolicy | string | `"IfNotPresent"` | The image pull policy to use in grafana operator container |
| image.repository | string | `"ghcr.io/grafana/grafana-operator"` | grafana operator image repository |
| image.tag | string | `""` | Overrides the image tag whose default is the chart appVersion. |
| imagePullSecrets | list | `[]` | image pull secrets |
| isOpenShift | bool | `false` | Determines if the target cluster is OpenShift. Additional rbac permissions for routes will be added on OpenShift |
| leaderElect | bool | `true` | This is recommended in most scenarios, even when only running a single instance of the operator. |
| logging.encoder | string | `"console"` | Log encoding ("console", "json") |
| logging.level | string | `"info"` | Configure the verbosity of logging ("debug", "error", "info") |
| logging.time | string | `"rfc3339"` | Time encoding ("epoch", "iso8601", "millis", "nano", "rfc3339", "rfc3339nano") |
| maxConcurrentReconciles | int | `1` | Maximum number of concurrent reconciles per Custom Resource. |
| metricsService.metricsPort | int | `9090` | metrics service port |
| metricsService.pprofPort | int | `8888` | port for the pprof profiling endpoint |
| metricsService.type | string | `"ClusterIP"` | metrics service type |
| nameOverride | string | `""` | Overrides the name of the chart. |
| namespaceOverride | string | `""` | Overrides the namespace name. |
| namespaceScope | bool | `false` | If the operator should run in namespace-scope or not, if true the operator will only be able to manage instances in the same namespace |
| nodeSelector | object | `{}` | pod node selector |
| podAnnotations | object | `{}` | pod annotations |
| podLabels | object | `{}` | pod labels |
| podSecurityContext | object | `{}` | pod security context |
| priorityClassName | string | `""` | pod priority class name |
| rbac.create | bool | `true` | Specifies whether to create the ClusterRole and ClusterRoleBinding. If "namespaceScope" is true or "watchNamespaces" is set, this will create Role and RoleBinding instead. |
| replicas | int | `1` | The number of operators to run simultaneously. With leader election, only one instance reconciles CRs preventing duplicate reconciliations. Note: Multiple replicas increase stability, it does not increase throughput. |
| resources | object | `{}` | grafana operator container resources |
| securityContext.allowPrivilegeEscalation | bool | `false` | Whether to allow privilege escalation |
| securityContext.capabilities | object | `{"drop":["ALL"]}` | A list of capabilities to drop |
| securityContext.readOnlyRootFilesystem | bool | `true` | Whether to allow writing to the root filesystem |
| securityContext.runAsNonRoot | bool | `true` | Whether to require a container to run as a non-root user |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| serviceMonitor.additionalLabels | object | `{}` | Set of labels to transfer from the Kubernetes Service onto the target |
| serviceMonitor.enabled | bool | `false` | Whether to create a ServiceMonitor |
| serviceMonitor.interval | string | `"1m"` | Set how frequently Prometheus should scrape |
| serviceMonitor.metricRelabelings | list | `[]` | MetricRelabelConfigs to apply to samples before ingestion |
| serviceMonitor.relabelings | list | `[]` | Set relabel_configs as per https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config |
| serviceMonitor.scrapeTimeout | string | `"10s"` | Set timeout for scrape |
| serviceMonitor.targetLabels | list | `[]` | Set of labels to transfer from the Kubernetes Service onto the target |
| serviceMonitor.telemetryPath | string | `"/metrics"` | Set path to metrics path |
| tolerations | list | `[]` | pod tolerations |
| watchLabelSelectors | string | `""` | Sets the `WATCH_LABEL_SELECTORS` environment variable, it defines which CRs are watched according to their labels. By default, the operator watches all CRs. To make it watch only a subset of CRs, define the variable as a *stringified label selector*. See also: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/ Beware: Always label Grafana CRs before enabling to ensure labels are inherited. # Existing Secrets/ConfigMaps referenced in CRs also need to be labeled to continue working. |
| watchNamespaceSelector | string | `""` | Sets the `WATCH_NAMESPACE_SELECTOR` environment variable, it defines which namespaces the operator should be listening for based on a namespace label (e.g. `"environment: dev"`). By default, the operator watches all namespaces. To make it watch only its own namespace, check out `namespaceScope` option instead. |
| watchNamespaces | string | `""` | Sets the `WATCH_NAMESPACE` environment variable, it defines which namespaces the operator should be listening for (e.g. `"grafana, foo"`). By default, the operator watches all namespaces. To make it watch only its own namespace, check out `namespaceScope` option instead. |

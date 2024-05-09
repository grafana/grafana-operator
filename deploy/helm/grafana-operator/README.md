---
title: "Helm installation"
linkTitle: "Helm installation"
---

# grafana-operator

[grafana-operator](https://github.com/grafana/grafana-operator) for Kubernetes to manage Grafana instances and grafana resources.

![Version: 0.1.4](https://img.shields.io/badge/Version-0.1.4-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v5.9.0](https://img.shields.io/badge/AppVersion-v5.9.0-informational?style=flat-square)

## Installation

This is a OCI helm chart, helm started support OCI in version 3.8.0.

```shell
helm upgrade -i grafana-operator oci://ghcr.io/grafana/helm-charts/grafana-operator --version v5.9.0
```

Sadly helm OCI charts currently don't support searching for available versions of a helm [oci registry](https://github.com/helm/helm/issues/11000).

## Upgrading

Helm does not provide functionality to update custom resource definitions. This can result in the operator misbehaving when a release contains updates to the custom resource definitions.
To avoid issues due to outdated or missing definitions, run the following command before updating an existing installation:

```shell
kubectl apply --server-side --force-conflicts -f https://github.com/grafana/grafana-operator/releases/download/v5.9.0/crds.yaml
```

The `--server-side` and `--force-conflict` flags are required to avoid running into issues with the `kubectl.kubernetes.io/last-applied-configuration` annotation.
By using server side apply, this annotation is not considered. `--force-conflict` allows kubectl to modify fields previously managed by helm.

## Development

For general and helm specific development instructions please read the [CONTRIBUTING.md](../../../CONTRIBUTING.md)

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
| env | list | `[]` | Additional environment variables |
| fullnameOverride | string | `""` | Overrides the fully qualified app name. |
| image.pullPolicy | string | `"IfNotPresent"` | The image pull policy to use in grafana operator container |
| image.repository | string | `"ghcr.io/grafana/grafana-operator"` | grafana operator image repository |
| image.tag | string | `""` | Overrides the image tag whose default is the chart appVersion. |
| imagePullSecrets | list | `[]` | image pull secrets |
| isOpenShift | bool | `false` | Determines if the target cluster is OpenShift. Additional rbac permissions for routes will be added on OpenShift |
| leaderElect | bool | `false` | If you want to run multiple replicas of the grafana-operator, this is not recommended. |
| metricsService.metricsPort | int | `9090` | metrics service port |
| metricsService.type | string | `"ClusterIP"` | metrics service type |
| nameOverride | string | `""` | Overrides the name of the chart. |
| namespaceOverride | string | `""` | Overrides the namespace name. |
| namespaceScope | bool | `false` | If the operator should run in namespace-scope or not, if true the operator will only be able to manage instances in the same namespace |
| nodeSelector | object | `{}` | pod node selector |
| podAnnotations | object | `{}` | pod annotations |
| podSecurityContext | object | `{}` | pod security context |
| priorityClassName | string | `""` | pod priority class name |
| resources | object | `{}` | grafana operator container resources |
| securityContext | object | `{"capabilities":{"drop":["ALL"]},"readOnlyRootFilesystem":true,"runAsNonRoot":true}` | grafana operator container security context |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| serviceMonitor | object | `{"additionalLabels":{},"enabled":false,"interval":"1m","metricRelabelings":[],"relabelings":[],"scrapeTimeout":"10s","targetLabels":[],"telemetryPath":"/metrics"}` | Enable this to use with Prometheus Operator |
| serviceMonitor.additionalLabels | object | `{}` | Set of labels to transfer from the Kubernetes Service onto the target |
| serviceMonitor.enabled | bool | `false` | When set true then use a ServiceMonitor to configure scraping |
| serviceMonitor.interval | string | `"1m"` | Set how frequently Prometheus should scrape |
| serviceMonitor.metricRelabelings | list | `[]` | MetricRelabelConfigs to apply to samples before ingestion |
| serviceMonitor.relabelings | list | `[]` | Set relabel_configs as per https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config |
| serviceMonitor.scrapeTimeout | string | `"10s"` | Set timeout for scrape |
| serviceMonitor.targetLabels | list | `[]` | Set of labels to transfer from the Kubernetes Service onto the target |
| serviceMonitor.telemetryPath | string | `"/metrics"` | Set path to metrics path |
| tolerations | list | `[]` | pod tolerations |
| watchNamespaceSelector | string | `""` | Sets the WATCH_NAMESPACE_SELECTOR environment variable, it defines which namespaces the operator should be listening for based on label and key value pair added on namespace kind. By default it's all namespaces. |
| watchNamespaces | string | `""` | Sets the WATCH_NAMESPACE environment variable, it defines which namespaces the operator should be listening for. By default it's all namespaces, if you only want to listen for the same namespace as the operator is deployed to look at namespaceScope. |

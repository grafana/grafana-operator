---
title: "Helm installation"
linkTitle: "Helm installation"
---

# grafana-operator

[grafana-operator](https://github.com/grafana-operator/grafana-operator) for Kubernetes to manage Grafana instances and grafana resources.

![Version: 0.1.1](https://img.shields.io/badge/Version-0.1.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 5.0.0](https://img.shields.io/badge/AppVersion-5.0.0-informational?style=flat-square)

## Installation

This is a OCI helm chart, helm started support OCI in version 3.8.0.

```shell
helm upgrade -i grafana-operator oci://ghcr.io/grafana-operator/helm-charts/grafana-operator --version v5.0.0
```

Sadly helm OCI charts currently don't support searching for available versions of a helm [oci registry](https://github.com/helm/helm/issues/11000).

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
| affinity | object | `{}` |  |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"ghcr.io/grafana-operator/grafana-operator"` |  |
| image.tag | string | `""` |  |
| imagePullSecrets | list | `[]` |  |
| kubeRbacProxy.args[0] | string | `"--secure-listen-address=0.0.0.0:8443"` |  |
| kubeRbacProxy.args[1] | string | `"--upstream=http://127.0.0.1:8080/"` |  |
| kubeRbacProxy.args[2] | string | `"--logtostderr=true"` |  |
| kubeRbacProxy.args[3] | string | `"--v=10"` |  |
| kubeRbacProxy.image | string | `"gcr.io/kubebuilder/kube-rbac-proxy:v0.8.0"` | The image to use in kubeRbacProxy container |
| kubeRbacProxy.imagePullPolicy | string | `"IfNotPresent"` |  |
| kubeRbacProxy.resources | object | `{}` |  |
| kubeRbacProxy.service.port | int | `8443` |  |
| kubeRbacProxy.service.type | string | `"ClusterIP"` |  |
| leaderElect | bool | `false` | If you want to run multiple replicas of the grafana-operator, this is not recommended. |
| nameOverride | string | `""` |  |
| namespaceScope | bool | `false` | If the operator should run in namespace-scope or not, if true the operator will only be able to manage instances in the same namespace |
| nodeSelector | object | `{}` |  |
| podAnnotations | object | `{}` |  |
| podSecurityContext | object | `{}` |  |
| priorityClassName | string | `""` | podPriorityClass |
| resources | object | `{}` |  |
| securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| securityContext.readOnlyRootFilesystem | bool | `true` |  |
| securityContext.runAsNonRoot | bool | `true` |  |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| tolerations | list | `[]` |  |
| watchNamespaces | string | `""` | Sets the WATCH_NAMESPACES environment variable, it defines which namespaces the operator should be listening for. By default it's all namespaces, if you only want to listen for the same namespace as the operator is deployed to look at namespaceScope. |

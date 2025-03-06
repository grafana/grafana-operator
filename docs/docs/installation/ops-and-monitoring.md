---
title: "Operations and Observability"
linkTitle: "Operations and Observability"
description: "Operations and monitoring of the Grafana Operator itself"
---

# Grafana Operator Operational Monitoring

## Dashboard

By default we provide a Dashboard that leverages the operator metrics to give a overview of the operator state. This dashboard is based on the [Grafana Operator Dashboard (ID 22785)](https://grafana.com/grafana/dashboards/22785-grafana-operator/).

The dashboard provides insights into the operator's performance and health within your Kubernetes cluster. You can enable it by using Helm, or alternatively, by manually creating the dashboard via a Grafana.com link or the provided JSON definition.

### Enabling the Dashboard with Helm

To enable the dashboard using Helm, you must set the following values in your `values.yaml`:

```yaml
dashboard:
  enabled: true
```

When enabled, Helm will create a ConfigMap containing the Grafana Operator dashboard definition as part of your deployment.

The Dashboard by default gets created inside a ConfigMap to avoid the chicken and egg problems that arrise when we use operator-managed Custom Resources in the same chart that is deploying the Custom Resource Definitions.

If your Grafana instance has a sidecar looking for ConfigMaps containing dashboards, then leveraging the `dashboard.additionalLabels` and `dashboard.annotations` values we can manipulate the dashboard ConfigMap labels and annotations so that the sidecar can find it and load the dashboard into Grafana.

#### Helm Values Breakdown

| **Value**                        | **Type**  | **Default**  | **Description**                                                     |
|----------------------------------|-----------|-------------|----------------------------------------------------------------------|
| `dashboard.additionalLabels`     | object    | `{}`        | Labels to add to the Grafana dashboard ConfigMap.                    |
| `dashboard.annotations`          | object    | `{}`        | Annotations to add to the Grafana dashboard ConfigMap.               |
| `dashboard.enabled`              | bool      | `false`     | Set to `true` to create a ConfigMap containing the dashboard.        |

### Alternative Methods for Loading the Dashboard

If you are not using Helm, or prefer using the Operator Custom Resources, then you can load the Grafana dashboard in the following ways:

#### Option 1: Load the Dashboard from Grafana.com

You can manually create a GrafanaDashboard Custom Resource (CR) to point to the Grafana.com dashboard.

To create the Grafana dashboard, use the following Custom Resource definition:

```yaml
---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafana-operator-dashboard-id
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  grafanaCom:
    id: 22785
    revision: 2
```

#### Option 2: Use the JSON Definition

Alternatively, you can use the JSON definition of the dashboard. The `files/dashboard.json` file contains the complete dashboard definition.

To use it, create a `GrafanaDashboard` Custom Resource (CR) pointing to the JSON definition:

```yaml
---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafana-operator-dashboard-remote-json
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  url: "https://raw.githubusercontent.com/grafana/grafana-operator/refs/heads/master/deploy/helm/grafana-operator/files/dashboard.json"
```

#### Option 3: Use the Helm generated ConfigMap in a CR

If we enable ConfigMap creation through the Helm values but cannot rely on the sidecar approach to load the dashboard into Grafana, we can still create a `GrafanaDashboard` Custom Resource (CR) that references the ConfigMap.

To use this approach, create a `GrafanaDashboard` CR that points to the existing ConfigMap:

```yaml
---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafana-operator-dashboard-from-configmap
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  configMapRef:
    name: grafana-operator-dashboard
    key: grafana-operator.json
```

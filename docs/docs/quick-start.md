---
title: Quick Start
weight: 5
---

This guide will help you get started using the Grafana Operator.

To follow along, you will need:

* Cluster admin access to a Kubernetes cluster
* The Helm CLI installed locally

## Installing the Operator

To install the Grafana Operator in your Kubernetes cluster, Run the following command in your terminal:

```bash
helm upgrade -i grafana-operator oci://ghcr.io/grafana/helm-charts/grafana-operator --version {{<param version>}}
```

This will install the grafana operator in the current namespace.

For a detailed installation guide, check out [the installation documentation]({{<relref installation>}}).

## Creating a Grafana instance

The `Grafana` custom resource describes the deployment of a single Grafana instance. A minimal starting point looks like this:

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  name: grafana
  labels:
    dashboards: "grafana"
spec:
  config:
    security:
      admin_user: root
      admin_password: secret
```

Save this to a file called `grafana.yaml` and apply it using `kubectl apply -f grafana.yaml`

This creates a Grafana deployment in the same namespace as the `Grafana` resource.

Run `kubectl get pods -w` to see the status of the deployment. Once the `grafana-deployment` pod is ready, continue to the next step.

## Adding a data source

The operator uses the `GrafanaDatasource` resource to configure data sources in Grafana.

An example data source connecting to a Prometheus backend is provided below:

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDatasource
metadata:
  name: prometheus
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  datasource:
    name: prom1
    type: prometheus
    access: proxy
    url: http://prometheus-service:9090
    isDefault: true
    jsonData:
      'tlsSkipVerify': true
      'timeInterval': "5s"
```

Save the file to `datasource.yaml` and apply it using `kubectl apply -f datasource.yaml`

It is important that the `instanceSelector` matches the `metadata.labels` field of the Grafana instance.
Otherwise the data source will not show up.

## Adding a dashboard

Adding a dashboard works the same way data sources do.
The `GrafanaDashboard` resource provides the dashboard specification as well as the `instanceSelector`.

Dashboards can be defined through JSON, jsonnet, a grafana.com dashboard catalog ID or a remote url.
For more information, check out our examples.

Using JSON directly embedded into the resource is the simplest approach.
The following resource defines a dashboard with a single panel:

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: example-dashboard
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  json: >
    {
      "annotations": {},
      "editable": true,
      "fiscalYearStartMonth": 0,
      "graphTooltip": 0,
      "id": 222,
      "links": [],
      "panels": [
        {
          "gridPos": {
            "h": 3,
            "w": 8,
            "x": 8,
            "y": 0
          },
          "id": 1,
          "options": {
            "code": {
              "language": "plaintext",
              "showLineNumbers": false,
              "showMiniMap": false
            },
            "content": "# Greetings from the Grafana Operator!",
            "mode": "markdown"
          },
          "type": "text"
        }
      ],
      "schemaVersion": 39,
      "tags": [],
      "time": {
        "from": "now-6h",
        "to": "now"
      },
      "timeRangeUpdatedDuringEditOrView": false,
      "timepicker": {},
      "timezone": "browser",
      "title": "Example Dashboard",
      "weekStart": ""
    }
```

Save the file to `dashboard.yaml` and apply it using `kubectl apply -f dashboard.yaml`

You will find the dashboard in a folder with the same name as your namespace.

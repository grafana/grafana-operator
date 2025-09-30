---
title: Library panels
weight: 14
---

[Library panels](https://grafana.com/docs/grafana/latest/dashboards/build-dashboards/manage-library-panels/)
are a reusable panel that you can use in any dashboard. When you make a change
to a library panel, that change propagates to all instances of where the panel
is used. Library panels streamline reuse of panels across multiple dashboards.

To find all possible configuration options, look at our [API documentation](/docs/api/#grafanalibrarypanelspec).

## Library panel management

Library panels are managed in almost exactly the same way as [dashboards](../dashboards), and as such,
most of the features of dashboard management in the operator can be used for library panels:

* Variety of content sources (JSON, gzip, URL, jsonnet)
* Automatic plugin provisioning
* Content caching
* Stable `uid` derived from the CR, if not defined explicitly on the content model
* Environment variable interpolation
* Mapping datasource references

Here is an example library panel that shows a graph of container restarts, using data
fetched from Google Cloud Monitoring:

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaLibraryPanel
metadata:
  name: gke-container-restarts-over-time
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  datasources:
    - inputName: "DS_GCP"
      datasourceName: "gcp"
  json: >
    {
      "__inputs": [
        {
          "name": "DS_GCP",
          "label": "gcp",
          "description": "",
          "type": "datasource",
          "pluginId": "stackdriver",
          "pluginName": "Google Cloud Monitoring"
        }
      ],
      "name": "GKE Container Restarts (Over Time)",
      "uid": "gke-container-restarts-over-time",
      "kind": 1,
      "datasource": {
        "type": "stackdriver",
        "uid": "${DS_GCP}"
      },
      "description": "",
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "palette-classic"
          },
          "custom": {
            "axisBorderShow": false,
            "axisCenteredZero": false,
            "axisColorMode": "text",
            "axisLabel": "",
            "axisPlacement": "auto",
            "barAlignment": 0,
            "drawStyle": "bars",
            "fillOpacity": 0,
            "gradientMode": "none",
            "hideFrom": {
              "legend": false,
              "tooltip": false,
              "viz": false
            },
            "insertNulls": false,
            "lineInterpolation": "linear",
            "lineWidth": 1,
            "pointSize": 5,
            "scaleDistribution": {
              "type": "linear"
            },
            "showPoints": "auto",
            "spanNulls": false,
            "stacking": {
              "group": "A",
              "mode": "normal"
            },
            "thresholdsStyle": {
              "mode": "off"
            }
          },
          "mappings": [],
          "min": 0,
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {
                "color": "green",
                "value": null
              },
              {
                "color": "red",
                "value": 80
              }
            ]
          }
        },
        "overrides": []
      },
      "options": {
        "legend": {
          "calcs": [],
          "displayMode": "list",
          "placement": "bottom",
          "showLegend": false
        },
        "tooltip": {
          "mode": "single",
          "sort": "none"
        }
      },
      "targets": [
        {
          "aliasBy": "",
          "datasource": {
            "type": "stackdriver",
            "uid": "gcp"
          },
          "hide": false,
          "queryType": "timeSeriesQuery",
          "refId": "CPU Usage Time",
          "timeSeriesQuery": {
            "graphPeriod": "",
            "projectName": "$cloud_account_id",
            "query": "fetch k8s_container\n| metric 'kubernetes.io/container/restart_count'\n| filter resource.cluster_name == '${k8s_cluster_name}' && resource.namespace_name == '${k8s_namespace_name}'\n  && metadata.system.top_level_controller_name == '${k8s_workload_name}'\n  && resource.container_name == '${k8s_container_name}'\n| delta\n| group_by [resource.pod_name],\n    [value: sum(value.restart_count)]"
          }
        }
      ],
      "title": "Container Restarts",
      "type": "timeseries"
    }
```

### Referencing library panels in dashboards

Once a library panel is provisioned in the Grafana instance, it can either be used to create ad-hoc dashboards
not backed by code, or you can reference it by uid in dashboard definitions. All that is needed is a panel
with a `libraryPanel` field that refers to the panel by uid, e.g.:

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: jason-test-gcp-librarypanel
spec:
  folder: gcp
  instanceSelector:
    matchLabels:
      env: stg
      region: us-central1
  json: >
    {
      "id": null,
      "title": "Simple Dashboard",
      "tags": [],
      "style": "dark",
      "timezone": "browser",
      "editable": true,
      "hideControls": false,
      "graphTooltip": 1,
      "panels": [
        {
          "libraryPanel": {
            "uid": "gke-container-restarts-over-time"
          }
        }
      ],
      "time": {
        "from": "now-6h",
        "to": "now"
      },
      "timepicker": {
        "time_options": [],
        "refresh_intervals": []
      },
      "templating": {
        "list": []
      },
      "annotations": {
        "list": []
      },
      "refresh": "5s",
      "schemaVersion": 17,
      "version": 0,
      "links": []
    }
```

You can adjust positioning/size on the other panel fields just as you would for a normal panel.

{{% alert title="Note" color="primary" %}}
library panels likely depend on dashboard variables to be defined, and they do _not_
automatically configure any. Any dashboards that utilize a library panel must define any required
variable names.
{{% /alert %}}

## Exporting library panels from Grafana

If you have library panels already configured in your environment and would like to bring them
in as GrafanaLibraryPanel resources, the simplest way is to query Grafana's Library Elements API:

```shell
curl "$GRAFANA_HOST/api/library-elements"
```

This will list all library panels; if you know a panel's UID you can target it explicitly:

```shell
curl "$GRAFANA_HOST/api/library-elements/$PANEL_UID"
```

However, the form of the library element contains many additional metadata fields that are irrelevant
to the Grafana operator. If you have `jq` installed, you can run this script to "fix up" the library
panel representation so it can be imported as JSON:

```json
jq '.model | del(.libraryPanel) | .datasource.type as $pluginId | .datasource.uid = "${" + (.__inputs | map(select(.pluginId == $pluginId))[0].name) + "}"'
```

This script does the following:

* only looks at the library panel content model located in the `.model` field
* attempts to rewrite datasource references so they refer to `__inputs`, which are used for late-binding
  the uid of the datasource installed and present on the target Grafana instance
* removes the `.libraryPanel` self-reference

Hopefully, in the future, it will be simpler to export library panels from Grafana for sharing.

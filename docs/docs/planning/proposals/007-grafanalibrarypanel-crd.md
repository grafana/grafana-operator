---
title: "GrafanaLibraryPanel CRD"
linkTitle: "GrafanaLibraryPanel CRD"
---

## Summary

Add a new GrafanaLibraryPanel CRD so that the operator can provision Grafana
[Library Panels](https://grafana.com/docs/grafana/latest/dashboards/build-dashboards/manage-library-panels/)
automatically to the Grafana instance(s) it is managing. Currently library
panels must be manually configured via the UI (or through some other bespoke
process that utilizes the Grafana API.)

## Info

status: Implemented

## Motivation

Library Panels allow Grafana administrators to provide a set of pre-defined
panels that can be easily imported into existing/new dashboards by Grafana users.
When the Library Panel definition is updated, all dashboards that reference the
panel use the new definition, making them very useful for reducing boilerplate
and rolling out cross-cutting improvements with greater ease.

Providing support for managing Library Panels as custom resources would make it
simpler to scale provisioning these across multiple Grafana instances and make
it easier to track changes to these assets via GitOps workflows, if desired.

## Verification

- Create e2e tests for the operator creating GrafanaLibraryPanels from baseline definition
- Create e2e tests to verify that Library Panels are not pruned if referenced by some Dashboard

## Current solution

It's possible for end-users to convert existing Dashboard panels into Library Panels
via the Grafana UI. Similar to other assets configured in Grafana, if there is no
persistent storage layer, the changes are tied to the lifecycle of the ephemeral storage
(e.g., the pod tmp space.)

## Proposal

Given that Library Panels are very similar to Dashboards in how they are modeled and
provisioned, use the GrafanaDashboard CRD as inspiration, and add a new GrafanaLibraryPanel
CRD that supports defining a panel model as JSON, YAML, or via external link. At
reconcile time, the Library Panel model definition will be provisioned to Grafana
via the [Library Element API](https://grafana.com/docs/grafana/latest/developers/http_api/library_element/).

### Example Custom Resource

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaLibraryPanel
metadata:
  name: grpc-server-success-rate
spec:
  allowCrossNamespaceImport: true
  contentCacheDuration: 1m0s
  folderRef: folder-ref
  instanceSelector:
    matchLabels:
      env: dev
      region: us-central1
  url: http://assets.example.com/library-panels/grpc-server-success-rate.json
```

### Referencing in Dashboards

#### Static referencing

When GrafanaLibraryPanels are configured for a Grafana instance, users will be able to browse
and utilize library panels in dashboards via the GUI workflows. They can export the JSON model
of the dashboard and store the model in a GrafanaDashboard CR, if desired. Library panels are
referenced by UID, e.g.:

```json
"panels": [
  {
    "gridPos": {
      "h": 8,
      "w": 12,
      "x": 0,
      "y": 0
    },
    "id": 2,
    "libraryPanel": {
      "uid": "ddkbyftwuqfpcf",
      "name": "gRPC Server Success Rate"
    }
  }
]
```

The UID of the library panel is defined on the `model` field within the GrafanaLibraryPanel
custom resource. It can thus be provided by the CR owner. By setting the UID to a stable value
as opposed to letting Grafana autogenerate it, it's possible to provision both Library Panels
and Dashboards that reference them as CRs.

#### Dynamic referencing

Similar to how Datasources can be dynamically referenced via the
`.spec.datasources` field, we can extend the GrafanaDashboard CR to similarly
allow mapping library panels via a new `.spec.libraryPanels` field structured
very similarly. The CR author can specify a replacement string and a reference
to the library panel deployed itself as a CR.

When reconciling the GrafanaDashboard, the controller will fetch the dependent
GrafanaLibraryPanel CRs, look up their UIDs, and then perform the string
replacement.

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grpc-overview
spec:
  libraryPanels:
    - inputName: "LP_GRPC_SUCCESS_RATE"
      libraryPanelRef: "grpc-server-success-rate"
  allowCrossNamespaceImport: true
  contentCacheDuration: 1m0s
  folderRef: folder-ref
  instanceSelector:
    matchLabels:
      env: dev
      region: us-central1
  json: >
    {
      "__inputs": [
        {
          "name": "DS_PROMETHEUS",
          "label": "prometheus",
          "description": "",
          "type": "datasource",
          "pluginId": "prometheus",
          "pluginName": "Prometheus"
        }
      ],
      "__elements": {},
      "__requires": [
        {
          "type": "grafana",
          "id": "grafana",
          "name": "Grafana",
          "version": "9.1.6"
        },
        {
          "type": "datasource",
          "id": "prometheus",
          "name": "Prometheus",
          "version": "1.0.0"
        },
        {
          "type": "panel",
          "id": "timeseries",
          "name": "Time series",
          "version": ""
        }
      ],
      "annotations": {
        "list": [
          {
            "builtIn": 1,
            "datasource": {
              "type": "grafana",
              "uid": "-- Grafana --"
            },
            "enable": true,
            "hide": true,
            "iconColor": "rgba(0, 211, 255, 1)",
            "name": "Annotations & Alerts",
            "target": {
              "limit": 100,
              "matchAny": false,
              "tags": [],
              "type": "dashboard"
            },
            "type": "dashboard"
          }
        ]
      },
      "editable": true,
      "fiscalYearStartMonth": 0,
      "graphTooltip": 1,
      "id": null,
      "links": [],
      "liveNow": false,
      "panels": [
        {
          "gridPos": {
            "h": 8,
            "w": 12,
            "x": 0,
            "y": 0
          },
          "id": 2,
          "libraryPanel": {
            "uid": "${LP_GRPC_SERVER_SUCCESS_RATE}",
            "name": "gRPC Server Success Rate"
          }
        }
      ],
      "refresh": "5s",
      "schemaVersion": 37,
      "style": "dark",
      "tags": [],
      "templating": {
        "list": []
      },
      "time": {
        "from": "now-6h",
        "to": "now"
      },
      "timepicker": {
        "refresh_intervals": [],
        "time_options": []
      },
      "timezone": "browser",
      "title": "gRPC Overview",
      "uid": "10605d5d-da66-4bb0-a77a-2635fae239f7",
      "version": 3,
      "weekStart": ""
    }
```

### Generalize Dashboard content read/cache logic

Much of the logic in the GrafanaDashboard controller can be re-used with minimal changes,
such as the mechanisms that read the dashboard JSON model from various sources (configmap,
JSON string, external URL) and the caching mechanism.

### Deleting panels that are referenced by Dashboards

If a Library Panel managed by a GrafanaLibraryPanel CR is deleted, we will check to
ensure it does not have any existing references to any dashboards. If there are
references, we do not continue with deletion and log a message indicating that
manual cleanup should first be done. We will manage this with a finalizer.

---
title: Update GrafanaDashboard folder management strategy
linkTitle: Update GrafanaDashboard folder strategy
---

## Summary

Update the folder creation/selection mechanism in the GrafanaDashboard CRD by introducing the possibility to target a GrafanaFolder CR declared in the cluster.

This document contains the complete design required to support targeting a GrafanaFolder has a reference. This method would permit to handle the subfolder feature from Grafana (version >= 10) if combined with the [proposal on GrafanaFolder CRD]({{< relref "004-grafanafolder-parent-folder-management.md" >}}).

The suggested new features are:

- Permits to use a `folderRef` field to target a folder deployed thanks to the GrafanaFolder CR.
- Permits to use `folderUID` field to target an existing folder in Grafana.

## Info

status: Decided <!-- TODO: update when validated by maintainers -->

## Motivation

The GrafanaDashboard CRD currently deployed by the operator permits to:

- create a folder at the root level of Grafana if it does not exist. If no folder is declared in the operator, the namespace of the folder will be used as a name.
- create a dashboard in the previously created folder.

With the arrival of Grafana 11, subfolder functionality becomes stable and is enabled by default. However, this functionality is not handled by the grafana-operator yet and its implementation create unwanted behavior by creating an intermediate level folder (the namespace folder) when declared.

## Proposal

The proposal of this pull request is to update the GrafanaDashboard CR to:

- add a field `folderRef` which permits to target an existing GrafanaFolder CR where the dashboard will be created.
- add a field `folderUID` which permits to target an existing folder in Grafana by it UID.

### Proposal 1: Target an existing folder in Grafana using its reference in the operator

The first proposal is to add a folderRef field in the GrafanaDashboard CRD which could request the GrafanaFolder CR by it name, retrieve the id and ask for the creation to the Grafana API.
We can find the Grafana API reference for:

- [The creation](https://grafana.com/docs/grafana/latest/developers/http_api/dashboard/#create--update-dashboard)

We have found that the grafana golang sdk already handle the folderUID field natively for dashboard creation: <https://github.com/grafana/grafana-openapi-client-go/blob/main/models/save_dashboard_command.go>

With this field implemented, the GrafanaDashboard CRD would look like this:

```yaml
---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaFolder
metadata:
  name: parent-folder
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  title: "Parent"

---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: cr_dashboard_ref
  namespace: namespace
  labels:
    dashboards: "grafana"
spec:
  folderRef: "parent-folder"
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  json: |
    {
      "id": null,
      "title": "Simple Dashboard REF",
      "tags": [],
      "style": "dark",
      "timezone": "browser",
      "editable": true,
      "hideControls": false,
      "graphTooltip": 1,
      "panels": [],
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

[abc](https://github.com/aaaaaaxcvxcvxcvxcv)

### Proposal 2: Target an existing folder in Grafana using its UID

The second proposal is to add a folderUID field in the GrafanaDashboard CRD which could request the Grafana API to create and reconcile the folder.
We can find the Grafana API reference for:

- [The creation](https://grafana.com/docs/grafana/latest/developers/http_api/dashboard/#create--update-dashboard)

We have found that the grafana golang sdk already handle the folderUID field [natively](https://github.com/grafana/grafana-openapi-client-go/blob/main/models/import_dashboard_request.go) for dashboard creation.

With this field implemented, the CRD GrafanaDashboard would look like this:

```yaml
---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: cr_dashboard_uid
  namespace: namespace
  labels:
    dashboards: "grafana"
spec:
  folderUID: "test123456789" # ID of the parent Folder in Grafana
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  json: |
    {
      "id": null,
      "title": "Simple Dashboard UID",
      "tags": [],
      "style": "dark",
      "timezone": "browser",
      "editable": true,
      "hideControls": false,
      "graphTooltip": 1,
      "panels": [],
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

> Note: this field is mutually exclusive with the folderRef presented in the first proposal.

### Backwards Compatibility

If one of the new fields (`folderUID`/`folderRef`) is set, the legacy `folder` field is ignored and the behavior of `folderRef`/`folderUID` is executed.

In all other cases, the operator behaves the same as in previous versions

```bash
# first evolution
.
L existing_folder
  L cr_grafanafolder_created_folder
    L cr_dashboard
```

The manifests will look like this:

```yaml
---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaFolder
metadata:
  name: cr_grafanafolder_created_folder
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  title: "cr_grafanafolder_created_folder"
  parentFolderUID: "<existing_folder_uid>"
---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: cr_dashboard_3
  labels:
    dashboards: "grafana"
spec:
  folderRef: cr_grafanafolder_created_folder
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  json: |
    {
      "id": null,
      "title": "Simple Dashboard 3",
      "tags": [],
      "style": "dark",
      "timezone": "browser",
      "editable": true,
      "hideControls": false,
      "graphTooltip": 1,
      "panels": [],
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

## Impact on the already existing CRD

By keeping the behaviour backwards compatible, existing CRDs will continue to be supported in the same way.

## Decision Outcome

`folderUid` and `folderRef` will be added to the `GrafanaDashboard` CRD. The behaviour is as follows:

- If no folder specification (old or new) is set -> use namespace folder
- If `folder` is set and none of the other fields is set (backwards compat. case) -> Create the folder
- If `folderUID` or `folderRef` is set -> don't create the folder
- If `folderUID` or `folderRef` **AND** folder is set, the new fields take priority -> don't create the folder

## Related discussions

- [PR 1564](https://github.com/grafana/grafana-operator/pull/1564)
- [Issue 1222](https://github.com/grafana/grafana-operator/issues/1222)
- [Issue 1514](https://github.com/grafana/grafana-operator/issues/1514)

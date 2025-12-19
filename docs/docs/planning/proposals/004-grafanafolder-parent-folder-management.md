---
title: "GrafanaFolder parent folder management"
linkTitle: "GrafanaFolder parent folder management"
---

## Summary

Introduce the possibility to handle a subfolder in an already declared folder (requires Grafana version >= 10).

This document contains the complete design required to support configuring subfolder by extending the GrafanaFolder CRD.

The suggested new features are:
- Permits to declare the parentFolderUID property to target an already existing folder in Grafana as a parent folder.
- Permits to declare the parentFolderRef property to target an already existing GrafanaFolder CR handled by the grafana-operator as a parent folder.

## Info

status: Implemented

## Motivation

With the arrival of Grafana 11, subfolder functionality becomes stable and is enabled by default.

The GrafanaFolder CRD currently deployed by the operator permits to create a folder at the root level of Grafana but does not handle the subfolder feature yet.

## Proposal

This document proposes to extend the GrafanaFolder to add fields to allow targeting of:
* An existing folder in Grafana using its UID
* An existing GrafanaFolder CR deployed by the grafana-operator

### Proposal 1: Target an existing folder in Grafana using its UID

The first proposal is to add a parentFolderUID field in the GrafanaFolder CRD which could request the Grafana API to create and reconcile the folder.

We can find the Grafana API reference for:
* [The creation](https://grafana.com/docs/grafana/latest/developers/http_api/folder/#create-folder)
* [The reconciliation loop](https://grafana.com/docs/grafana/latest/developers/http_api/folder/#create-folder)

We have found that the grafana golang sdk already handle the parentUID field natively for folder creation: https://github.com/grafana/grafana-openapi-client-go/blob/main/models/create_folder_command.go#L25

With this field implemented, the CRD GrafanaFolder would look like this:

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaFolder
metadata:
  name: subfolder
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  title: "SubFolder"
  parentFolderUID: "test123456789" # ID of the parent Folder in Grafana
```

> Note: this field is mutually exclusive with the parentFolderRef presented in the second proposal.

### Proposal 2: Target an existing GrafanaFolder deployed by the grafana-operator

The second proposal is to add a parentFolderRef field in the GrafanaFolder CRD which could request an already deployed GrafanaFolder CR in the current Kubernetes namespace.

To do this, we could use a similar mechanism like the GrafanaFolder discovery used in the code [here](https://github.com/grafana/grafana-operator/blob/master/controllers/grafanafolder_controller.go#L162)

With this field implemented, the CRD GrafanaFolder would look like this:

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaFolder
metadata:
  name: parent-folder
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  title: "Parent Folder"

---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaFolder
metadata:
  name: sub-folder
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  title: "SubFolder"
  parentFolderRef: parent-folder
```

> Note: this field is mutually exclusive with the parentFolderUID presented in the first proposal.

## Impact on the already existing CRD

Currently, we don't see any impact in the GrafanaFolder resource.
Tests will require an update to Grafana 10 or 11.

## Decision Outcome

We will implement this proposal by combining both options, similar to the way we handle this topic in alert rule groups. Both `parentFolderUID` and `parentFolderRef` will be available (mutually exclusivity validated by the CR spec).

## Related discussions

- [Issue 1222](https://github.com/grafana/grafana-operator/issues/1222)

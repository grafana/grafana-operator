---
title: Subfolders and Hierarchies
linkTitle: Subfolders
weight: 10
tags:
  - Folders
---

With the arrival of Grafana 10, you can create a complete `Folder` hierarchy in Grafana. To do this, you have two choices:

## Create named folder references

with the `.spec.parentFolderRef` it's possible to use existing `GrafanaFolders` in the same namespace.

```yaml
---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaFolder
metadata:
  name: folder-with-parent
spec:
  title: parent folder
  instanceSelector:
    matchLabels:
      dashboards: "grafana"

---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaFolder
metadata:
  name: subfolder-in-parent
spec:
  title: subfolder
  # GrafanaFolder parent folder reference
  parentFolderRef: folder-with-parent
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
```

## Reference ParentFolder with UIDs

Under a lot of circumstances, referencing parent folders using the name of the CR is not an option.

It is therefore possible to use the UID of the parent folder directly.
This allows the operator to create subfolders inside externally managed folders created/managed by other means.

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaFolder
metadata:
  name: existing-folder-uid
spec:
  # parent Folder uid to retrieve in your Grafana
  parentFolderUID: "3e7b4fe1-ca90-4125-a8ab-06567c1971b5"
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
```

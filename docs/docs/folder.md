---
title: Folders
weight: 15
---

Dashboard folders is a good way to manage your dashboards.

In a standard scenario, a folder with default settings gets created through a `GrafanaDashboard` CR. It either matches the Kubernetes namespace a dashboard exist in or `spec.folder` field of the CR.

If you need more control over folders (such as RBAC settings), it can be achieved through a `GrafanaFolder` CR.

{{% alert title="Note" color="primary" %}}
The folder reconciler attempts to take control over existing folders if a folder with the same name already exists.
This can lead to unpredictable behavior and will be removed in future versions.
Please take care to make sure any managed folders are created by the operator.
{{% /alert %}}

To view all configuration you can do within folders, look at our [API documentation](../api/#grafanafolderspec).

## Folder with custom title

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaFolder
metadata:
  name: test-folder
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  # If title is not defined, the value will be taken from metadata.name
  title: custom title
```

## Subfolder into an already existing Folder

With the arrival of Grafana 10, you can create a complete `Folder` hierarchy in Grafana. To do this, you have two choices:

* Use an existing GrafanaFolder CR as reference with the `parentFolderRef` field:

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

* Select an existing Folder in Grafana using its UID:

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


## Folder with custom permissions

When `permissions` value is empty/absent, a folder is created with default permissions. In all other scenarios, a raw JSON is passed to Grafana API, and it's up to Grafana to interpret it.

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaFolder
metadata:
  name: test-folder
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  permissions: |
    {
      "items": [
        {
          "role": "Admin",
          "permission": 4
        },
        {
          "role": "Editor",
          "permission": 2
        }
      ]
    }
```

**NOTE:** When an empty JSON is passed (`permissions: "{}"`), the access is stripped for everyone except for Admin (default Grafana behaviour).

## Updating folder configuration

When updating a folder's configuration, please do so via the `GrafanaFolder` CR. Any changes made directly to the folder in Grafana will be overwritten by the Grafana operator as per the configuration defined in the CR.

### Fixing conflicts
Avoid changing the name of the folder directly in Grafana as this may result in conflicts as the operator attempts to reconcile it (see [issues/1171](https://github.com/grafana/grafana-operator/issues/1171)). When this occurs, any subsequent updates will not occur, resulting in the rest of the folder configuration to not be updated to the correct state (i.e. permissions).

To resolve this, delete the folder in Grafana that matches the `uid` specified in the `GrafanaFolder` CR. This should allow the Grafana operator to update the remaining folder with the correct UID and permissions on the next reconcile.

> **IMPORTANT NOTE**: Before deleting the folder, please take care of moving any manually created dashboards out of it as the operator will not be able to recreate them.

---
title: Folders
weight: 20
---

Folders are great way to group your Dashboards and a requirement to for AlertRuleGroups.

Resources like `GrafanaDashboards` and `GrafanaAlertRuleGroups` can reference folders using `.spec.folderUID`.
Which is useful when folders are managed through other means than the operator.

But creating a `GrafanaFolder` allows other CRs to use `.spec.folderRef` which enables named references.

To view all configuration options for Folders, look at our [API documentation](/docs/api/#grafanafolderspec).

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

{{% alert title="Note" color="primary" %}}
The folder reconciler attempts to take control over existing folders if a folder with the same name already exists and `.spec.uid` is _empty/absent_.
This can lead to unpredictable behavior and will be removed in future versions.
Please take care to make sure any managed folders are created by the operator.
{{% /alert %}}


{{% alert title="Warning" color="secondary" %}}
Before deleting a GrafanaFolder CR, take care of moving any manually created dashboards and alerts out as the operator _**will delete them.**_
{{% /alert %}}

---
title: "Folder Permissions"
linkTitle: "Folder Permissions"
weight: 20
---

Create a folder with custom title and [permissions](https://grafana.com/docs/grafana/v8.4/http_api/folder_permissions/#update-permissions-for-a-folder).

When `.spec.permissions` is empty/absent, a folder is created with default permissions. In all other scenarios, the raw JSON is passed to Grafana API, and it's up to Grafana to interpret it.

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}


{{% alert title="Note" color="primary" %}}
When an empty JSON is passed, `.spec.permissions: "{}"`, the access is stripped for everyone except for Admin (default Grafana behaviour).
{{% /alert %}}

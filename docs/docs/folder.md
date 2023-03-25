---
title: Folders
weight: 14
---

Dashboard folders is a good way to manage your dashboards.

In a standard scenario, a folder with default settings gets created through a `GrafanaDashboard` CR. It either matches the Kubernetes namespace a dashboard exist in or `spec.folder` field of the CR.

If you need more control over folders (such as RBAC settings), it can be achieved through a `GrafanaFolder` CR.

To view all configuration you can do within folders, look at our [API documentation](../api/#grafanafolderspec).

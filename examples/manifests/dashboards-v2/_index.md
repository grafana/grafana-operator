---
title: "Dashboards V2"
---

While not natively supported by the `GrafanaDashboard` resource, dashboards using the new [dashboard V2 schema](https://grafana.com/whats-new/2025-04-11-new-dashboards-schema/) can be provisioned using the `GrafanaManifest` resource.

To do this, wrap the dashboard in a `GrafanaManifest` under the `spec.template` field like in the example below:

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

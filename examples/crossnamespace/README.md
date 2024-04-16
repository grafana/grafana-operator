---
title: "Cross namespace"
linkTitle: "Cross namespace"
---


If you want your dashboard or datasource to be able to be used by a grafana instance in another namespace you need to set `spec.allowCrossNamespaceImport: true`.

In the resources file you will find examples how to do this.

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

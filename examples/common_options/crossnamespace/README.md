---
title: "Cross namespace"
linkTitle: "Cross namespace"
---


If you want your resources synced to a grafana instance outside their namespace you need to set `.spec.allowCrossNamespaceImport: true`.

In the resources file you will find examples how to do this.

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

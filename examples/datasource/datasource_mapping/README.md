---
title: "Datasource mapping"
linkTitle: "Datasource mapping"
---

This example shows how to map an internal data source to an input from an exported dashboard.

There are two datasources in this example.
This to visualize that we can use multiple datasources and specify which `datasourceName` to use when overwriting the `DS_PROMETHEUS` config in the grafana dashboard json.

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

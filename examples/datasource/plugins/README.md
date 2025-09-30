---
title: "Datasource Plugins"
linkTitle: "Datasource Plugins"
tags:
  - Plugins
---
This won't work on external grafana instances.
As the operator does not manage the instance itself and thus we can't set the environment variable that we use to make grafana install plugins for us.

{{< readfile file="datasource.yaml" code="true" lang="yaml" >}}

{{% alert title="Note" color="primary" %}}
A plugin doesn't have to be pinned to a specific version. If it's set to `latest` instead, Grafana will install the latest available version upon start. Please, keep in mind that the grafana-operator doesn't track new plugin releases, so it's up to an administrator to make sure Grafana pods are occasionally recreated (in most setups, it happens naturally due to dynamic nature of Kubernetes workloads).
{{% /alert %}}

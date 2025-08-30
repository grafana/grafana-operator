---
title: "Grafana plugins"
linkTitle: "Grafana plugins"
---
This won't work on external grafana instances.
Due to the operator don't own and thus we can't set the environment variable that we use to make grafana install plugins for us.

{{< readfile file="dashboard.yaml" code="true" lang="yaml" >}}
{{< readfile file="datasource.yaml" code="true" lang="yaml" >}}

**NOTE**: A plugin doesn't have to be pinned to a specific version. If it's set to `latest` instead, Grafana will install the latest available version upon start. Please, keep in mind that the grafana-operator doesn't track new plugin releases, so it's up to an administrator to make sure Grafana pods are occasionally recreated (in most setups, it happens naturally due to dynamic nature of Kubernetes workloads).

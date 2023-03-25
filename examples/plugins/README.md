---
title: "Grafana plugins"
linkTitle: "Grafana plugins"
---
This won't work on external grafana instances.
Due to the operator don't own and thus we can't set the environment variable that we use to make grafana install plugins for us.

{{< readfile file="dashboard.yaml" code="true" lang="yaml" >}}
{{< readfile file="datasource.yaml" code="true" lang="yaml" >}}

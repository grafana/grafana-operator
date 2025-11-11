---
title: Service Accounts
weight: 80
---

`GrafanaServiceAccounts`(SA) are unique compared to other resources as the security implications are higher.
In order to avoid roque accounts in Grafana instances, the creation and matching of `SA` is intentionally limited.

Any `SA` matches exactly one Grafana instance through the `.spec.instanceName` field.
The `instanceName` being equal `.metadata.name`.

Additionally, the matching is limited within the same namespace as shown below.

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

The operator will then create a Secret for each token in `.spec.tokens`.

{{< readfile file="result.yaml" code="true" lang="yaml" >}}

For all possible configuration options, take a look at the [GrafanaAPI reference](/docs/api/#grafanaserviceaccountspec).

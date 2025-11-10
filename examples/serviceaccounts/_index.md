---
title: Service Accounts
weight: 80
---

`GrafanaServiceAccounts`(SA) are quite unique compared to other resources to the security implications that can arise.
In order to avoid roque accounts in Grafana instances, the creation and matching of `SA` is at the time of writing intentionally limited.

Any `SA` matches exactly one Grafana instance through the `.spec.instanceName` field.
The `instanceName` being equal `.metadata.name`.

Additionally, the matching is limited within the same namespace as shown below.

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

The operator will then create a Secret for each token in `.spec.tokens`.

{{< readfile file="result.yaml" code="true" lang="yaml" >}}

For all possible configuration options, take a look at the [GrafanaAPI reference](/docs/api/#grafanaserviceaccountspec).

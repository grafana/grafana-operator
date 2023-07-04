---
author: "Edvin 'NissesSenap' Norling"
date: 2023-05-27
title: "v4 to v5 migration"
linkTitle: "v4 to v5 migration"
description: "How to migrate from grafana-operator v4 to v5"
---

We are getting close to releasing version 5 of grafana-operator.

Version 5 comes with a number of breaking changes in API which were required to introduce support for managing multiple Grafana instances (including externally-hosted), arbitrary Grafana configuration (we no longer need to explicitly support each configuration field in grafana.ini), and other useful features. A more comprehensive list of changes can be found in our [intro blog post]({{< ref "/blog/v5-intro.md" >}}), and full documentation - [here]({{< ref "/docs/">}}).

Unfortunately, there is no automated way to migrate between the API versions due to fundamental changes in how CRs are described and managed.

The best way to prepare for the migration though would be by looking at our [example library]({{< ref "/docs/examples/">}}) and adjusting those examples according to your needs. For anything that is not covered there, our [API documentation]({{< ref "/docs/api.md">}}) should be very helpful.

**NOTE:** Since v4 and v5 use different `apiVersion` in CRs (`integreatly.org/v1alpha1` -> `grafana.integreatly.org/v1beta1`), you can run those versions in parallel during the transition period.

The biggest difference to keep in mind is that the label selector isn't in a `Grafana` spec anymore (`dashboardLabelSelector`), it's in other CRs instead (`instanceSelector`). Basically, it means that, say, a `GrafanaDashboard` "chooses" which `Grafana` instance to apply to, not the other way around. In terms of manifests, you get from:

```yaml
# v4
apiVersion: integreatly.org/v1alpha1
kind: Grafana
metadata:
  name: example-grafana
spec:
  # [...]
  dashboardLabelSelector:
    - matchExpressions:
        - { key: app, operator: In, values: [grafana] }
---
apiVersion: integreatly.org/v1alpha1
kind: GrafanaDashboard
metadata:
  name: simple-dashboard
  labels:
    app: grafana
spec:
  json: >
    {
      "id": null,
      "title": "Simple Dashboard",
      "tags": [],
      "style": "dark",
    }
```

to:

```yaml
# v5
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  name: grafana
  labels:
    dashboards: "grafana"
spec:
  # [...]
---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafanadashboard-sample
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  json: >
    {
      "id": null,
      "title": "Simple Dashboard",
      "tags": [],
      "style": "dark",
    }
```

As you can see, in a basic scenario of the dashboard with inline-JSON, a very few changes were needed. Although, other scenarios are likely to require a bit more effort.

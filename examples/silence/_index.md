---
title: "Silence"
weight: 60
tags:
  - Alerting
---

Shows how to create an alerting silence.

A silence stops notifications from one or more alerts for a fixed time window. The window
is defined with `startsAt` and `endsAt` (RFC3339, UTC), and the alerts to silence are
selected with `matchers`.  Note that the `endsAt` must be in the future and more recent than `startsAt`.

The Grafana-assigned silence ID is stored back on the resource in the
`grafana.integreatly.org/silence-id` annotation as a JSON map of
`<namespace>/<name>` instance to silence ID. To adopt (import) a silence that already
exists in Grafana, pre-populate that annotation with the instance and existing silence ID;
the operator will update that silence instead of creating a new one.

To view the entire configuration that you can do within Silences, look at our [API documentation](/docs/api/#grafanasilencespec).

To bind the silence directly to a single alert rule (creating a "rule-specific silence"), you must include the unique label matcher __alert_rule_uid__ set to your alert rule's UID
```
    "matchers": [
      {
        "name": "__alert_rule_uid__",
        "value": "d9x8a1b2c",
        "isRegex": false,
        "isEqual": true
      }
    ],
```
Note  to associate the alert to an alert rule the following conditions *must* be met:
* name must be exactly `__alert_rule_uid__`
* `value` must the exact alert guid
* `isRegex` must be set to `false`
* `isEqual` must be set to `true`

1. The finalizer runs and calls DeleteSilence(...) against each instance → ✅ this is firing (that's exactly why you see them flip to "expired" instead of staying active).
2. Grafana then keeps the expired silence in the list for a fixed 5-day retention, after which a background job (runs every ~15 min) permanently removes it.

For further info on silences see the [Grafana documentation](https://grafana.com/docs/grafana/latest/alerting/set-up/configure-rbac/silence-access/).

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

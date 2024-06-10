---
title: "Move plugin management to instance"
linkTitle: "Move plugin management to instance"
---

## Summary

The Grafana operator allows installation of plugins through the `GrafanaDashboard` and `GrafanaDatasource` resources.
This allows users to keep information about plugin requirements close to the place where they are used.

However, this also introduces security concerns as it allows everyone with the permission to create dashboards/data sources to install plugins as well.

This proposal outlines the steps to be taken to move plugin management to the `Grafana` resource which will allow for better permission management.
## Info

status: Suggested

## Motivation

By consolidating the plugin management into the `Grafana` CR, the group of people with permissions to install plugins is restricted to the same group which is also able to modify the instance.

## Proposal

Add a `plugins` field to the `Grafana` CRD as follows:

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  name: grafana
  labels:
    dashboards: "grafana"
spec:
  plugins:
    - name: grafana-clock-panel
      version: 1.3.0
```

## Impact on the already existing CRD

In the first step, this change is only additive, thus not impacting existing resources.
To adress the security concerns, this field should override any plugins specified by resources if set.
An empty array value then disables plugin installation through dashboards & data sources.

For [the v1 version of our CRDs](https://github.com/grafana/grafana-operator/milestone/4), the plugin fields can be removed from data sources & dashboards.

## Decision Outcome


## Related discussions

- [#1572](https://github.com/grafana/grafana-operator/issues/1572)

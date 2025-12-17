---
title: "Generic Grafana Resources"
---

## Summary

With the introduction of the [App Platform](https://github.com/grafana/grafana-app-sdk/blob/main/docs/application-design/platform-concepts.md) in Grafana, it is now possible to extend Grafana in a simliar way to kubernetes.
This allows for more in-depth plugins but supporting each resource individually would not be feasibile for the operator.
To still support these resources, we need a resource type that can interface with arbitrary grafana resources.

## Info

status: Accepted

## Proposal

We propose to create a new CRD called `GrafanaManifest`.
This resource contains all the common fields from the Grafana Operator (`resyncPeriod`,`instanceSelector`,`suspend`,`allowCrossNamespaceImort`) in addition to a `spec` field that contains the raw specification as required by the Grafana API:

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaManifest
metadata:
  name: example-playlist
spec:
  instanceSelector:
    matchLabels:
      env: dev
  template:
    apiVersion: playlist.grafana.app/v0alpha1
    kind: Playlist
    metadata:
      name: d7cd0c85-5c53-452c-bbef-84b5fe669ef2
      namespace: stacks-650111
    spec:
      interval: 5m
      items:
        - type: dashboard_by_uid
          value: 20201230-spring
      title: test
```


The above example applies the playlist with name `d7cd0c85-5c53-452c-bbef-84b5fe669ef2` to all instances matched by `env=dev`.

### Considerations

App platform resources contain a `metadata.namespace` field.
In most scenarios, this will be `default` but in multi tenant setups (like Grafana Cloud), this value is dynamic and different per instance.
To accomodate this, we need to extend the `Grafana` resource to support a field like `tenantNamespace` in `spec.external`.
During reconciliation, the `namespace` field of the `grafana.app` resource is set to the value of this new field if not set.
By making the `namespace` field optional, this still allows for a single `Grafana` resource being configured with `grafana.app` resources in multiple tenants.

### Future plans

Since core resources like Dashboards, Folders, Data sources etc. will also use this new API style, we can keep the reconciliation logic generic and auto-generate schemas for supported core resources from the API specification.
For now, stable schemas are still a work in progress.

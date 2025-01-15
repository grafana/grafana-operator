---
title: "Dynamic Notification Policy Routes"
linkTitle: "Dynamic Notification Policy Routes"
---

## Summary

Add a new CRD `GrafanaNotificationPolicyRoute` that will allow us to dynamically assemble a [Notification Policy](https://grafana.com/docs/grafana/latest/alerting/fundamentals/notifications/notification-policies/) from multiple individual `GrafanaNotificationPolicyRoute` resources in the cluster.

## Info

Status: Suggested

## Motivation

Today the `GrafanaNotificationPolicy` CRD allows configuration of a static Notification Policy in code.
This works great if you can define the full Notification Policy and all its routes upfront in a central place.

In todays microservice or multi-tenant environments it can be the case that new [Contact Points](https://grafana.com/docs/grafana/latest/alerting/fundamentals/notifications/contact-points/) are created dynamically for new teams or services, which would require adding new routes to these newly created Contact Points in the central `GrafanaNotificationPolicy`.

If both resources are defined in a central place, this can be achieved today, however if for example `Contact Points` are added dynamically with the deployment of a service, the central Notification Policy can only be updated by modifying the central version.

## Verification

- Create e2e tests for the operator assembling `GrafanaNotificationPolicy` from multiple `GrafanaNotificationPolicyRoute`s

## Proposal

Ideally there would be a new CRD `GrafanaNotificationPolicyRoute`, that allows specifying both `routes` and a `priority` in a separate resource.
- `routes` would be the same as the JSON routes in [GrafanaNotificationPolicy.spec.route](https://grafana.github.io/grafana-operator/docs/api/#grafananotificationpolicyspecroute).
- `policy` would be an integer that would allow modifying the merge-order of individual `NotificationPolicyRoutes`

The existing `GrafanaNotificationPolicy` could be extended with a `routeSelector`, which could discover `GrafanaNotificationPolicyRoute` objects via a label matcher.

### Example

New `GrafanaNotificationPolicy` that includes the new `routeSelector` using the `matchLabels` label selector:

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaNotificationPolicy
metadata:
  name: grafananotificationpolicy-sample
spec:
  allowCrossNamespaceImport: true
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  routeSelector:
    matchLabels:
      dynamicroute: "grafana"
  route:
    receiver: grafana-default-email
    group_by:
      - grafana_folder
      - alertname
    routes:
      - receiver: grafana-default-email
        object_matchers:
          - - team
            - =
            - a
          - - inline
            - =
            - first
      - receiver: grafana-default-email
        object_matchers:
          - - team
            - =
            - b
          - - inline
            - =
            - second
```

> In this example `routeSelector` and `routes` are used combined, we could also make them mutually exclusive, to simplify this even more.

Three example `NotificationPolicyRoute` resources:

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaNotificationPolicyRoute
metadata:
  name: dynamic-c
  namespace: grafana-crds
  labels:
    dynamicroute: "grafana"
spec:
  priority: 1
  route:
    receiver: grafana-default-email
    object_matchers:
      - - crossNamespace
        - =
        - "true"
      - - dynamic
        - =
        - c
      - - priority
        - =
        - "1"
```

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaNotificationPolicyRoute
metadata:
  name: dynamic-d
  labels:
    dynamicroute: "grafana"
spec:
  priority: 2
  route:
    receiver: grafana-default-email
    object_matchers:
      - - dynamic
        - =
        - d
      - - priority
        - =
        - "2"
```

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaNotificationPolicyRoute
metadata:
  name: dynamic-e
  labels:
    dynamicroute: "grafana"
spec:
  route:
    receiver: grafana-default-email
    object_matchers:
      - - dynamic
        - =
        - e
      - - priority
        - =
        - none
```

Which would result in the following merged routes:

```yaml
      routes:
        - receiver: grafana-default-email
          object_matchers:
            - - inline
              - =
              - first
            - - team
              - =
              - a
        - receiver: grafana-default-email
          object_matchers:
            - - inline
              - =
              - second
            - - team
              - =
              - b
        - receiver: grafana-default-email
          object_matchers:
            - - crossNamespace
              - =
              - "true"
            - - dynamic
              - =
              - c
            - - priority
              - =
              - "1"
        - receiver: grafana-default-email
          object_matchers:
            - - dynamic
              - =
              - d
            - - priority
              - =
              - "2"
        - receiver: grafana-default-email
          object_matchers:
            - - dynamic
              - =
              - e
            - - priority
              - =
              - none
```

> This example assumes that routes with a higher priority value are merged before lower priority values.


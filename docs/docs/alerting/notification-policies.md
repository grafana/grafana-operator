---
title: Notification Policies
---

Notification policies provide you with a flexible way of designing how to handle notifications and minimize alert noise.
For a complete explanation on notification policies, see the [upstream Grafana documentation](https://grafana.com/docs/grafana/latest/alerting/fundamentals/notifications/notification-policies/).

{{% alert title="Tip" color="secondary" %}}
If you already know which contact point an alert should send to, you can directly set the [`receivers`]({{% relref "/docs/api/#grafanaalertrulegroupspecrulesindexnotificationsettings" %}}) property on the alert rule.
{{% /alert %}}

## Simple Notification Policy

The following snippet shows an example notification policy routing to the `operations` or `security` team based on the `team` label.

{{< readfile file="../examples/notification-policy/resources.yaml" code="true" lang="yaml" >}}

## Dynamic Notification Policy Routes

There might be scenarios where you can not define the entire notification policy in a single place and you have to assemble it from multiple resouces.
In this case, you can use the `spec.routeSelector` in combination with multiple `GrafanaNotificationPolicyRoute` resources.

All `GrafanaNotificationPolicyRoute` resources will then be discovered based on the label selector defined in `spec.routeSelector`.
In case `spec.allowCrossNamespaceImport` is enabled, matching routes will be fetched from all namespaces.
Otherwise only routes from the same namespace as the `GrafanaNotificationPolicy` will be discovered.

All discovered routes will then get appended to the `spec.route.routes[]` of the `GrafanaNotificationPolicy` based on the priority defined in the `GrafanaNotificationPolicyRoute`.
Priorities can be in the range 1-100 with `1` being merged first and `100` last. If no priority is specified, it is treated as a priority of `100`.

The following shows an example of how dynamic routes will get merged.

{{< readfile file="../examples/notification-policy/routes.yaml" code="true" lang="yaml" >}}

The resulting Notification Policy will be the following:

```yaml
apiVersion: 1
policies:
    - orgId: 1
      receiver: grafana-default-email
      group_by:
        - grafana_folder
        - alertname
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
```

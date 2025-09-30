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

{{< readfile file="notification-policy/resources.yaml" code="true" lang="yaml" >}}

## Dynamic Notification Policy Routes

There might be scenarios where you can not define the entire notification policy in a single place and you have to assemble it from multiple resources.
In this case, you can use the `routeSelector` field in combination with multiple `GrafanaNotificationPolicyRoute` resources.

Both `GrafanaNotificationPolicy` and `GrafanaNotificationPolicyRoute` objects support the `routeSelector` field.

All `GrafanaNotificationPolicyRoute` resources will then be discovered based on the label selector defined in `spec.route.routeSelector`.
In case `spec.allowCrossNamespaceImport` is enabled, matching routes will be fetched from all namespaces.
Otherwise only routes from the same namespace as the `GrafanaNotificationPolicy` will be discovered.

The discovered routes are then used when applying the notification policy.

{{% alert title="Note" color="secondary" %}}
The `spec.route.routes` and `spec.route.routeSelector` fields are mutually exclusive.
When both fields are specified, the `routeSelector` takes precedence and overrides anything defined in `routes`.
{{% /alert %}}

The following shows an example of how dynamic routes will get merged.

{{< readfile file="notification-policy/routes.yaml" code="true" lang="yaml" >}}

The resulting Notification Policy will be the following:

![Dynamic notification policy tree after applying the example routes](../dynamic-notification-policy.png)

---
title: Notification Policies
---

Notification policies provide you with a flexible way of designing how to handle notifications and minimize alert noise.
For a complete explanation on notification policies, see the [upstream Grafana documentation](https://grafana.com/docs/grafana/latest/alerting/fundamentals/notifications/notification-policies/).

{{% alert title="Tip" color="secondary" %}}
If you already know which contact point an alert should send to, you can directly set the [`receivers`]({{% relref "/docs/api/#grafanaalertrulegroupspecrulesindexnotificationsettings" %}}) property on the alert rule.
{{% /alert %}}


The following snippet shows an example notification policy routing to the `operations` or `security` team based on the `team` label.

{{< readfile file="../examples/notification-policy/resources.yaml" code="true" lang="yaml" >}}

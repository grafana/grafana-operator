---
title: Contact Points
---

Contact points contain the configuration for sending alert notifications. You can assign a contact point either in the alert rule or notification policy options.
For a complete explanation on notification policies, refer to the [upstream Grafana documentation](https://grafana.com/docs/grafana/latest/alerting/fundamentals/notifications/contact-points/).

{{% alert title="Note" color="secondary" %}}
The Grafana operator currently only supports a single receiver per contact point definition.
As a workaround you can create multiple contact points with the same `spec.name` value.
Follow issue [#1529](https://github.com/grafana/grafana-operator/issues/1529) for further updates on this topic.
{{% /alert %}}

The following snippet shows an example contact point which notifies a specific email address.
It also highlights how secrets and config maps can utilized to externalize some of the configuration.
This is especially useful for contact points which contain sensitive information.

{{< readfile file="../examples/contactpoint_override/resources.yaml" code="true" lang="yaml" >}}

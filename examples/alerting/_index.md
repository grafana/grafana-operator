---
title: Alerting
weight: 13
---
{{% pageinfo color="primary" %}}
Alerting resources require Grafana version 9.5 or higher.
{{% /pageinfo %}}

The Grafana Operator currently only supports _Grafana Managed Alerts_.

For data source managed alerts, refer to the documentation and tooling available for the respective data source.
{{% alert title="Note" color="primary" %}}
When using Mimir/Prometheus, you can use the [`mimir.rules.kubernetes`](https://grafana.com/docs/alloy/latest/reference/components/mimir/mimir.rules.kubernetes/) component of [Grafana Alloy](https://grafana.com/docs/alloy/latest/) to deploy rules as Kubernetes resources.
{{% /alert %}}


## Full example

The following resources construct the flow outlined in the [Grafana notification documentation](https://grafana.com/docs/grafana/latest/alerting/fundamentals/notifications/).

They create:
1. Three alert rules across two different groups
2. Two contact points for two different teams
3. A notification policy to route alerts to the correct team

{{< figure src="notification-routing.png" title="Flowchart of alerts routed through this system" width="500" >}}

{{% alert title="Note" color="primary" %}}
If you want to try this for yourself, you can [get started with demo data in Grafana cloud](https://grafana.com/docs/grafana-cloud/get-started/#install-demo-data-sources-and-dashboards).
The examples below utilize the data sources to give you real data to alert on.
{{% /alert %}}

### Alert rule groups

The first resources in this flow are _Alert Rule Groups_.
An alert rule group can contain multiple alert rules.
They group together alerts to run on the same interval and are stored in a Grafana folder, alongside dashboards.

First, create the folder:

{{< readfile file="notification_policy/notifications-full/folder.yaml" code="true" lang="yaml" >}}

The first alert rule group is responsible for alerting on well known Kubernetes issues:

{{< readfile file="notification_policy/notifications-full/kubernetes-alert-rules.yaml" code="true" lang="yaml" >}}

The second alert rule group is responsible for alerting on security issues:

{{< readfile file="notification_policy/notifications-full/security-alert-rules.yaml" code="true" lang="yaml" >}}

After applying the resources, you can see the created rule groups in the _Alert rules_ overview page:

![Alert rules overview page](./overview-page.png)

### Contact Points

Before you can route alerts to the correct receivers, you need to define how these alerts should be delivered.
[Contact points](./contact-point/_index.md) specify the methods used to notify someone using different providers.

Since the two different teams get notified using different email addresses, two contact points are required.

{{< readfile file="notification_policy/notifications-full/contact-points.yaml" code="true" lang="yaml" >}}

### Notification Policy

Now that all parts are in place, the only missing component is the notification policy.
The instances notification policy routes alerts to contact points based on labels.
A Grafana instance can only have one notification policy applied at a time as it's a global object.

The following notification policy routes alerts based on the team label and further configures the repetition interval for high severity alerts belonging to the operations team:

{{< readfile file="notification_policy/notifications-full/notification-policy.yaml" code="true" lang="yaml" >}}

After applying the resource, Grafana shows the following notification policy tree:

![Notification policy tree after applying the resource](./notification-policy-tree.png)

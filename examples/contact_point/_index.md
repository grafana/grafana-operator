---
title: Contact Points
weight: 60
tags:
  - Alerting
  - ValuesFrom
---

Contact points contain the configuration for sending alert notifications. You can assign a contact point either in the alert rule or notification policy options.
For a complete explanation on notification policies, refer to the [upstream Grafana documentation](https://grafana.com/docs/grafana/latest/alerting/fundamentals/notifications/contact-points/).

The following snippet shows an example contact point which notifies a specific email address and a webhook endpoint.
It also highlights how secrets and config maps can utilized to externalize some of the configuration.
This is especially useful for contact points which contain sensitive information.

To view the entire configuration that you can do within Contact-Points, look at our [API documentation](/docs/api/#grafanacontactpointspec).

{{< readfile file="./resources.yaml" code="true" lang="yaml" >}}

### Deprecated Single receiver format

`GrafanaContactPoint` did not support multiple receivers prior to `v5.21.0`.

The previous format is now deprecated, but is still supported.

The fields `.spec.type`, `.spec.settings`, and `.spec.valuesFrom` are entirely ignored when `.spec.recievers[...]` is configured, but the below configuration is still valid for the duration of `v1beta1`.

This means full backwards compatibility during transitions to the receivers field.

{{< readfile file="./previous-format.yaml" code="true" lang="yaml" >}}

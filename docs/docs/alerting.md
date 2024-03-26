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
When using Mimir/Prometheus, you can use the [`mimir.rules.kubernetes`](https://grafana.com/docs/agent/latest/flow/reference/components/mimir.rules.kubernetes/) component of the Grafana Agent to deploy rules as Kubernetes resources.
{{% /alert %}}


## Alert rule groups

Alert Rule Groups contain a list of alerts which should evaluate at the same interval.
Every rule group must belong to a folder and contain at least one rule.

The easiest way to get the YAML specification for an alert rule is to use the [modify export feature](https://grafana.com/docs/grafana/latest/alerting/set-up/provision-alerting-resources/export-alerting-resources/), introduced in Grafana 10.

The following snippet shows an example alert rule group with a single alert that fires when the temperature is below zero degrees.

{{< readfile file="examples/alertrulegroups/resources.yaml" code="true" lang="yaml" >}}

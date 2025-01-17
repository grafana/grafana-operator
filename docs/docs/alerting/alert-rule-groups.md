---
title: Alert Rule Groups
---

Alert Rule Groups contain a list of alerts which should evaluate at the same interval.
Every rule group must belong to a folder and contain at least one rule.

The easiest way to get the YAML specification for an alert rule is to use the [modify export feature](https://grafana.com/docs/grafana/latest/alerting/set-up/provision-alerting-resources/export-alerting-resources/), introduced in Grafana 10.

The following snippet shows an example alert rule group with a single alert that fires when the temperature is below zero degrees.

{{< readfile file="../examples/alertrulegroups/resources.yaml" code="true" lang="yaml" >}}

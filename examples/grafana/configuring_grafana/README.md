---
title: "Configure Grafana"
linkTitle: "Configure Grafana"
weight: 10
---

## Grafana config

We offer the `grafana.config` field where you can pass any Grafana configuration values you want.

The operator does not make any extra validation of your configuration, so just like a non-operator deployment of Grafana, your Grafana instance might be broken due to a configuration error.

To find all possible configuration options, look at the [official documentation](https://grafana.com/docs/grafana/latest/setup-grafana/configure-grafana/).

In case you need to specify top level options like `app_mode` or `instance_name`, specify them in the `global` section like this:

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  name: grafana
  labels:
    dashboards: "grafana"
spec:
  config:
    global:
      app_mode: "development"
```

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

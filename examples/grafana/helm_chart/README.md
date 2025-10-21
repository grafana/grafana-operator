---
title: "Helm Grafana deployments"
weight: 30
---

Using the official [Grafana Helm Chart](https://github.com/grafana/helm-charts/pkgs/container/helm-charts%2Fgrafana), it's easy to extend the values to make the instance visible to the operator:

This takes multiple values into account: `release name and namespace, labels, and secret configs`, but may not be what you want.

It is recommended to use this as a guide in case you have complex requirements.

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

{{% alert title="Note" color="primary" %}}
If you're using the official chart via `kube-prometheus-stack`.
Indent the entire example once under the `.grafana` key
```yaml
grafana:
  extraObjects:
    - ...
```
{{% /alert %}}

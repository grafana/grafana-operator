---
title: "Introduction"
linkTitle: "Documentation"
weight: 20
menu:
  main:
    weight: 20
---

The Grafana operator allows you to:
* ⚙️ Deploy & Manage Grafana Instances inside of Kubernetes with ease
* 🌐 Manage externally hosted instances using Kubernetes resources (for example Grafana Cloud)

To install the Grafana Operator in your Kubernetes cluster, Run the following command in your terminal:

```bash
helm upgrade -i grafana-operator oci://ghcr.io/grafana/helm-charts/grafana-operator --version {{<param version>}}
```

For a detailed installation guide, refer to [the installation documentation]({{<relref installation>}}).

To get started, take a look at the [quick start guide]({{<relref quick-start.md>}}).

<div align="center">

# Grafana Operator

[![Grafana](https://img.shields.io/badge/grafana-%23F46800.svg?&logo=grafana&logoColor=white)](https://grafana.com/)
[![Grafana Operator](https://img.shields.io/badge/Grafana%20Operator-orange)](https://grafana.github.io/grafana-operator/)
[![GitHub tag](https://img.shields.io/github/tag/grafana/grafana-operator.svg)](https://github.com/grafana/grafana-operator/tags)
[![GitHub Last Commit](https://img.shields.io/github/last-commit/grafana/grafana-operator)](https://github.com/grafana/grafana-operator/tags)
[![GitHub Contributors](https://img.shields.io/github/contributors/grafana/grafana-operator)](https://github.com/grafana/grafana-operator/tags)

**[Official Documentation](https://grafana-operator.github.io/grafana-operator/)** | **[Quickstart](#getting-started)** | **[Installation](#installation)** | **[Tutorials](https://grafana.com/docs/grafana-cloud/developer-resources/infrastructure-as-code/grafana-operator/)**

</div>

The Grafana Operator is a Kubernetes operator built to help you manage your Grafana instances and its resources in and outside of
Kubernetes.

Whether you’re running one Grafana instance or many, the Grafana Operator simplifies the processes of installing, configuring, and maintaining Grafana and its resources. Additionally, it's perfect for those who prefer to manage resources using infrastructure as code or using GitOps workflows through tools like ArgoCD and Flux CD.

## Getting Started

### Installation

**Option 1: Helm Chart**

Deploy the Grafana Operator easily in your cluster using Helm:

```bash
helm upgrade -i grafana-operator oci://ghcr.io/grafana/helm-charts/grafana-operator --version 5.21.3
```

**Option 2: Kustomize & More**

Prefer Kustomize, Openshift OLM, or Kubernetes directly? Find detailed instructions in our [Installation Guide](https://grafana.github.io/grafana-operator/docs/installation/kustomize/).

For even more detailed setups, see our [documentation](docs/README.md).

### Example: Deploying Grafana & A Dashboard

Here's a simple example of deploying Grafana and a Grafana Dashboard using the custom resources (CRs) defined by the Grafana Operator:

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  name: grafana
  labels:
    dashboards: "grafana"
spec:
  config:
    log:
      mode: "console"
    security:
      admin_user: root
      admin_password: secret

---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: sample-dashboard
spec:
  resyncPeriod: 30s
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  json: >
    {
      "title": "Simple Dashboard",
      "timezone": "browser",
      "refresh": "5s",
      "panels": [],
      "time": {
        "from": "now-6h",
        "to": "now"
      }
    }
```

For more tailored setups and resources management, check out these guides:

- [Managing Data Sources and Dashboards](https://grafana.com/docs/grafana-cloud/developer-resources/infrastructure-as-code/grafana-operator/operator-dashboards-folders-datasources/)
- [GitOps Dashboards Management with ArgoCD](https://grafana.com/docs/grafana-cloud/developer-resources/infrastructure-as-code/grafana-operator/manage-dashboards-argocd/)

## Why Grafana Operator?

Switching to Grafana Operator from traditional deployments amplifies your efficiency by:

- Enabling multi-instance and multi-namespace Grafana deployments effortlessly.
- Simplifying dashboard, data sources, and plugin management through code.
- Supporting both Kubernetes and Openshift with smart adjustments based on the environment.
- Allowing management of external Grafana instances for robust GitOps integration.
- Providing multi-architecture support, making it versatile across different platforms.
- Offering one-click installation through Operatorhub/OLM.

## Get In Touch

Got questions or suggestions? Let us know! The quickest way to reach us is through our [GitHub Issues](https://github.com/grafana/grafana-operator/issues) or by joining our weekly public meeting on Mondays at 13:30 Central European (Summer) Time (11:30/12:30 UTC in Summer/Winter) (link [here](https://meet.google.com/sqk-kdsc-ntv)).

Feel free to drop into our Grafana Operator discussions on:

[![Grafana Slack](https://img.shields.io/badge/grafana%20community%20Slack-4A254A?logo=slack&logoColor=white)](https://join.slack.com/t/grafana/shared_invite/zt-2eqidcplt-QzkxMuhZA4tGQeFQenE_MQ)

## Contributing

For more information on how to contribute to the operator look at [CONTRIBUTING.md](CONTRIBUTING.md).

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=grafana/grafana-operator&type=date&legend=top-left)](https://www.star-history.com/#grafana/grafana-operator&type=date&legend=top-left)

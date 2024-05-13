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
helm upgrade -i grafana-operator oci://ghcr.io/grafana/helm-charts/grafana-operator --version v5.6.3
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
    auth:
      disable_login_form: false
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

## Get In Touch!

Got questions or suggestions? Let us know! The quickest way to reach us is through our [GitHub Issues](https://github.com/grafana/grafana-operator/issues) or by joining our weekly public meeting on Mondays at 11:00 Central European (Summer) Time (09:00/10:00 UTC in Summer/Winter) (link [here](https://meet.google.com/spw-jtbk-mwj)).

Feel free to drop into our Grafana Operator discussions on:

[![Kubernetes Slack](https://img.shields.io/badge/kubernetes%20slack-white?logo=slack&logoColor=black)](https://kubernetes.slack.com/archives/C019A1KTYKC) [![Grafana Slack](https://img.shields.io/badge/grafana%20community%20Slack-4A254A?logo=slack&logoColor=white)](https://join.slack.com/t/grafana/shared_invite/zt-2eqidcplt-QzkxMuhZA4tGQeFQenE_MQ)



## Contributing

For more information on how to contribute to the operator look at [CONTRIBUTING.md](CONTRIBUTING.md).

## Version Support and Development Mindset

> [!CAUTION]
> v4 will stop receiving bug fixes and security updates as of the 22nd of December 2023.
> We recommend you migrate to v5 if you haven't yet! Please follow our [v4 -> v5 Migration Guide](https://grafana.github.io/grafana-operator/blog/2023/05/27/v4-to-v5-migration/) to mitigate any potential future risks.


V5 is the current, actively developed and maintained version of the operator, which you can find on the
***[Master Branch](https://github.com/grafana/grafana-operator/tree/master)***.

A more in-depth overview of v5 is available in the [intro blog](docs/blog/v5-intro.md)

V5 is a ground-up rewrite of the operator to refocus development on:

- Performance
- Reliability
- Maintainability
- Extensibility
- Testability
- Usability

The previous versions of the operator have some serious tech-debt issues, which effectively prevent community members
that aren't massively
familiar with the project and/or its codebase from contributing features that they wish to see.

These previous versions, we're built on a "as-needed" basis, meaning that whatever was the fastest way to reach the
desired feature, was the way
it was implemented. This lead to situations where controllers for different resources were using massively different
logic, and features were added
wherever and however they could be made to work.

V5 aims to re-focus the operator with a more thought out architecture and framework, that will work better,
both for developers and users.
With certain standards and approaches, we can provide a better user experience through:

- Better designed Custom Resource Definitions (Upstream Grafana Native fields will be supported without having to
  whitelist them in the operator logic).
    - Upstream documentation can be followed to define the Grafana Operator Custom Resources.
    - This also means a change in API versions for the resources, but we see this as a benefit, our previous mantra of
      maintaining a seamless upgrade from version to version, limited us in the changes we wanted to make for a long
      time.
- A more streamlined Grafana resource management workflow, one that will be reflected across all controllers.
- Using an upstream Grafana API client (standardizing our interactions with the Grafana API, moving away from bespoke
  logic).
- The use of a more up-to-date Operator-SDK version, making use of newer features.
    - along with all relevant dependencies being kept up-to-date.
- Proper testing.
- Cleaning and cutting down on code.
- Multi-instance and Multi-namespace support!

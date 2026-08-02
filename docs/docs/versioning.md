---
title: Versioning
weight: 60
---

## Supported versions

For Grafana, the operator intends to support only the last 3 major versions.

For Kubernetes, we follow its [release schedule](https://kubernetes.io/releases/). Once a particular version reaches its EOL, we remove it from our test matrix.

> Depending on a controller, the operator is likely to work fine with versions outside of those boundaries, but we do not guarantee that.

## Default Grafana version

The Grafana version when unspecified in `Grafana#spec.version`

> Only versions that have changes around Grafana image tags are mentioned below.

| Operator Version | Default Grafana Image |
|-|-|
| `v5.23.0` | `13.0.1` |
| `v5.22.1` | `12.4.1` |
| `v5.22.0` | `12.3.3` |
| `v5.21.0` | `12.3.0` |
| `v5.19.0` | `12.1.0` |
| `v5.16.0` | `11.3.0` |
| `v5.9.2` | `10.4.3` |
| `v5.7.0` | `9.5.17` |
| `v5.0.0` | `9.1.6` |

## Published Artifacts and Changelogs

Changelogs are published under [Github Releases](https://github.com/grafana/grafana-operator/releases/).

Previous artifacts are available under [Github Packages](https://github.com/orgs/grafana/packages?repo_name=grafana-operator).

|     Type          |                        Source                        |
| ----------------- | ---------------------------------------------------- |
| Docker Image      | `ghcr.io/grafana/grafana-operator`                   |
| Helm Chart        | `oci://ghcr.io/grafana/helm-charts/grafana-operator` |
| Flux OCI Artifact | `ghcr.io/grafana/kustomize/grafana-operator`         |

---
title: "Basic HTTPRoute Example"
linkTitle: "Basic HTTPRoute"
---

This example demonstrates basic HTTPRoute configuration for Grafana using Gateway API.

## Prerequisites

- Kubernetes cluster with Gateway API CRDs installed
- A Gateway resource named `my-gateway` in the `default` namespace
- grafana-operator installed

## Install Gateway API CRDs

```bash
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.3.0/standard-install.yaml
```

## Deploy

```bash
kubectl apply -f resources.yaml
```

## Example Manifest

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

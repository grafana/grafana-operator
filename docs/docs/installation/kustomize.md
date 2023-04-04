---
title: "Kustomize installation"
linkTitle: "Kustomize installation"
description: "How to install grafana-operator using Kustomize"
---

We are using Flux to package our Kustomize files through OCI, and they are built and released just as our helm solution.

To our knowledge there are no way of downloading manifest files through the [Kustomize CLI](https://kustomize.io/).

So if you want to download the Kustomize manifest you need to install the [Flux cli](https://fluxcd.io/flux/installation/).

## Download Kustomize files

After you have downloaded Flux you can use `flux pull artifact` to download the manifests.

```shell
mkdir grafana-operator
flux pull artifact oci://ghcr.io/grafana-operator/kustomize/grafana-operator:v5.0.0-rc1 --output ./grafana-operator/
```

This will provide you the manifest files unpacked and ready to use.

## Install

Two overlays are provided, for namespace scoped and cluster scoped installation.
For more information look at our [documentation](https://grafana-operator.github.io/grafana-operator/docs/grafana/#where-should-the-operator-look-for-grafana-resources).

This will install the operator in the grafana namespace.

```shell
kubectl apply -k deploy/overlays/cluster_scoped
```

for a namespace scoped installation

```shell
kubectl apply -k deploy/overlays/namespace_scoped
```

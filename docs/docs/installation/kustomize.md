---
title: "Kustomize installation"
linkTitle: "Kustomize installation"
description: "How to install grafana-operator using Kustomize"
---

We are using Flux to package our Kustomize files through OCI, and they are built and released just as our helm solution.

There is no way of downloading manifest files through the [Kustomize CLI](https://kustomize.io/), but hopefully something that will be supported in the [future](https://github.com/kubernetes-sigs/kustomize/issues/5134).

So if you want to download the Kustomize manifest you need to install the [Flux cli](https://fluxcd.io/flux/installation/).

## Download Kustomize files

After you have downloaded Flux you can use `flux pull artifact` to download the manifests.

```shell
mkdir grafana-operator
flux pull artifact oci://ghcr.io/grafana/kustomize/grafana-operator:{{<param version>}} --output ./grafana-operator/
```

This will provide you the manifest files unpacked and ready to use.

## Install

Two overlays are provided, for namespace scoped and cluster scoped installation.
For more information look at our [documentation](https://grafana-operator.github.io/grafana-operator/docs/grafana/#where-should-the-operator-look-for-grafana-resources).

This will install the operator in the grafana namespace.

```shell
kubectl create -k deploy/overlays/cluster_scoped
```

for a namespace scoped installation

```shell
kubectl create -k deploy/overlays/namespace_scoped
```

### Patching grafana-operator

When you want to path the grafana operator instead of using `kubectl apply` you need to use `kubectl replace`.
Else you will get the following error `invalid: metadata.annotations: Too long: must have at most 262144 bytes`.

For example

```shell
kubectl replace -k deploy/overlays/cluster_scoped
```

For more information how `kubectl replace` works we recommend reading this [blog](https://blog.atomist.com/kubernetes-apply-replace-patch/).

---
title: "Kustomize installation"
linkTitle: "Kustomize installation"
description: "How to install grafana-operator using Kustomize"
---

## Flux

We are using Flux to package our Kustomize files through OCI, and they are built and released just as our helm solution.

So if you want to download the Kustomize manifest you need to install the [Flux cli](https://fluxcd.io/flux/installation/).

## Download Kustomize files

After you have downloaded Flux you can use `flux pull artifact` to download the manifests.

```shell
mkdir grafana-operator
flux pull artifact oci://ghcr.io/grafana/kustomize/grafana-operator:{{<param version>}} --output ./grafana-operator/
```

This will provide you the manifest files unpacked and ready to use.

## Kustomize / Kubectl

You can also find the yaml for the `cluster_scoped` and `namespace_scoped` release on the [release page](https://github.com/grafana/grafana-operator/releases/latest)

## Install

Two overlays are provided, for namespace scoped and cluster scoped installation.
For more information look at our [documentation](https://grafana-operator.github.io/grafana-operator/docs/examples/grafana/#where-should-the-operator-look-for-grafana-resources).

This will install the operator in the grafana namespace.

```shell
kubectl create -f https://github.com/grafana/grafana-operator/releases/latest/download/kustomize-cluster_scoped.yaml
```

For a namespace scoped installation:

```shell
kubectl create -f https://github.com/grafana/grafana-operator/releases/latest/download/kustomize-namespace_scoped.yaml
```

Note `kubectl apply -f ...` instead of `kubectl create -f ...` may produce the following error: `invalid: metadata.annotations: Too long: must have at most 262144 bytes`

### Patching grafana-operator

When you want to patch the grafana operator instead of using `kubectl apply` you need to use `kubectl replace`.
Else, you will get the following error: `invalid: metadata.annotations: Too long: must have at most 262144 bytes`

For example

```shell
kubectl replace -f https://github.com/grafana/grafana-operator/releases/latest/download/kustomize-namespace_scoped.yaml
```

### Kustomize

latest:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

# this will automatically pull the latest release when `kustomize build` is executed
resources:
  - https://github.com/grafana/grafana-operator/releases/latest/download/kustomize-cluster_scoped.yaml
```

pinned to release:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

# update the version to the release you need
resources:
  - https://github.com/grafana/grafana-operator/releases/download/{{<param version>}}/kustomize-cluster_scoped.yaml

```

## Configuration

Kustomize allows for customization through overlays. For example: if you want to
change log format to JSON, you can apply an override to the container and provide the
required arguments:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - https://github.com/grafana/grafana-operator/releases/download/{{<param version>}}/kustomize-cluster_scoped.yaml

patches:
  - target:
      group: apps
      version: v1
      kind: Deployment
      name: grafana-operator-controller-manager
    patch: |-
      - op: add
        path: /spec/template/spec/containers/0/args/-
        value: --zap-encoder=json
```

## Common Issues

### ArgoCD

If you are using ArgoCD you need to add this patch to fix the errors during apply of the CRD.

```yaml
patches:
  - patch: |-
      apiVersion: apiextensions.k8s.io/v1
      kind: CustomResourceDefinition
      metadata:
        annotations:
          argocd.argoproj.io/sync-options: Replace=true
        name: grafanas.grafana.integreatly.org
```

or

```yaml
patches:
  - patch: |-
      apiVersion: apiextensions.k8s.io/v1
      kind: CustomResourceDefinition
      metadata:
        annotations:
          argocd.argoproj.io/sync-options: ServerSideApply=true
        name: grafanas.grafana.integreatly.org
```

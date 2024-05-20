# Deploy with kustomize

Two overlays are provided, for namespace scoped and cluster scoped installation.
To install the Grafana operator, select one of the overlays and edit its `kustomization.yaml` file.
Make sure `namespace` is set to the namespace where you want to install the operator.
Then run:

```shell
kustomize build deploy/kustomize/overlays/cluster_scoped --load-restrictor LoadRestrictionsNone | kubectl apply -f -
```

for a cluster scoped installation, or:

```shell
kustomize build deploy/kustomize/overlays/namespace_scoped --load-restrictor LoadRestrictionsNone | kubectl apply -f -
```

for a namespace scoped installation.

When you want to patch the grafana operator instead of using `kubectl apply` you need to use `kubectl replace`.
Else you will get the following error `invalid: metadata.annotations: Too long: must have at most 262144 bytes`.

For example

```shell
kustomize build deploy/kustomize/overlays/namespace_scoped --load-restrictor LoadRestrictionsNone | kubectl replace -f -
```

# Deploy with kustomize

Two overlays are provided, for namespace scoped and cluster scoped installation.
To install the Grafana operator, select one of the overlays and edit its `kustomization.yaml` file.
Make sure `namespace` is set to the namespace where you want to install the operator.
Then run:

```shell
$ kubectl apply -k deploy/overlays/cluster_scoped
```

for a cluster scoped installation, or:

```shell
$ kubectl apply -k deploy/overlays/namespace_scoped
```

for a namespace scoped installation.

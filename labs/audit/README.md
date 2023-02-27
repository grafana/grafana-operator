# Audit stats lab

The goal of the lab is to get stats for interactions with kube-apiserver.

## Technical requirements

- bash
- docker
- jq
- kind

## Instructions

Create kind cluster:

```shell
kind create cluster --config kind-config.yaml
```

Deploy the operator and related resources from the root of the git folder:

```shell
export IMG=<XXX>
make install
make deploy
kubectl apply -f config/samples/grafana_v1beta1_grafana.yaml
kubectl apply -f config/samples/grafana_v1beta1_grafanadashboard.yaml
kubectl apply -f config/samples/grafana_v1beta1_grafanadatasource.yaml
```

NOTE: at the time of writing, kubebuilder lacks full definition for cluster-scope RBAC, you need to take care of that.

Collect stats after a few minutes:

```shell
./collect-audit-stats.sh
```

Example:

```shell
./collect-audit-stats.sh
   3 "get - /api/v1/namespaces/grafana-operator-system/configmaps/f75f3bba.integreatly.org"
   6 "get - /apis/coordination.k8s.io/v1/namespaces/grafana-operator-system/leases/f75f3bba.integreatly.org"
   3 "update - /api/v1/namespaces/grafana-operator-system/configmaps/f75f3bba.integreatly.org"
   3 "update - /apis/coordination.k8s.io/v1/namespaces/grafana-operator-system/leases/f75f3bba.integreatly.org"
```

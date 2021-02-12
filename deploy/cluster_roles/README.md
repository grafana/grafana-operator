# Grafana Operator | Cluster Roles

## Grant Grafana instance RBAC to GrafanaDashboard definitions in other projects/namespaces

By default a Grafana instance deployed by the Grafana Operator will only read GrafanaDashboard custom resources from the namespace/project that the Grafana instance is deployed in. If specifing the `--scan-all` or `--namespaces` flags or if your using `dashboardNamespaceSelector`, then the ServiceAccount that Grafana is running as needs view access to the GrafanaDashboard resources in other namespaces. To grant those permissions the following ClusterRole and ClusterRoleBinding need to be deployed.

Create the `ClusterRole`
```
kubectl create -f cluster_role_grafana_operator.yaml
```

Create the `ClusterRoleBinding` for the `ServiceAccount/grafana-operator` in the given namespace
```
GRAFANA_NAMESPACE=grafana
sed "s/namespace: grafana/namespace: ${GRAFANA_NAMESPACE}/g" cluster_role_binding_grafana_operator.yaml
```

## Grant non Cluster Admins permissions to deploy operator and associated Grafana instances

For a cluster administrator to allow other users to be able to deploy Grafana operators and the associated Custom Resources namespace/project admins/editors need edit access to the Grafana Custom Resources.

```
kubectl create -f cluster_role_aggregate_grafana_admin_edit.yaml
kubectl create -f cluster_role_aggregate_grafana_view.yaml
```

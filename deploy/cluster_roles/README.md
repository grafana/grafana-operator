# Grafana Operator | Cluster Roles

## Grant Grafana instance RBAC to GrafanaDashboard definitions in other projects/namespaces

By default, a Grafana instance deployed by the Grafana Operator will only read GrafanaDashboard custom resources
from the namespace/project that the Grafana instance is deployed in.

If specifying the `--scan-all`, `--namespaces`, `DASHBOARD_NAMESPACES_ALL="true"`  flags or if you're using
`dashboardNamespaceSelector`,
then the ServiceAccount that Grafana is running as needs view access to the GrafanaDashboard resources in other namespaces.
To grant those permissions the following ClusterRole and ClusterRoleBinding need to be deployed.

Create the `ClusterRole`

```shell
kubectl create -f deploy/cluster_roles/cluster_role_grafana_operator.yaml
```

Create the `ClusterRoleBinding` for the `ServiceAccount/grafana-operator` in the given namespace

```shell
GRAFANA_NAMESPACE=grafana
sed "s/namespace: grafana/namespace: ${GRAFANA_NAMESPACE}/g" cluster_role_binding_grafana_operator.yaml
```

## Grant non Cluster Admins permissions to deploy operator and associated Grafana instances

For a cluster administrator to allow other users to be able to deploy Grafana operators and the associated Custom Resources namespace/project admins/editors need edit access to the Grafana Custom Resources.

```shell
kubectl create -f cluster_role_aggregate_grafana_admin_edit.yaml
kubectl create -f deploy/cluster_roles/cluster_role_aggregate_grafana_admin_edit.yaml
kubectl create -f deploy/cluster_roles/cluster_role_aggregate_grafana_view.yaml
```

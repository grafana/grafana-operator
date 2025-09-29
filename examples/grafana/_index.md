---
title: "Grafanas"
linkTitle: "Grafanas"
weight: 10
---

The grafana Custom Resource Definition (CRD) exist to manage one or multiple managed and external grafana instances.

A basic Grafana deployment of Grafana with a dashboard.

{{< readfile file="./basic-grafana.yaml" code="true" lang="yaml" >}}

## Where should the operator look for Grafana resources?

The operator needs extensive access to be able to manage Grafana instances. Depending on your needs, you might want to restrict the operator to specific namespaces.

To support that, we offer 4 operational modes (you can switch between those through `WATCH_NAMESPACE` environment variable):

- cluster-wide (`WATCH_NAMESPACE: ""`);
- multiple namespaces (`WATCH_NAMESPACE: "grafana, foo"`):
  - Cluster-wide permissions are still required;
- single namespace (`WATCH_NAMESPACE: "grafana"`):
  - With this mode, it's possible to use `Role` + `RoleBinding` to grant the operator the required access.
- multiple namespaces using label selector (`WATCH_NAMESPACE_SELECTOR: "environment: dev"`):
  - With this mode, it is possible detect and load all namespaces that match the label selector automatically.
    New namespaces won't be automatically included until the Grafana operator is restarted.
  - Cluster-wide permissions are still required;

## Delete instances

Deleting instances will clean up all associated resources *except* associated volumes.
Persistent Volume Claims are not deleted to prevent data loss on accidental deletion.
If you want to recreate an instance, be sure to delete the volume as well.
Otherwise, the new instance will start up with the old database and encounter authentication issues.

## Organizations

There have been much design work around how it could be done, but no one have managed to come up with a good design that would be simple-to-use for end users and be easy-to-manage code-wise from maintainer's perspective.
Instead we suggest that you use multiple grafana instances together with good CI/CD solutions to manage your dashboards, datasources, etc.
If you really need support for organizations, you can use the `spec.client.headers` map to set the `X-Grafana-Org-Id` Header.

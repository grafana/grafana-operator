---
title: Grafana
weight: 11
---

The grafana Custom Resource Definition (CRD) exist to manage one or multiple managed and external grafana instances.

## Grafana config

We offer the `grafana.config` field where you can pass any Grafana configuration values you want.

The operator does not make any extra validation of your configuration, so just like a non-operator deployment of Grafana, your Grafana instance might be broken due to a configuration error.

To find all possible configuration options, look at the [official documentation](https://grafana.com/docs/grafana/latest/setup-grafana/configure-grafana/).

In case you need to specify top level options like `app_mode` or `instance_name`, specify them in the `global` section like this:

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  name: grafana
  labels:
    dashboards: "grafana"
spec:
  config:
    global:
      app_mode: "development"
```

## Where should the operator look for Grafana resources?

The operator needs extensive access to be able to manage Grafana instances. Depending on your needs, you might want to restrict the operator to specific namespaces.

To support that, we offer 4 operational modes (you can switch between those through `WATCH_NAMESPACE` environment variable):

- cluster-wide (`WATCH_NAMESPACE: ""`);
- multiple namespaces (`WATCH_NAMESPACE: "grafana, foo"`):
  - Cluster-wide permissions are still required;
- single namespace (`WATCH_NAMESPACE: "grafana"`):
  - With this mode, it's possible to use `Role` + `RoleBinding` to grant the operator the required access.
- mutliple namespaces using label selector (`WATCH_NAMESPACE_SELECTOR: "environment: dev"`):
  - With this mode, it is possible detect and load all namespaces that match the label selector automatically.
    New namespaces won't be automatically included until the Grafana operator is restarted.
  - Cluster-wide permissions are still required;

## External Grafana instances

The operator is able to work with external Grafana instances (pretty much any SaaS offering) with a few limitations in mind:

- `grafana.spec.config` has no effect as the operator cannot mount a custom `grafana.ini` file into the instance;
- not possible to install plugins as the operator cannot overwrite `GF_INSTALL_PLUGINS` environment variable of the instance.

The `grafana.spec.external` allows you to set a URL and provide credentials for external Grafana. Example:

```yaml
---
kind: Secret
apiVersion: v1
metadata:
  name: grafana-admin-credentials
stringData:
  GF_SECURITY_ADMIN_USER: root # Username
  GF_SECURITY_ADMIN_PASSWORD: secret # Password
type: Opaque
---
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  name: external-grafana
  labels:
    dashboards: "external-grafana"
spec:
  external:
    url: http://test.io # Grafana URL
    adminPassword:
      name: grafana-admin-credentials
      key: GF_SECURITY_ADMIN_PASSWORD
    adminUser:
      name: grafana-admin-credentials
      key: GF_SECURITY_ADMIN_USER
```

A more comprehensive example can be found [here](../examples/external_grafana/readme).

## Delete instances

Deleting instances will clean up all associated resources *except* associated volumes.
Persistent Volume Claims are not deleted to prevent data loss on accidental deletion.
If you want to recreate an instance, be sure to delete the volume as well.
Otherwise, the new instance will start up with the old database and encounter authentication issues.

## Organizations

There have been much design work around how it could be done, but no one have managed to come up with a good design that would be simple-to-use for end users and be easy-to-manage code-wise from maintainer's perspective.
Instead we suggest that you use multiple grafana instances together with good CI/CD solutions to manage your dashboards, datasources, etc.
If you really need support for organizations, you can use the `spec.client.headers` map to set the `X-Grafana-Org-Id` Header.

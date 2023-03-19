---
title: Grafana
weight: 11
---

The grafana Custom Resource Definition (CRD) exist to manage one or multiple managed and external grafana instances.

## Grafana config

In version 4 of the operator we defined all the configuration values that could be done in the CRD.
In version 5 we don't longer do that, instead we supply you with `grafana.config` where you can add any
grafana configuration values that you want.

The operator do not make any extra validation of your configuration so just like Grafana your Grafana instance might be broken due to configuration error.

To find all possible configuration options look at the [official documentation](https://grafana.com/docs/grafana/latest/setup-grafana/configure-grafana/).

## Where should the operator look for Grafana resources?

The operator needs lots of access to be able to manage Grafana instances.
Depending on your needs you might want to limit the Grafana operator to only watch specific namespaces.

Due to this we have given the operator 3 deployment modes.

- cluster wide
- single namespace
- multiple namespaces

To enter these specific modes you need to define the `WATCH_NAMESPACE` environment variable in the Grafana operator.
How you do this will differ between the different deployment methods.

`WATCH_NAMESPACE: ""`
Will make the operator work cluster wide.

`WATCH_NAMESPACE: "grafana"`
Will make the operator only watch the grafana namespace, using this mode, makes it possible to use a role instead of a cluster role to give access to the operator.
Depending on which deployment solution you use how to do this will differ.

`WATCH_NAMESPACE: "grafana, monitor, foo, bar"`
Will make the operator watch the following namespaces grafana, monitor, foo, bar. This will force you to still use the cluster wide rbac settings to enable the Grafana operator to monitor all of those namespaces.

## External Grafana instances

With the operator you can manage external Grafana instances.
So even if you don't want to manage Grafana on your own you can still use dashboards as code.

The `grafana.spec.external` config allows you to set an URL and provide authentication to talk to your Grafana instance.

Since the operator do not own the external Grafana instance you are **not** able to send in any configuration values to grafana it self. So for example `grafana.spec.config` will not do anything if you want to manage an external Grafana instances.
This is due to Grafana only have a ini files that it looks in during start and don't have any API that we can use to configure it.

You are also not able to install any plugins using the operator since they are also installed during startup.

This is how you can setup an External grafana instance, we assume that you have created a secret called `grafana-admin-credentials` that contains the correct keys.

```yaml
---
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  name: external-grafana
  labels:
    dashboards: "external-grafana"
spec:
  external:
    url: http://test.io
    adminPassword:
      name: grafana-admin-credentials
      key: GF_SECURITY_ADMIN_PASSWORD
    adminUser:
      name: grafana-admin-credentials
      key: GF_SECURITY_ADMIN_USER
```

To see the entire example you can look [here](../examples/external_grafana/readme)

## Organizations

In grafana-operator v4 there have been multiple requests around adding support for Grafana organizations.
There have been much design work around how it could be done but no one have managed to come up with a good design that would be simple to use as a developer and be possible to manage from a code perspective.

Since version 5 now supports multiple Grafana instances we are taking the same stance as Grafana cloud does and we will not support organizations in the operator.

Instead we suggest that you use multiple grafana instances together with good CI/CD solutions to manage your dashboards datasources etc.

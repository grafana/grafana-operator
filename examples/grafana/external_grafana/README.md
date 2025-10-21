---
title: "External Grafana Instances"
linkTitle: "External Grafana Instances"
weight: 20
---

## External Grafana instances

The operator is able to work with external Grafana instances (pretty much any SaaS offering) with a few limitations in mind:

- `.spec.config` has no effect as the operator cannot mount a custom `grafana.ini` file into the instance;
- not possible to install plugins as the operator cannot overwrite the `GF_INSTALL_PLUGINS` environment variable of the instance.

The `.spec.external` requires a URL and credentials to manage the Grafana instance:

```yaml
---
kind: Secret
apiVersion: v1
metadata:
  name: grafana-admin-credentials
stringData:
  GF_SECURITY_ADMIN_USER: root # Username
  GF_SECURITY_ADMIN_PASSWORD: secret # Password
  # service_account_token
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
    # Using an service accountt apiKey instead of basic auth is also an option with external instances
    # apiKey:
    #   name: grafana-admin-credentials
    #   key: service_account_token
```

## Internal and External in one

In this case we manage a Grafana instance through the operator as if it were two separate instances.

Notice the `instanceSelector` of the grafana dashboard and the grafana instances.
It's only the `external-grafana` that will match with the grafana dashboard called `external-grafanadashboard-sample`.

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

If you look in the status of external-grafana you will also see the hash of the dashboard that have been applied.

```shell
kubectl get grafanas.grafana.integreatly.org external-grafana -o yaml
... # Redacted
status:
  adminUrl: http://test.io
  dashboards:
  - grafana-operator-system/external-grafanadashboard-sample/cb1688d2-547a-465b-bc49-df3ccf3da883
  stage: complete
  stageStatus: success
```

## FYI

If you want run the same test locally on your computer remember to ether update the ingress host to match your settings or change `client.preferIngress` to false assuming that you run the operator within the cluster.

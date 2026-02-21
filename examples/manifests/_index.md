---
title: "Manifests"
weight: 60
---

`GrafanaManifest` is our way to support resources not natively implemented in the Grafana operator.

By utilizing the new kubernetes style APIs available in Grafana, this allows you to use the Grafana operator with resources provided by plugins and more.


For example, this manifest configures a playlist using the `GrafanaManifest` resource:
{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

The above example also showcases how to use the `patch` capability to dynamically replace values in the manifest.


## Configuring the namespace

In single tenant installations of Grafana (such as instances created by the operator), all resources live in the `default` namespace.
If this is not the case for your instance, you'll either need to specify the namespace in the `Grafana` resource or set it on every `GrafanaManifest` resource.

When using Grafana Cloud, you can find your namespace information in the account console.
Go to the details page of your Grafana instance for the desired stack and look for a field called _Instance ID_.
Use this value as the `tenantNamespace` in your `Grafana` instance, prefixed by `stacks-`.

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  name: your-gcloud-instance
spec:
  external:
    apiKey:
      key: SERVICE_ACCOUNT_TOKEN
      name: grafana-cloud-credentials
    tenantNamespace: stacks-<your-instance-id>
    url: https://<your-stack>.grafana.net
```

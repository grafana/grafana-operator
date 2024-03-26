---
title: Common options
weight: 11
---

## ResyncPeriod

Grafana doesn't have any webhooks or similar ways of giving information to the operator that a Grafana resource, like a dashboard, has changed.
Due to this the Grafana operator has to constantly poll the Grafana API to test for changes in the dashboard.

To avoid control how often this polling should occur, you can set the `spec.resyncPeriod` field.
This field tells the operator how often it should poll the Grafana instance for changes.

So, if for example, a dashboard has changed, the operator comes in and overwrite those settings after `5m` by default.
If you never want the operator to poll for changes in the dashboards you need to set this value to `0m`:

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafanadashboard-from-url
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  url: "https://grafana.com/api/dashboards/7651/revisions/44/download"
  resyncPeriod: 0m
```

This can of course be annoying for your dashboard developers. The recommended workflow is to copy the dashboard and work on the copy instead.
When you finish your changes, export the changes and update the dashboard CR.

## InstanceSelector

The `spec.instanceSelectors` field is used to tell the operator which Grafana instance the resource applies to.

It's a regular [Kubernetes label selector](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/) and for example used in deployments.

To link resources to instances, first define a label on your Grafana instance:

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  name: grafana
  labels:
    dashboards: "grafana" # Notice the label
spec: ...
```

And in the resource, set the `spec.instanceSelector` to match the previously set labels.

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafanadashboard-sample
spec:
  resyncPeriod: 30s
  instanceSelector:
    matchLabels:
      dashboards: "grafana" # Notice the label
  json: ...
```

## Cross namespace Grafana instances

To avoid exposing possible sensitive data sources to arbitrary instances, resources only affect instances in the same namespace.

In case you need this functionality for your specific setup, use the `spec.allowCrossNamespaceImport` field of the Grafana instance.
This setting allows resources in arbitrary namespaces to be applied to the Grafana instance.

More information can be found in [#44](https://github.com/grafana-operator/grafana-operator-experimental/issues/44).

## Using a proxy server

The Operator can use a proxy server when fetching URL-based / Grafana.com dashboards or making requests to external Grafana instances.
[Environment variables](https://pkg.go.dev/golang.org/x/net/http/httpproxy#FromEnvironment) control the proxy

## Deleting resources with a finalizer

The operator marks some resources with [finalizers](https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers/) to clean up resources on deletion.
If the operator isn't running, marked resources are unable to be deleted. This is intended.
However, this behavior can cause issues if, for example, you uninstalled the operator and now can't remove resources.
To manually remove the finalizer, use the following command:

```bash
kubectl patch GrafanaAlertRuleGroup <rule-group-name> -p '{"metadata":{"finalizers":null}}' --type=merge
```

After running this, the resource can be deleted as usual.

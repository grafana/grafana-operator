---
title: Common options
weight: 20
---

## ResyncPeriod

Grafana doesn't have any webhooks or similar ways of notifying the operator that a Grafana resource, like a dashboard, has changed.
Due to this the Grafana operator constantly polls the Grafana API to test for changes and overwrite the resources, reconciling towards the desired state.

We describe this loop as a synchronizing resources with Grafana instances.

To control how often this polling should occur, you can set the `spec.resyncPeriod` field.
This field tells the operator how often it should poll the Grafana instance for changes for the specific resource.

If a dashboard has changed, the operator will overwrite and synchronize the dashboard after `10m` by default.

This can of course be annoying for developers actively updating a resource. The recommended workflow is to duplicate the dashboard/alert/other and work on the copy.
When finished, export the changes and update the resource manifest to update the original.

This can be disabled by setting a the value to `0m`

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafanadashboard-from-url
spec:
  resyncPeriod: 0m
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  url: "https://grafana.com/api/dashboards/7651/revisions/44/download"
```

The `resyncPeriod` is applied on successful synchronizations.

When an error is encountered during a synchronization, like a request timeout, the operator repeatedly retries the full synchronization with an exponential backoff.

The first retry is immediate, the second retry is delayed by a second and so on.

{{% alert title="Warning" color="warning" %}}
Even after setting `resyncPeriod` to `0m`, the operator will still sync the resource whenever it changes or the operator is restarted
{{% /alert %}}

## InstanceSelector

The `spec.instanceSelectors` field tells the operator which Grafana instance the resource applies to.

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
  instanceSelector:
    matchLabels:
      dashboards: "grafana" # Notice the label
  json: ...
```

As it is a normal label selector, it is possible to leave it empty to match everything.

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: match-all-in-namespace
spec:
  instanceSelector: {}
  json: ...
```

Additionally, the more powerful `matchExpressions` are available and can be used alone or in combination with `matchLabels`

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: match-expressions
spec:
  instanceSelector:
    matchExpressions:
    - { key: dashboards, operator: In, values: ["grafana", "loki", "tempo"]}
    - { key: datasources, operator: NotIn, values: ["internal"]}
  json: ...
```

## AllowCrossNamespaceImport

Allows resources in one namespace to be applied to Grafana instances in other namespaces as well.

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: match-across-namespaces
spec:
  allowCrossNamespaceImport: true
  instanceSelector: {}
  json: ...
```

Disabled by default to avoid exposing possibly sensitive resources by accident.

More information can be found in [#44](https://github.com/grafana-operator/grafana-operator-experimental/issues/44).

`allowCrossNamespaceImport` is one-way mutable, you can always enable it (`true`) but never disable.

Disabling requires a full recreate, delete and apply.

This helps guaranteeing proper cleanup across all matched Grafana instances.

## Suspending resources

It can in many scenarios be useful to temporarily halt the reconciliation and synchronization of resources.

This can be achieved with the `.spec.suspend` option on all resources, with the exception of `GrafanaNotificationPolicyRoutes`.

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: ...
metadata:
  name: suspended
spec:
  suspend: true
...
status:
  conditions:
  - lastTransitionTime: "2025-07-06T15:58:29Z"
    message: Resource changes are ignored
    observedGeneration: 2
    reason: ApplySuspended
    status: "True"
    type: Suspended
  lastResync: "2025-07-06T15:58:29Z"
```

When `.spec.suspend` is `true` The Operator will ignore any changes where they are normally synchronized immediately.

## Using a proxy server

The Operator can use a proxy server when fetching URL-based / Grafana.com dashboards or making requests to external Grafana instances.
[Environment variables](https://pkg.go.dev/golang.org/x/net/http/httpproxy#FromEnvironment) control the proxy

## Deleting resources with a finalizer

The operator uses [finalizers](https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers/) to help clean up resources from all instances on deletion.
If the operator isn't running, marked resources are unable to be deleted. This is intended.
However, this behavior can cause issues if, for example, you uninstalled the operator and now can't remove resources.
To manually remove the finalizer, use the following command:

```bash
kubectl patch GrafanaAlertRuleGroup <rule-group-name> -p '{"metadata":{"finalizers":null}}' --type=merge
```

After running this, the resource can be deleted as usual.

---
title: "Documentation"
linkTitle: "Documentation"
weight: 20
menu:
  main:
    weight: 20
---

Our official documentation for the operator.
Hopefully you will find everything you need in here, if not feel free to open an issue, write on slack or even better submit a pr.

## Examples

Just like in v4 we have a number of [examples](examples/) to look at.

## ResyncPeriod

Grafana doesn't have any webhooks or similar ways of giving information to the operator that a grafana resource, like a dashboard, has been changed.
Due to this the Grafana operator has to constantly check the Grafana API to see if something changed in the dashboard.

That is why we introduced `spec.resyncPeriod`, this is a configuration that makes it possible to tell the operator
how often it should check with the Grafana instance if the dashboard matches the settings that are defined in the Kubernetes CR.

So, if for example, a dashboard is changed, the operator will come in and overwrite those settings after `5m` by default.
If you never want the operator to check if the dashboards have changed you need to set this value to `0m`:

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

This can of course be annoying for your dashboard developers. But we recommend that before doing any change to a dashboard in the Grafana UI that you first copy the existing dashboard and work on the copy instead.
When you are finished with your changes export the changes and update the dashboard CR.

## InstanceSelector

For the operator to know which dashboard, datasource or folder that should be applied to a specific Grafana instance we use `instanceSelectors`.

It's the normal label selector that is built in to [Kubernetes](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/) and for example used in deployments.
So you should hopefully be fairly used to how it works.

But in short you define a label on your Grafana instance.

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  name: grafana
  labels:
    dashboards: "grafana" # Notice the label
spec: ...
```

And for example in the dashboard we define a instanceSelector.

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

## Cross namespace grafana instances

As described in [#44](https://github.com/grafana-operator/grafana-operator-experimental/issues/44) we didn't want it
to be to easy to get access to a grafana datasource that wasn't defined the same namespace as the grafana instance.

To solve this we introduced `spec.allowCrossNamespaceImport` option to, dashboards, datasources and folders to be false by default.
This setting makes it so a grafana instance in another namespace don't get the grafana resources applied to it even if the label matches.

This is because especially the data sources contain secret information and we don't want another team to be able to use your datasource unless defined to do so in both CR:s.

## Using a proxy server

The Operator can use a proxy server when fetching URL-based / Grafana.com dashboards or making requests to external Grafana instances.
The proxy settings can be controlled through environment variables as documented [here](https://pkg.go.dev/golang.org/x/net/http/httpproxy#FromEnvironment).

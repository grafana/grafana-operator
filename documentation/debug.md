# Debug grafana-operator

This document is created to help people debug the grafana-operator and it's Custom Resources (CR).

First of all read the documentation, they are not perfect but they will help you allong in most cases.

## Grafana instance

After you have setup your [grafana instance](deploy_grafana.md) it should at least generate a kubernetes deployment, but it can also generate more depending on how you have connfigured your grafana instance.

If that is not the case you need to debug it.

Start by checking the events and status field of the the grafana CR.

```shell
$ kubectl describe grafana example-grafana

Status:
  Message:                Ingress.extensions "grafana-ingress" is invalid: spec.rules[0].http.paths[0].pathType: Required value: pathType must be specified
  Phase:                  failing
  Previous Service Name:  grafana-service
Events:
  Type     Reason           Age               From              Message
  ----     ------           ----              ----              -------
  Warning  ProcessingError  9s (x4 over 27s)  GrafanaDashboard  Ingress.extensions "grafana-ingress" is invalid: spec.rules[0].http.paths[0].pathType: Required value: pathType must be specified
```

`kubectl get grafana example-grafana -o yaml` will also give you the status output.

For example in this case my grafana CR is missing the definition of `ingress.pathType`

```.yaml
spec:
  ingress:
    enabled: True
    pathType: Prefix
```

In general if you are having issues setting up your grafana instance simply your grafana CR.
First try [deploy/examples/Grafana.yaml](deploy/examples/Grafana.yaml) and get that
grafana instance up and running.
When that is done add small config changes to start build the config you want.

## Grafana Dashboards

There are generally three reasons why you can't find your grafana dashboard.

1. You are creating a grafanadashboard in another namespace then the grafana-operator
  but you haven't defined --scan-all in the [operator](deploy_grafana.md).
2. Or you haven't confgiured cluster wide RBAC [config](deploy/cluster_roles/README.md).
3. Or you haven't defined the correct labelSelector that matches the operator on your dashboard.

In your grafana CR you have deinfed somehing like:

```.yaml
  dashboardLabelSelector:
    - matchExpressions:
        - { key: app, operator: In, values: [grafana] }
```

Just like in [deploy/examples/dashboards/SimpleDashboard.yaml](deploy/examples/dashboards/SimpleDashboard.yaml)
you need to define the correct labels that matches the grafana CR.

```.yaml
apiVersion: integreatly.org/v1alpha1
kind: GrafanaDashboard
metadata:
  name: simple-dashboard
  labels:
    app: grafana
spec:
  json: >
    {
```

Another good place to see if the dashboard have been applied to the grafana dashboard
is to check the kubernetes events.

``` shell
$ kubectl get events

2d16h       Warning   ProcessingError                grafanadashboard/simple-dashboard                           Get "http://admin:***@grafana-service.grafana-operator-system.svc.cluster.local:3000/api/folders": dial tcp 10.96.171.19:3000: connect: connection refused
2d16h       Normal    Success                        grafanadashboard/simple-dashboard                           dashboard grafana-operator-system/simple-dashboard successfully submitted
```

For example the events that you can see above the grafana dashboard CR:s was most likley created before
the grafana CR was up and running and thus coulden't answer the grafana operator when it tried
to apply the dashboard.

## Operator logs

Check the logs from the operator and see what you can find.

Depending on how you have setup the operator you should be able to get the logs running:

```shell
kubectl logs -l control-plane=controller-manager -c manager
```

## Bugs

If you still are unable to deploy the grafana instance that you want and you think it's a bug please create an issue.

When creating the issue please add as much information as possible from this debug documentation.
It will make it easier for us to debug the issues faster.

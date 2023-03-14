---
title: "External grafana"
linkTitle: "External grafana"
---
A basic example configuring an external grafana instance.

In this case we create a grafana instance through the operator just to showcase that it can be done.

Notice the `instanceSelector` of the grafana dashboard and the grafana instances.
You will notice that it's only the grafana instance called `external-grafana` that will match with the grafana dashboard called `external-grafanadashboard-sample`.

If you look in the status of external-grafana you will also see the hash of the dashboard that have been applied.

```shell
kubectl get grafanas.grafana.integreatly.org external-grafana -o yaml
# redacted output
status:
  adminUrl: http://test.io
  dashboards:
  - grafana-operator-system/external-grafanadashboard-sample/cb1688d2-547a-465b-bc49-df3ccf3da883
  stage: complete
  stageStatus: success
```

## FYI

If you want run the same test locally on your computer remember to ether update the ingres host to match your settings or change `client.preferIngress` to false assuming that you run the operator within the cluster.

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

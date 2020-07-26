# Jsonnet support

The operator supports importing dashboards in [jsonnet](https://jsonnet.org/) format. The [grafonnet](https://grafana.github.io/grafonnet-lib/) library is available out of the box, other libraries can be imported via config maps.

## Creating a jsonnet dashboard

As with `json`, Grafana Dashboard CRs have a `jsonnet` field:

 ```yaml
apiVersion: integreatly.org/v1alpha1
kind: GrafanaDashboard
metadata:
  name: grafonnet-dashboard
  labels:
    app: grafana
spec:
  name: simple-dashboard.json
  jsonnet: |-
    <jsonnet source code goes here>
 ```

The `grafonnet` library is provided by the operator and can be imported using:

```
local grafana = import 'grafonnet/grafana.libsonnet';
```

## Creating jsonnet libraries

Jsonnet libraries can be imported from config maps in the same namespace as the operator:

```yaml
kind: ConfigMap
apiVersion: v1
metadata:
  name: monitoring
  namespace: grafana
  labels:
    app: grafana
  annotations:
    jsonnet/library: "true"
data:
  monitoring.libsonnet: |-
    <jsonnet source code goes here>
```

Config maps must have the `jsonnet/library: "true"` annotation in order to be recognized by the operator.
They also need to have one or more labels that match a selector given in the Grafana CR:

```yaml
spec:
  jsonnet:
    libraryLabelSelector:
      matchLabels:
        app: grafana
```

The libary will be imported as a file with the name `monitoring.libsonnet` into a directory with the same name of the config map.
It can be imported in a dashboard using the following code:

```
local monitoring = import 'monitoring/monitoring.libsonnet';
```

*NOTE*: The keys of the config map must be valid filenames, and the extension must be `.libsonnet`

*NOTE*: Multiple jsonnet files can be in the same config map
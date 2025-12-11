---
title: Dashboards
weight: 40
tags:
  - Folders
---

[Dashboards](https://grafana.com/docs/grafana/latest/dashboards/) is the core feature of Grafana and of course something that you can manage through the operator.

To view the entire configuration that you can do within Dashboards, look at our [API documentation](/docs/api/#grafanadashboardspec).

## Dashboard management

You can configure and reference dashboards as code in many different ways.

- [JSON](#json)
- [gzipJson](#gzipjson)
- [URL](#url)
- [Jsonnet](#jsonnet) (Deprecated)
- [JaaS](#jaas)
- [ConfigMap](#configmap)

To view all configuration options for folders, look at our [API documentation](/docs/api/#grafanadashboardspec).

### JSON

A pure JSON representation of your Grafana dashboard.
Normally you would create your dashboard manually within Grafana, when you have come up with how you want the dashboard to look like, you export it as JSON,
grab the JSON using the export function in grafana and put inside the GrafanaDashboard CR.

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafanadashboard-sample
spec:
  resyncPeriod: 30s
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  json: >
    {
      "id": null,
      "title": "Simple Dashboard",
      "tags": [],
      "style": "dark",
      "timezone": "browser",
      "editable": true,
      "hideControls": false,
      "graphTooltip": 1,
      "panels": [],
      "time": {
        "from": "now-6h",
        "to": "now"
      },
      "timepicker": {
        "time_options": [],
        "refresh_intervals": []
      },
      "templating": {
        "list": []
      },
      "annotations": {
        "list": []
      },
      "refresh": "5s",
      "schemaVersion": 17,
      "version": 0,
      "links": []
    }
```

### gzipJson

It's just like JSON but instead of adding pure JSON to the dashboard CR you add a gzipped representation.
This allows you to do really **big** dashboards while not hitting the etcd maximum request size of 1,5 MiB.

Assuming a dashboard is already saved in `dashboard.json`, the steps below describe how you can prepare a CR.

```shell
cat dashboard.json | gzip | base64 -w0
```

Take the output and put it in your GrafanaDashboard CR, for example:

```yaml
---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafanadashboard-gzipped
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  gzipJson: |-
    H4sIAAAAAAAAA4WQQU/DMAyF7/0VVc9MggMgcYV/AOKC0OQubmM1jSPH28Sm/XfSNJ1WcaA3f+/l+dXnqk5fQ6Z5qf3eubt5VlKHCTXvNAaH9RtE2zKI2fQnCgFNsxihj8n39V3mqD/zQwMyXE004ol95q3wMaIsEhpSaPMTlT0WasngK3sVdlN6By4uUi8Q7AezUwpJeig4gEe3ajItTfM5T5l0wuNUwfNx82RLg9nLhTeZXW4iAu2GVHcVNPEtByX2tyuzJtgJRrslrygHKJ3WsZhuCkq+X8c6ivrXDd6zwrLrX3vZP/3PY1yuHHcWR/hEiSlmutpzEQ5XdF+IIz+Uzpeq+gWtMMT1HwIAAA==
```

[Example documentation](./gzip_json/readme).

### URL

Probably the easiest way to get started to add dashboards to your Grafana instances.

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafanadashboard-from-grafana
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  url: "https://grafana.com/api/dashboards/1860/revisions/30/download"
```

{{% alert title="Note" color="primary" %}}
You don't have to rely on Grafana Dashboard registry for this, any URL reachable by the operator would work.
{{% /alert %}}

[Example documentation](./url/readme).

### Jsonnet

The Jsonnet dashboard type is deprecated. It uses the old and now unmaintained [grafonnet-lib](https://github.com/grafana/grafonnet-lib) library. Users who rely on Jsonnet based dashboards should switch to [JaaS](#jaas) instead which supports the new [grafonnet](https://github.com/grafana/grafonnet) library as well as any additional custom libraries you have created yourself. See the [discussion](https://github.com/grafana/grafana-operator/discussions/2171) for more details.

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafanadashboard-jsonnet
spec:
  resyncPeriod: 30s
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  jsonnet: |
   local grafana = import 'grafonnet/grafana.libsonnet';
   local dashboard = grafana.dashboard;
   local row = grafana.row;
   local singlestat = grafana.singlestat;
   local prometheus = grafana.prometheus;
   local template = grafana.template;

   dashboard.new(
     'JVM',
     tags=['java'],
   )
   .addTemplate(
     grafana.template.datasource(
       'PROMETHEUS_DS',
       'prometheus',
       'Prometheus',
       hide='label',
     )
   )
   .addTemplate(
     template.new(
       'env',
       '$PROMETHEUS_DS',
       'label_values(jvm_threads_current, env)',
       label='Environment',
       refresh='time',
     )
   )
   .addTemplate(
     template.new(
       'job',
       '$PROMETHEUS_DS',
       'label_values(jvm_threads_current{env="$env"}, job)',
       label='Job',
       refresh='time',
     )
   )
   .addTemplate(
     template.new(
       'instance',
       '$PROMETHEUS_DS',
       'label_values(jvm_threads_current{env="$env",job="$job"}, instance)',
       label='Instance',
       refresh='time',
     )
   )
   .addRow(
     row.new()
     .addPanel(
       singlestat.new(
         'uptime',
         format='s',
         datasource='Prometheus',
         span=2,
         valueName='current',
       )
       .addTarget(
         prometheus.target(
           'time() - process_start_time_seconds{env="$env", job="$job", instance="$instance"}',
         )
       )
     )
   )
```

## JaaS

[Jsonnet-as-a-Service](https://github.com/metio/jaas) (JaaS) is a webservice that evaluates Jsonnet snippets which allows you to reference your Jsonnet based dashboards by [URL](#url) as explained above. It comes with [Grafonnet](https://grafana.github.io/grafonnet/) pre-installed and supports user supplied custom libraries as well.

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafanadashboard-jaas
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  url: "http://jaas.jaas.svc.cluster.local:8080/jsonnet/your-dashboard"
```

[Example documentation](./jaas/readme).

## Plugins

[Plugins](https://grafana.com/grafana/plugins/) is a way to extend the grafana functionality in dashboards and datasources.

Plugins can be installed to grafana instances managed by the operator and be defined in both datasources and dashboards.

They **cannot** be installed using external grafana instances due to how the installation of plugins is done at the start of the instance using environment variables which is a built in feature in grafana.

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: keycloak-dashboard
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  plugins:
    - name: grafana-piechart-panel
      version: 1.3.9
  # NOTE: the json block below is incomplete as it's not the main focus of the example
  json: >
    {
      "__inputs": [
        {
          "name": "DS_PROMETHEUS",
          "label": "Prometheus",
          "description": "",
          "type": "datasource",
          "pluginId": "prometheus",
          "pluginName": "Prometheus"
        }
      ],
    }
```

Look here for more examples on how to install [plugins](./plugins/readme)

## Content cache duration

To not constantly perform requests to external URL every time a dashboard reconcile or a resync period expires we save URLs in a cache in the operator.
This cache is saved in the status field of the dashboard CR.

By default this cache is `24h` long, you can change this value by setting contentCacheDuration manually per dashboard.

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
  contentCacheDuration: 48h
```

Remember, depending on where you get your dashboards you might become rate limited if you have multiple dashboards with relatively short `contentCacheDuration` or if all the requests happens at the same time.

## Dashboard uid management

Whenever a dashboard is imported into a Grafana, it gets assigned a random `uid` unless it's hardcoded in dashboard's code. Random `uid` is undesirable from the operator's perspective as it would create the need to track those uids across Grafana instances.

To mitigate the scenario, if `uid` is not hardcoded, the operator will insert the value taken from CR's `metadata.uid` (this value is automatically generated by Kubernetes itself for all resources).

## Select the folder where the dashboard will be deployed

By default, a dashboard will appear in a Folder with the name of the namespace

```yaml
---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafana-folder-default
  labels:
    dashboards: "grafana"
spec:
  # the folder will be "default" by default
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  json: |
    {
      "id": null,
      "title": "Folder Without Folder Defined",
      "tags": [],
      "style": "dark",
      "timezone": "browser",
      "editable": true,
      "hideControls": false,
      "graphTooltip": 1,
      "panels": [],
      "time": {
        "from": "now-6h",
        "to": "now"
      },
      "timepicker": {
        "time_options": [],
        "refresh_intervals": []
      },
      "templating": {
        "list": []
      },
      "annotations": {
        "list": []
      },
      "refresh": "5s",
      "schemaVersion": 17,
      "version": 0,
      "links": []
    }
```

If you want to put the dashboard in a specific folder, you have two choices:

* Use an `GrafanaFolder` resource as a reference:

```yaml
---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaFolder
metadata:
  name: folder-ref
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  title: "Folder"

---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafana-folderref
  labels:
    dashboards: "grafana"
spec:
  folderRef: folder-ref
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  json: |
    {
      "id": null,
      "title": "Folder Ref With Folder",
      "tags": [],
      "style": "dark",
      "timezone": "browser",
      "editable": true,
      "hideControls": false,
      "graphTooltip": 1,
      "panels": [],
      "time": {
        "from": "now-6h",
        "to": "now"
      },
      "timepicker": {
        "time_options": [],
        "refresh_intervals": []
      },
      "templating": {
        "list": []
      },
      "annotations": {
        "list": []
      },
      "refresh": "5s",
      "schemaVersion": 17,
      "version": 0,
      "links": []
    }
```

* Use a static UID as a folder reference


```yaml
---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafana-folderuid
  labels:
    dashboards: "grafana"
spec:
  folderUID: ec94e912-02f4-463b-9968-1f5f7db3531f
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  json: |
    {
      "id": null,
      "title": "Folder UID With Folder",
      "tags": [],
      "style": "dark",
      "timezone": "browser",
      "editable": true,
      "hideControls": false,
      "graphTooltip": 1,
      "panels": [],
      "time": {
        "from": "now-6h",
        "to": "now"
      },
      "timepicker": {
        "time_options": [],
        "refresh_intervals": []
      },
      "templating": {
        "list": []
      },
      "annotations": {
        "list": []
      },
      "refresh": "5s",
      "schemaVersion": 17,
      "version": 0,
      "links": []
    }
```

## Custom folders

{{% alert title="Warning" color="secondary" %}}
This method is not recommended. Prefer to use the GrafanaFolder CR with the `folderRef` field, or `folderUID` with the UID of an existing folder instead.
{{% /alert %}}

In a standard scenario, the operator would use the namespace a CR is deployed to as a folder name in grafana. `folder` field can be used to set a custom folder name:

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafanadashboard-with-custom-folder
spec:
  folder: "Custom Folder"
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  url: "https://raw.githubusercontent.com/grafana-operator/grafana-operator/master/examples/dashboard_from_url/dashboard.json"
```

{{% alert title="Note" color="primary" %}}
the `.spec.folder` field is ignored when either `.spec.folderUID` or `.spec.folderRef` is present in the GrafanaDashboard declaration.
{{% /alert %}}

## Dashboard customization by providing environment variables

Will be pleasant for scenarios when you would like to extend the behaviour of jsonnet generation by parametrizing it with runtime Env vars:

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafanadashboard-jsonnet
spec:
  resyncPeriod: 30s
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  envs:
    - name: API_VERSION
      value: "1.0.0"
    - name: ENV_FROM_CM # just example, such cm and secrets are not provided by vendor
      valueFrom:
        configMapKeyRef:
          name: custom-grafana-dashboard-cm
          key: GRAFANA_URL
    - name: ENV_FROM_SECRET
      valueFrom:
        secretKeyRef:
          name: custom-grafana-dashboard-secrets
          key: PROMETHEUS_USERNAME
  envFrom: # just example, such cm and secrets are not provided by vendor
    - configMapRef:
        name: custom-grafana-dashboard-cm
    - secretRef:
        name: custom-grafana-dashboard-secrets
  jsonnet: |
   local grafana = import 'grafonnet/grafana.libsonnet';
   local dashboard = grafana.dashboard;
   local row = grafana.row;
   local singlestat = grafana.singlestat;
   local prometheus = grafana.prometheus;
   local template = grafana.template;
   local labelFromEnvs = std.extVar('API_VERSION');

   dashboard.new(
     'JVM',
     tags=['java'],
   )
   .addTemplate(
     grafana.template.datasource(
       'PROMETHEUS_DS',
       'prometheus',
       'Prometheus',
       hide='label',
     )
   )
   .addTemplate(
     template.new(
       'env',
       '$PROMETHEUS_DS',
       'label_values(jvm_threads_current, env)',
       label='Environment',
       refresh='time',
     )
   )
   .addTemplate(
     template.new(
       'job',
       '$PROMETHEUS_DS',
       'label_values(jvm_threads_current{env="$env"}, job)',
       label='Job',
       refresh='time',
     )
   )
   .addTemplate(
     template.new(
       'instance',
       '$PROMETHEUS_DS',
       'label_values(jvm_threads_current{env="$env",job="$job"}, instance)',
       label='Instance',
       refresh='time',
     )
   )
   .addRow(
     row.new()
     .addPanel(
       singlestat.new(
         'uptime',
         format='s',
         datasource='Prometheus',
         span=2,
         valueName='current',
       )
       .addTarget(
         prometheus.target(
           'time() - process_start_time_seconds{env="$env", job="$job", instance="$instance", apiVersion="$labelFromEnvs"}',
         )
       )
     )
   )
```

## Providing runtime to build jsonnet dashboards
This feature provides the ability to pass your jsonnet project with all own or external runtime-required libs/dependencies required in runtime to build your dashboard.

It bridges the signature of the jsonnet generation command ```jsonnet -J path/to/libs target.jsonnet``` and Grafana Operator Dashboard CRD.
To do this, there are 3 parameters:
* ```jPath``` - Jsonnet local libs path, must be the same as in your local jsonnet project. Optional part.
* ```fileName``` - Jsonnet file name which must be built. Required part.
* ```gzipJsonnetProject``` - Gzip archived project in a byte array representation. Only .tar.gz files are supported. Required part.


```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: my-favorite-dashboard-with-internal-dependencies
spec:
  resyncPeriod: 30s
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  envs:
    - name: API_VERSION
      value: "1.0.0"
    - name: ENV_FROM_CM # just example, such cm and secrets are not provided by vendor
      valueFrom:
        configMapKeyRef:
          name: custom-grafana-dashboard-cm
          key: GRAFANA_URL
    - name: ENV_FROM_SECRET
      valueFrom:
        secretKeyRef:
          name: custom-grafana-dashboard-secrets
          key: PROMETHEUS_USERNAME
  envFrom: # just example, such cm and secrets are not provided by vendor
    - configMapRef:
        name: custom-grafana-dashboard-cm
    - secretRef:
        name: custom-grafana-dashboard-secrets
  jsonnetLib:
    jPath:
      - "vendor"
    fileName: "overview.jsonnet"
    gzipJsonnetProject: |-
      {{- (.Files.Get "dashboards.tar.gz") | b64enc | nindent 6 }}
```

## ConfigMap

Alternatively a ConfigMap can be referenced which contains the dashboard.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: dashboard-definition
  labels:
    app.kubernetes.io/managed-by: grafana-operator
data:
  json: >
    {
      "id": null,
      "title": "Simple Dashboard from ConfigMap",
      "tags": [],
      "style": "dark",
      "timezone": "browser",
      "editable": true,
      "hideControls": false,
      "graphTooltip": 1,
      "panels": [],
      "time": {
        "from": "now-6h",
        "to": "now"
      },
      "timepicker": {
        "time_options": [],
        "refresh_intervals": []
      },
      "templating": {
        "list": []
      },
      "annotations": {
        "list": []
      },
      "refresh": "5s",
      "schemaVersion": 17,
      "version": 0,
      "links": []
    }
---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafanadashboard-from-configmap
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  configMapRef:
    name: dashboard-definition
    key: json
```

By default the GrafanaDashboard resource is not reconciled if the contents of the referenced ConfigMap changes.
This is due the fact that the default settings for the controller are optimized for performance and ConfigMaps therefore are cached.
If the Dashboard should be reconciled immediately after a ConfigMap change it requires either changing the controller behaviour or labeling the ConfigMap accordingly.

This can be achieved with either option:

* Set the `app.kubernetes.io/managed-by: grafana-operator` label to the ConfigMap as in the example above.
* Disable the controller cache. Set the env variable `ENFORCE_CACHE_LABELS=off` on the controller.
  **Note**: This can have a significant impact on performance depending on the size and numbers of resources in the cluster.
* Use a custom sharding key. Set the env variable `WATCH_LABEL_SELECTORS` to a custom resource selector on the controller.


{{% alert title="Note" color="primary" %}}
In a standard scenario, a folder with default settings gets created through a `GrafanaDashboard` CR. It either matches the Kubernetes namespace a dashboard exist in or `spec.folder` field of the CR.

If you need more control over folders (such as RBAC settings), it can be achieved through a `GrafanaFolder` CR.
{{% /alert %}}

# Deploying a Grafana instance

This document describes how to get up and running with a new Grafana instance on Kubernetes.

## Deploying the operator

The first step is to install the Grafana operator to a namespace in your cluster.

To create a namespace named `grafana` run:

```sh
$ kubectl create namespace grafana
```

Create the custom resource definitions that the operator uses:

```sh
$ kubectl create -f deploy/crds
```

Create the operator roles:

```sh
$ kubectl create -f deploy/roles -n grafana
```

If you want to scan for dashboards in other namespaces you also need the cluster roles:

```sh
$ kubectl create -f deploy/cluster_roles
```

To deploy the operator to that namespace you can use `deploy/operator.yaml`:

```sh
$ kubectl create -f deploy/operator.yaml -n grafana
```

Check the status of the operator pod:

```sh
$ kubectl get pods -n grafana
NAME                                READY     STATUS    RESTARTS   AGE
grafana-operator-78cfcbf8db-ssrgq   1/1       Running   0          17s
```

## Operator flags

The operator accepts a number of flags that can be passed in the `args` section of the container in the deployment:

* *--grafana-image*: overrides the Grafana image, defaults to `quay.io/openshift/origin-grafana`.
* *--grafana-image-tag*: overrides the Grafana tag. See `controller_config.go` for default.
* *--grafana-plugins-init-container-image*: overrides the Grafana Plugins Init Container image, defaults to `quay.io/integreatly/grafana_plugins_init`.
* *--grafana-plugins-init-container-tag*: overrides the Grafana Plugins Init Container tag, defaults to `0.0.2`.
* *--scan-all*: watch for dashboards in all namespaces. This requires the the operator service account to have cluster wide permissions to `get`, `list`, `update` and `watch` dashboards. See `deploy/cluster_roles`.
* *--pod-label-value*: override the value of the `app` label that gets attached to pods and other resources.
* *--namespaces*: watch for dashboards in a list of namespaces. Mutually exclusive with `--scan-all`.

See `deploy/operator.yaml` for an example.

## Deploying Grafana

Create a custom resource of type `Grafana`, or use the one in `deploy/examples/Grafana.yaml`.

The resource accepts the following properties in it's `spec`:

* *dashboardLabelSelector*: A list of either `matchLabels` or `matchExpressions` to filter the dashboards before importing them.
* *containers*: Extra containers to be added to the Grafana deployment. Can be used for example to add auth proxy side cars.
* *secrets*: A list of secrets that are added as volumes to the deployment. Useful in combination with extra `containers` or when extra configuraton files are required.
* *configMaps*: A list of config maps that are added as volumes to the deployment. Useful in combination with extra `containers` or when extra configuraton files are required.
* *config*: The properties used to generate `grafana.ini`. All properties defined in the [official documentation](https://grafana.com/docs/installation/configuration/) are supported although some of them are not allowed to be overridden (path configuration). See `deploy/examples/Grafana.yaml` for an example.  
* *ingress*: Allows configuring the Ingress / Route resource (see [here](#configuring-the-ingress-or-route)).
* *service*: Allows configuring the Service resource (see [here](#configuring-the-service)).
* *initialReplicas*: Allows scaling the number of Grafana pods to the specified replicas.

The other accepted properties are `logLevel`, `adminUser`, `adminPassword`, `basicAuth`, `disableLoginForm`, `disableSignoutMenu` and `anonymous`. They are supported for legacy reasons, but new instances should use the `config` field. If a value is set in `config` then it will override the legacy field. 

*NOTE*: by default no Ingress or Route is created. It can be enabled with `spec.ingress.enabled`.

To create a new Grafana instance in the `grafana` namespace, run:

```sh
$ kubectl create -f deploy/examples/Grafana.yaml -n grafana
```

Get the URL of the instance and open it in a browser:

```sh
$ kubectl get ingress -n grafana
NAME              HOSTS                           ADDRESS   PORTS     AGE
grafana-ingress   grafana.apps.127.0.0.1.nip.io             80        28s
```

## Config reconciliation

When the config object in the `Grafana` CR is modified, then `grafana.ini` will be automatically updated and Grafana will be restarted.

*NOTE*: there is a known issue when removing whole sections from the config object. The operator might not detect the update in such cases. As a workaround we recommend to leave the section header in place and only removing all the sections properties.

## Configuring the Ingress or Route

By default the operator will not create an Ingress or Route. This can be enabled via `spec.ingress` in the `Grafana` CR. 
The operator will create a Route when running on OpenShift, otherwise an Ingress. Various other properties can also be configured:

```yaml
spec:
  ingress:
    enabled:  <Boolean>   # Create an Ingress (or Route if on OpenShift)
    hostname: <String>    # Sets the hostname. Automatically set for Routes on OpenShift.
    labels:               # Additional labels for the Ingress or Route
      app: grafana
      ...
    annotations:          # Additional annotations for the Ingress or Route
      app: grafana
      ...
    path:                 # Sets the path of the Ingress. Ignored for Routes
```

## Configuring the Service

Various properties of the Service can be configured:

```yaml
spec:
  service:
    labels:               # Additional labels for the Service
      app: grafana
      ...
    annotations:          # Additional annotations for the Service
      app: grafana
      ...
    type: NodePort        # Set Service type, either NodePort, ClusterIP or LoadBalancer
```

---
author: "Edvin 'NissesSenap' Norling"
date: 2023-03-14
title: "Version 5 getting started"
linkTitle: "Version 5 getting started"
description: "How to get started with version 5 of the operator?"
---

It might be a good idea to read through the [version 5 introduction](v5-intro.md) which goes through a bit more about the new concepts that we have introduced in this version.
In this blog we will focus on how to install version 5 of the grafana-operator.

For this example we will be using a small [kind](https://kind.sigs.k8s.io/) cluster to get access to Kubernetes but this should of course work with any other Kubernetes installation.

And we will be installing the operator using helm, but you can also use Kustomize and openshift OLM.

At the time of writing this blog we haven't created any migration flow from version 4 to version 5.
Since it contain lots of breaking changes we won't provide any for the grafana instance.
But we might write a script that solves the other resources in the future.

## Prerequisites

- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start)
- [Helm](https://helm.sh/docs/intro/install/) >= v3.8.0
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

By default Kind uses docker to spin up a cluster but any container runtime should work.

### Setup cluster

Create a Kind cluster with [ingress support](https://kind.sigs.k8s.io/docs/user/ingress/).

```shell
cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF
```

When the cluster is up install [ingress-nginx](https://github.com/kubernetes/ingress-nginx).

```shell
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
# Wait for ingress-nginx to become ready
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s
```

When ingress-nginx is up and running we should be ready to install the grafana-operator.

## Install operator

So lets start with hosting the operator in a separate namespace.

```shell
kubectl create ns grafana-operator
```

We are hosting our Helm chart in an OCI repo so it's a bit different from what you might be used to,
notice the `oci://` part of the URL.

```shell
helm upgrade -i grafana-operator oci://ghcr.io/grafana/helm-charts/grafana-operator --version {{<param version>}} -n grafana-operator
```

## Use operator

Easiest way to get started is to look in our [examples](../../examples/) which contains multiple examples on how to configure a grafana instance.

Other then looking at the examples it's also good to use one of the most underrated command in Kubernetes, `explain`.
For example, `kubectl explain grafanadashboard.spec` will give you insights on how you can configure the grafanadashboard.

### Basic example

Lets start with the basic example, this isn't something that you should use in production due to how we define the admin password but it's a simple way of getting started.

```shell
kubectl apply -f https://raw.githubusercontent.com/grafana-operator/grafana-operator/master/examples/basic/resources.yaml
```

Notice the label on the grafana resource, this is the one that GrafanaDashboard will use to find this instance.

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  name: grafana
  labels:
    dashboards: "grafana"
spec:
  config:
    log:
      mode: "console"
    auth:
      disable_login_form: "false"
    security:
      admin_user: root
      admin_password: secret
```

And this is how the GrafanaDashboard, looks like.
Sadly there is no good webhook solution or similar in grafana so we have to continuously poll the grafana API and see if there have been any changes made to the dashboard.
This is the same way we did it in version 4.

> **Note**
> We also need to set the instanceSelector to find the grafana instance that this dashboard should be applied to.

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

For simplicity lets port-forward to the grafana-deployment that we have created.

```shell
kubectl port-forward svc/grafana-service 3000
```

You should now be able to go to localhost:3000 in your browser and login with `username: root` and `password: secret`.

But wait a second didn't we setup a kind cluster with ingress support?
Yes we did, so in the next session lets use it.

### Ingress example

Now lets use the ingress example instead, you can find the [example](../../examples/ingress_http/resources.yaml).
But this time we will do some modifications to it.

Same settings but updating the label and the name to show case that we can run multiple instances of grafana without any issues.
You will need to adapt your hostname to the domain of your picking. Or you can use [nip.io](https://nip.io/) which will steer traffic to your local deployment through a DNS response (e.g. `nslookup grafana.127.0.0.1.nip.io` will respond with `127.0.0.1`).

```shell
kubectl apply -f - <<EOF
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  name: grafana-ingress
  labels:
    dashboards: "grafana-ingress"
spec:
  config:
    log:
      mode: "console"
    auth:
      disable_login_form: "false"
    security:
      admin_user: root
      admin_password: secret
  ingress:
    spec:
      ingressClassName: nginx
      rules:
        - host: grafana.127.0.0.1.nip.io
          http:
            paths:
              - backend:
                  service:
                    name: grafana-ingress-service
                    port:
                      number: 3000
                path: /
                pathType: Prefix
EOF
```

```shell
kubectl apply -f - <<EOF
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafanadashboard-sample-ingress
spec:
  resyncPeriod: 30s
  instanceSelector:
    matchLabels:
      dashboards: "grafana-ingress"
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
EOF
```

You should now be able to reach your domain using http traffic.

### Status

Status messages is something that we have been working hard on to make the grafana-operator easier to use by providing you with information.

For example looking at the grafana-ingress status

```shell
kubectl get grafana grafana-ingress -o yaml | grep status -A 10
```

Should show you something like this

```yaml
status:
  adminUrl: http://grafana-ingress-service.default:3000
  dashboards:
  - default/grafanadashboard-sample-ingress/6eaed1ab-0b0a-4d7f-bc46-0e3c1f58c8a8
  stage: complete
  stageStatus: success
```

We have `adminUrl` that is used by the operator internally.
But the more important part is that we can see that `stage: complete` and we can also see which dashboard that is applied to your grafana instance.

### Cleanup

```shell
# Delete the basic grafana example
kubectl delete -f https://raw.githubusercontent.com/grafana-operator/grafana-operator/master/examples/basic/resources.yaml
# Delete the ingress example
kubectl delete grafanadashboards grafanadashboard-sample-ingress
kubectl delete grafana grafana-ingress
# Uninstall grafana-operator
helm uninstall grafana-operator . -n grafana-operator
# Remove kind
kind delete cluster
```

## Helm CRD Caveats

If you are using helm to install the operator and in the future want to upgrade the operator, make sure that the CRDs get upgraded manually.

This is due to how [Helm](https://helm.sh/docs/chart_best_practices/custom_resource_definitions/#method-1-let-helm-do-it-for-you) works. But it's important to remember this since we most likely will do lots of CRD changes in the near future and we want you to follow along with the CRD changes.

## Lessons learned

So hopefully you have learned how to setup grafana-operator version 5 using helm.
Learned about the new concepts that we have introduced with the operator and hopefully enjoying the new functionality.

If you find any issues feel free to create one after reading through the existing once in v5 [labels](https://github.com/grafana/grafana-operator/labels/v5) to see all open issues.

To give feedback you can also join us in the [Kubernetes Slack](https://slack.k8s.io/) in the [grafana-operator channel](https://kubernetes.slack.com/messages/grafana-operator/).

And of course we are happy to receive PRs.

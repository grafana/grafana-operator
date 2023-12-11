---
author: "Edvin 'NissesSenap' Norling"
date: 2023-03-14
title: "Version 5 RC introduction"
linkTitle: "Version 5 RC introduction"
description: "Grafana-operator version 5 RC is now out, lets find out what has changed."
---

If you are reading this it means that we have finally got our first RC of version 5 out of the gates.
The maintainers are extremely happy to finally being able to do so.

As a part of this we have merged version 5 in to the master branch and we will continue to support version 4 through the v4 branch. Any changes that needs to be done towards version 4 should be directed to that branch.

In this blog we will try to explain the changes that we have made and the future of the operator.
In [v5 getting started](v5-getting-started.md), you will be able to read how to try out the operator.

Version 5 or v5 as I will write from now on is a complete rewrite of version 4 of the operator.
The API versions are different so if you want you can run both the operators at the same time while you do a migration.

There are multiple reasons why we did this, but the main reason is the number of architectural changes that we have done.

Just a reminder this is an RC and not production ready, we will probably do a few breaking changes before releasing v5 but we really want to bring you something to try out and get some feedback on what we have done.

## Architectural changes

The main change is that instead of the grafana Custom Resource(CR) having an instance selector looking for a specific label in the dashboards CR and data sources CR it's the other way around and the grafanadashboard CR now selects the grafana instance.

This way we can have the controller managing the grafana client which makes the code easier to maintain.

In this basic example you will see the grafana instance with a label.

```.yaml
---
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

And the grafana dashboard with the instanceSelector, the same thing goes for the other grafana resources.

```.yaml
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

## New Features

This is probably the part of the blog that you are most excited about, it goes through a number of the new features that we support.

### Multi grafana instances support

This is one of our most requested and [oldest issues](https://github.com/grafana/grafana-operator/issues/174) in the grafana-operator and we have finally been able to provide it.

In v5 you can now run multiple instances of grafana in the same namespace and externally.

### External grafana instances

It's now possible to configure external grafana instances through the operator.
So if you are using grafana cloud or a hosted grafana instances in one of the big cloud providers you can still use gitops to apply your dashboards to them.

One of the limitations with the external grafana instances is that we don't support installing plugins.
If your dashboard is using a plugin we assume that you have manually installed it to your grafana instance.

### Grafana config

The deployment configuration is more **opinionated** in it's default setup while at the same time giving you the user access to the whole deployment spec so you can overwrite anything you want.
For example when you start a grafana instance it will by default have decent security defaults, it will set `RunAsNonRoot` on your grafana instance when they get created.
But since you have access to the whole deployment spec you can always overwrite these kind of settings.

The same thing applies to all our resources, for example ingress.
The bad thing here is of course that we won't hide any of the logic that made the operator "easy" to get started with.
So if you want to configure an ingress for your grafana instance you will have to define it just like any other ingress resources.
But it makes it extremely extensible and hopefully this will make the operator support any setup you want.

In the example below you can see how we define a ingress and overwrite settings in the deployment.

```.yaml
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
  deployment:
    spec:
      template:
        spec:
          containers:
            - name: grafana
              securityContext:
                readOnlyRootFilesystem: false
  ingress:
    spec:
      ingressClassName: nginx
      rules:
        - host: example.com
          http:
            paths:
              - backend:
                  service:
                    name: grafana-service
                    port:
                      number: 3000
                path: /
                pathType: Prefix
      tls:
        - hosts:
            - example.com
          secretName: core-cert
```

### Grafana ini

Just like deployment config the same thing goes for grafana config, instead of having a CRD that contains all the grafana ini values we have instead made it possible to write any value as you want.
The **bad** thing is of course that you as an end user don't get any help with verifying types etc.
The **good** thing is that you can set any value you want, if you want to try out an experimental feature you can do so without first having to create a PR to the operator.

### Usage of status fields

The usability of v4 have more then once been laking. It's been really hard to understand why something is happening and you more or less have been forces to go in and look at the logs.
This is something that we wanted to change with v5.

In v5 you will be able to look at your grafana dashboard CR and see that it got a matching grafana instance (the label selector is correct) and if it's been applied to the grafana instances.

### No more plugin sidecar

As you might know we have had a python container that have been an [sidecar](https://github.com/grafana/grafana_plugins_init) to the operator that have been in charge of installing the plugins to grafana.
The sidecar have been badly maintained and to be honest it's just not needed. Instead we use an environment variable to install plugins during startup.
So when version 4 is out of support we will archive the grafana plugins init repo.

### Helm chart

We have decided to start to manage our own Helm chart. The main reason for this is that we want a simple way of giving you test versions of the operator and many of you use Helm to install the operator.

We still see a big place for the [community Helm chart](https://bitnami.com/stack/grafana-operator/helm) that bitnami is hosting and we hope that they will keep on doing so.

### Namespace mode

Another issue that we have gotten lots of feedback on is [Excessive rbac access](https://github.com/grafana/grafana-operator/issues/604)
This wasn't something that we could easily fix in v4 but in v5 we have really tried to accommodate this by introducing namespace mode.

In namespace mode you can run the operator to only look for grafana resources in the same namespace that you are in or you can define it to watch a number of specific namespaces. By default it will use cluster wide access.

If you install through Helm or Kustomize you will be able to define if you should create cluster roles or only roles.
Sadly for our Openshift users this won't be possible due to the limitations of OLM, even if you run the operator in namespace mode, which you can. OLM will still use the cluster wide rbac settings.

## v5

So there is still lots to do before version 5 will be released.

There are still a number of features that is missing and we have tried to create issues around this.
Please have a look in our v5 [milestone](https://github.com/grafana/grafana-operator/milestone/3) or the v5 [labels](https://github.com/grafana/grafana-operator/labels/v5) see all open issues.

But we need **your** help, first of all try out version 5.

- Do you like it?
- Is everything working as you expected?
- Is there any features missing?
- Is there an open issue for that feature? Please search for it and if it's not please create an issue.
- Help out with PR:s, to get started look at the [CONTRIBUTING.md](CONTRIBUTING.md) file.

We probably won't support all the features that we did in v4, due to maintenance burden, but if a missing feature get lots of up votes we might give it a second review.

An extra note about the **do you like it?**, getting feedback is one of the hardest things in open-source.
Even if "all" you did was to follow the new v5 [installation guide](v5-getting-started.md) we want to know about it, you might feel that you don't have specific insights or thoughts and that is fine.
But writing one row on slack and saying "hey I installed version v5 and it worked as I expected". Helps us a lot since we get to know that people are trying it.

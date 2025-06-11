---
title: "OpenShift example"
linkTitle: "OpenShift example"
---

A basic deployment that makes use of OpenShift routes.

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

By default Routes are used on OpenShift, but configuring `.spec.ingress` and leaving `.spec.route` empty signals to the operator to use an Ingress instead.

{{< readfile file="openshift_ingress.yaml" code="true" lang="yaml" >}}

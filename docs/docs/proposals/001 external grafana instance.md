---
title: "External Grafana Instance Integration"
linkTitle: "External Grafana Instance Integration"
---

## Summary

Introduce integration to External Grafana Instances with Grafana Operator.

This document contains the complete design required for integrating external Grafana instances with Grafana Operator.
This includes design elements to support the following with Grafana operator :

- Ability to define an external Grafana instance as Grafana source to Grafana operator
- Ability to create Grafana dashboards on a remote Grafana Instance.
- Ability to add external cloud data sources to an external Grafana instance.

## Info

status: Implemented

## Motivation

Cloud providers have started providing managed remote grafana services which decouples the responsibilities of managing a grafana instanaces from ops personas and kubernetes environment.

Currently Grafana operator has an integration to add external data sources as a data source to Grafana instances hosted in a kubernetes environment. As more customers starting to use external grafana services, expanding the Grafana operator to support remote Grafana instances becomes inevitable. Adding ability to integrate with external grafana services, adding data sources, creating dashboards and alerting on a remote Grafana instances offloads responsibilities of managing a grafana instanaces from ops personas which helps them to focus on developing the features required for their business. This helps the customer teams to move from self managed Grafana instance on their Kubernetes environments to Pay as you go model on Grafana instances provided by providers.

## Verification

- Create integration tests for adding keys for remote Grafana instance.
- Create integration tests to create Grafana dashboards on a remote Grafana Instance.
- Create integration tests to add cloud data sources to remote Grafana instance.

## Current

Currently the grafana operator supports the following for only self managed Grafana Instance :
- Adding remote remote cloud data sources.
- Creating Dashboards.
- Setting up alerting.
- And Many more.

## Proposal

In short the proposal in this document is about enhancing the Grafana Operator to support the integration to managed grafana services. We would need to enhance the current version of Grafana Operator to support the following :

- Defining external Grafana instance as Grafana source to Grafana operator.

### Defining external Grafana instance as Grafana source to Grafana operator.

Today Grafana operator supports self managed grafana instance as a Grafana source to Grafana operator. `Grafana` CRD of the grafana operator should be enhanced to integrate with external grafana instance as shown below in three options.

> Option1 : CRD `Grafana` With `grafana_api_key`. In this design, `grafana_api_key` should be loaded as a Secret. Choice of using external secrets or loading secrets manually is end user responsibility :

```.yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  annotations:
  labels:
    app: grafana
  name: grafana
  namespace: grafana-operator-system
spec:
  external:
    url: <external grafana url, type string>
    grafana_api_key: <type SecretKeySelector>
```

> Option 2 : CRD `Grafana` with `admin_username`, `admin_password`. In this design, `admin_username`, `admin_password` should be loaded as a Secret. Choice of using external secrets or loading secrets manually is end user responsibility :

```.yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  annotations:
  labels:
    app: grafana
  name: grafana
  namespace: grafana-operator-system
spec:
  external:
    url: <external grafana url, type string>
    admin_username: <type SecretKeySelector>
    admin_password: <type SecretKeySelector>
```

> Option 3 : CRD `Grafana` with both `grafana_api_key` and`admin_username`, `admin_password`. In this design, `grafana_api_key`, `admin_username`, `admin_password` should be loaded as a Secret. In this case `grafana_api_key` takes higher precedence over `admin_username`, `admin_password`. Choice of using external secrets or loading secrets manually is end user responsibility :

```.yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  annotations:
  labels:
    app: grafana
  name: grafana
  namespace: grafana-operator-system
spec:
  external:
    url: <external grafana url, type string>
    grafana_api_key: <type SecretKeySelector>
    admin_username: <type SecretKeySelector>
    admin_password: <type SecretKeySelector>
```

Note: Adding `examples` on using accessing external `url`, `grafana_api_key` would really help the users to get up to speed to use external Grafana instance feature.

## Related issues

- [Issue 402](https://github.com/grafana/grafana-operator/issues/402)
---
title: "GrafanaServiceAccount CRD"
linkTitle: "GrafanaServiceAccount CRD"
---

## Summary

Add a GrafanaServiceAccount CRD so the operator can create Grafana Service Accounts using CRs to automate the setup of Grafana Service Accounts in newly setup grafana instances. 

Today its required to manually set them up in a running grafana instance using the Grafana GUI or the HTTP-API. This document introduces the suggestion of having them as separate objects that can be setup by the operator on deploy.

The suggested new features are:
* Let the operator create a grafana service account using a GrafanaServiceAccount CRD.
* Let the operator store the token as a k8s-secret

## Info

Status: Suggested

## Motivation

Today Grafana Service Accounts has to be created after deploy when the grafana is running using the HTTP-API or the GUI. I instead suggest to have a Grafana Service Account CRD so that the Service Accounts could be predefined and created by the operator at deploy and the tokens will be created as k8s-secrets that then can be read by applications when needed.

## Verification

- Create integration tests for the operator creating grafana service accounts from a bare minimum yaml
- Create integration tests for the operator creating grafana service accounts from a fully specified yaml
- Create integration tests to check that it can rotate/invalidate tokens with TTL set (and passed).

## Current solution

Currently you are only able to create these grafana service accounts using the grafana GUI or by using the HTTP-API after the grafana has already been deployed and is running. And its removed when the grafana pod is restarted/redeployed without persistent storage. Meaning a new service account has to be manually created and its new token has to be updated where its being used.

## Proposal

My proposal is to handle grafana service account as a separate CRD that can be specified by the user even before setup and that can be included in a CICD pipeline. It will enable so that the operator can create predefined service accounts on deploy and store the token in a k8s-secret readable by other applications without any manual steps.

### Defining what Grafana Service Account belongs to what grafana.
I suggest that its done in the same way as dashboards and datasources are using instanceSelector to find the proper grafana, the same thing could be done with the grafana service accounts using the same labels etc.

### Defining Grafana Service Account to Grafana operator.

Today the Grafana Service Account is only held in memory if not using persistent, so when a pod is restarted or redeployed the grafana service account is removed. The grafana service accounts are also only possible to create using the Grafana GUI or with the HTTP-API and cannot be pre-defined before deploy or kept as IAC.

> Proposed CRD for GrafanaServiceAccounts

```.yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaServiceAccount
metadata:
  labels:
    app: grafana
  name: grafana-sa
  namespace: grafana-namespace
spec:
  create:
    generateTokenSecret: [true/false] #Will create the k8s-secret with a default name if true. Defaults to true.
  serviceaccount:
    id: grafana-sa
    roles: [Viewer/Editor/Admin]
  tokens:
    - Name: grafana-sa-token-<name-of-GSA>
      expires: <Absolute date for expiration, defaults to Never>
  permissions:
    - user: <users in the cluster/root user etc>
      permission: [Edit/Admin]
  instanceSelector:
    matchLabels:
      some.selector.label-A: "grafana-A"
      some.selector.label-B: "grafana-B"
```

My suggestions is that "Last used" value for the token would be kept in memory and wiped at restart/redeploy just like it would be today when the SA is removed completely together with any "last used" information. This is in order to avoid having additional requirements on storage, either being persistent storage in the cluster or an external db.

As for the creation time for both the service account and the token should be possible to fetch from the objects themselves. Kubernetes objects have a creationTimestamp in their metadata.

### The handling of TTL of tokens
This suggestion would still allow for the operator to handle TTL by just replacing the secret with a new token that then can be picked up by applications in the cluster.
    


## Related issues

- [Issue 1388](https://github.com/grafana/grafana-operator/issues/1388)

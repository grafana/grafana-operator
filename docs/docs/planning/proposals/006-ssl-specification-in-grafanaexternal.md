---
title: "Add TLS management block in Grafana CR External block"
linkTitle: "Add TLS management block in Grafana CR External block"
---
{{% pageinfo color="info" %}}
TLS settings are top level options (`spec.client.tls`) as of [#1690](https://github.com/grafana/grafana-operator/pull/1690) and apply to _all_ Grafana instances.
Other than this change, the original proposal still holds
{{% /pageinfo %}}

## Summary

Introduce the possibility to give a tls specification to the Grafana CR's external block.

This document contains the complete design required to support configuring tls by extending the Grafana CRD.

The suggested new feature is:
- Permits to declare a tls block which will give the possibility to use a certificate to connect a Grafana instance.

## Info

status: Implemented

## Motivation

Currently, the operator does not permits to connect with a Grafana with a not thrusted certificate without rebuilding the full container.

## Proposal

This document proposes to extend the Grafana CRD external block to add a block with tls information. In this block, we will find:
- `certSecretRef`: This block will contains the name of a secret which will contained certificates based on the `kubernetes.io/tls` format (e.g. `ca.crt`, `tls.crt` and `tls.key`). This secret can contains only `ca.crt` or `tls.crt` and `tls.key` at the same time. Both solution are not mutually exclusive.
- `insecureSkipVerify`: Disable the server certificate check (facultative - default: false)

The tls block should be facultative. However, if the tls block is set, at least of it subfield should be present.

Doing this, the Grafana CRD will evolve to look into something like this:
```yaml
---
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  name: external-grafana
  labels:
    dashboards: "external-grafana"
spec:
  external:
    url: https://test.io
    adminPassword:
      name: grafana-admin-credentials
      key: GF_SECURITY_ADMIN_PASSWORD
    adminUser:
      name: grafana-admin-credentials
      key: GF_SECURITY_ADMIN_USER
    tls:
      certSecretRef:
        name: tls-certificate
      insecureSkipVerify: false
```

## Impact on the already existing CRD

Because this block is an addition to the existing Grafana CRD, the already deployed Grafana CR will not be impacted.
However, because this functionality touch to the Grafana client, we need to be sure the evolution does not introduce regression in the product.

## Decision Outcome

We're going to implement CA verification simmilar to [the way flux does it](https://fluxcd.io/flux/components/source/helmrepositories/#cert-secret-reference) to keep in line with the rest of the Kubernetes ecosystem

## Related discussions

- [PR 1590](https://github.com/grafana/grafana-operator/pull/1590)
- [PR 1594](https://github.com/grafana/grafana-operator/pull/1594)

---
title: "Add TLS management block in Grafana CR External block"
linkTitle: "Add TLS management block in Grafana CR External block"
---

## Summary

Introduce the possibility to give a tls specification to the Grafana CR's external block.

This document contains the complete design required to support configuring tls by extending the Grafana CRD.

The suggested new feature is:
- Permits to declare a tls block which will give the possibility to use a certificate to connect a Grafana instance.

## Info

status: Suggested

## Motivation

Currently, the operator does not permits to connect with a Grafana with a not thrusted certificate without rebuilding the full container.

## Proposal

This document proposes to extend the Grafana CRD external block to add a block with tls information. In this block, we will find:
- `caBundle`: This block will contain the name of the secret where the CA bundle is currently stored. To do this, we could use the `*v1.SecretKeySelector` mechanism already used in the Grafana CRD. (facultative - default: empty)
- `insecureSkipVerify`: Disable the server certificate check (facultative - default: false)
- `cert`: A certificate that can be used to authenticate the request to the response server (mandatory when `key` is defined - default: empty)
- `key`: The key associated with the certificate declared in `cert_pem` field (mandatory when `cert` is defined - default: empty)

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
      caBundle:
        name: tls-certificate
        key: ca-bundle.pem
      insecureSkipVerify: false
      cert:
        name: tls-certificate
        key: cert.crt
      key:
        name: tls-certificate
        key: key.key
```

## Impact on the already existing CRD

Because this block is an addition to the existing Grafana CRD, the already deployed Grafana CR will not be impacted.
However, because this functionality touch to the Grafana client, we need to be sure the evolution does not introduce regression in the product.

## Decision Outcome

<!-- TODO: to be discussed with maintainer -->

## Related discussions

- [PR 1590](https://github.com/grafana/grafana-operator/pull/1590)

---
title: Security
weight: 50
---

## Verification of container images

Grafana-operator container images are signed by cosign using identity-based ("keyless") signing and transparency. Executing the following command can be used to verify the signature of a container image:

To verify the grafana-operator run

Pre-requirement

- cosign v2.0.0 or higher [installation instructions](https://docs.sigstore.dev/system_config/installation/).

```shell
cosign verify ghcr.io/grafana/grafana-operator@<version> \
  --certificate-identity-regexp 'https://github\.com/grafana/grafana-operator/\.github/workflows/.+' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com | jq
```

For example

```shell
cosign verify ghcr.io/grafana/grafana-operator@v5.6.1 \
  --certificate-identity-regexp 'https://github\.com/grafana/grafana-operator/\.github/workflows/.+' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com | jq
```

## SBOM

As a part of our release cycle we also generate SBOMs.
You can find them as artifacts in our supported repositories.

To download the sbom you can run

```shell
cosign download sbom --platform linux/amd64 ghcr.io/grafana/grafana-operator:<version>
```

example:

```shell
cosign download sbom --platform linux/amd64 ghcr.io/grafana/grafana-operator:v5.6.1
```

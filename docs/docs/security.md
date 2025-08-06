---
title: Security
weight: 50
---

## Verification of container images

Grafana-operator container images are signed by github [attestation](https://docs.github.com/en/actions/how-tos/security-for-github-actions/using-artifact-attestations/using-artifact-attestations-to-establish-provenance-for-builds). Executing the following command can be used to verify the signature of a container image:

> This applies to grafana-operator version 5.19.1 and forward

### Prerequisites

- cosign v2.5.0 or higher [installation instructions](https://docs.sigstore.dev/cosign/system_config/installation/).
- gh 2.72.0 or higher [cli](https://github.com/cli/cli/releases)
- crane [installation instructions](https://github.com/google/go-containerregistry/blob/main/cmd/crane/doc/crane.md) or oras [installation instructions](https://oras.land/docs/installation)

### Get container digest

```shell
# Get the image digest using crane
crane digest --platform linux/amd64 ghcr.io/grafana/grafana-operator:v5.19.1

# Or using oras
oras resolve --platform linux/amd64 ghcr.io/grafana/grafana-operator:v5.19.1
```

### Verify the grafana-operator image

```shell
gh attestation verify --owner grafana oci://ghcr.io/grafana/grafana-operator@<sha256>
```

For example

```shell
gh attestation verify --owner grafana oci://ghcr.io/grafana/grafana-operator@$(oras resolve --platform linux/amd64 ghcr.io/grafana/grafana-operator:v5.19.1)
```

Or if you prefer, you can use cosign.

```shell
cosign verify-attestation --certificate-identity-regexp 'https://github\.com/grafana/grafana-operator/\.github/workflows/.+'  --certificate-oidc-issuer https://token.actions.githubusercontent.com --new-bundle-format  --type=slsaprovenance1 ghcr.io/grafana/grafana-operator:@$(oras resolve --platform linux/amd64 ghcr.io/grafana/grafana-operator:v5.19.1) | jq -r '.payload | @base64d | fromjson'
```

### Verify SBOM

As a part of our release cycle we also generate SBOMs.
You can find them as artifacts in our github repositorie or in the public cosign instance.

> Notice the platform specification in the commands.
> This is needed since the sbom is matching to the platform specific container image.

```shell
# Download the SBOM attestation using the digest (example with oras)
cosign download attestation --predicate-type https://spdx.dev/Document \
  ghcr.io/grafana/grafana-operator@$(oras resolve --platform linux/amd64 ghcr.io/grafana/grafana-operator:v5.19.0)
```

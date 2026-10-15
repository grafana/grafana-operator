---
title: Security
weight: 50
---

## Hardening the Operator

When we develop features for the operator, we assume a certain level of trust granted to users who will be able to create CRDs.
This might not work in every situation though so this section documents common pitfalls, security implications you might not expect and how to harden the operator to allow secure deployments in every situation.

### Per resource implications
The following parts outline capabilities of our custom resources.
You should only grant access to these resources to users that already have the same powers in a specific namespace.

#### `Grafana`
The `Grafana` resources is a very powerful resource.
When used to deploy a managed instance, it can override almost any field in the deployment and as such, run arbitrary workloads in the cluster.
This includes mounting arbitrary secrets, configmaps and volumes from the same namespace.

#### `GrafanaDatasource`, `GrafanaDashboards`, `GrafanaContactPoints`, `GrafanaManifest`, `GrafanaLibraryPanel`, `GrafanaContactPoint`
These resources all allow users to modify them based on `ConfigMaps` and `Secrets` in the same namespace.
Consequently, this means that if you have a `Grafana` instance in the same namespace as a `GrafanaDashboard`, a malicious user with edit permissions for `GrafanaDashboards` can use this power to extract the credential secret of the instance. See the hardening section on various approaches to prevent this if required.

Some resources (Data sources, Dashboards, Library Panels) also allow installing plugins in the targeted Grafana instance.

#### `GrafanaFolder`, `GrafanaNotificationPolicy`, `GrafanaNotificationPolicyRoute`, `GrafanaNotificationTemplate`, `GrafanaSilence`, `GrafanaMuteTiming`
Right now, these resources do not allow customization through secrets or config maps. However, with the [upcoming patching support](./planning/proposals/009-dynamic-resource-patching.md) this will change and they'll also have the capability to reference external data.

#### `GrafanaServiceAccount`
The `GrafanaServiceAccount` resource is a bit special. It dynamically creates secrets for configured tokens which might cause existing secrets to be overwritten.

### Hardening

If you require a hardened setup and want to allow low-trust users to manage some of the before mentioned resources, here are some patterns you can use to enable this.

#### Isolate resources in namespaces

This mitigates leaking of the instance secret. Create the `Grafana` instance in one namespace (e.g. `grafana-instance`) and enable `spec.allowCrossNamespaceImport`.
You can then create a new empty namespace (e.g. `grafana-resources`) which does not contain any sensitive secrets and have users create their resources in this namespace.

#### ValidatingAdmissionPolicy

To restrict usage of certain fields, you can make use of [Validating Admission Policies](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/).
These allow you to validate resources before they reach the API server.

For example, you can prevent creation of resources that reference secrets with the following policy:

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicy
metadata:
  name: "prevent-secret-referencing"
spec:
  failurePolicy: Fail
  matchConstraints:
    resourceRules:
      - apiGroups: ["grafana.integreatly.org"]
        apiVersions: ["v1beta1"]
        operations: ["CREATE", "UPDATE"]
        resources: ["grafanadashboards", "grafanalibrarypanels"]
  validations:
    - expression: |
        (
          !has(object.spec.envs) ||
          object.spec.envs.all(v,
            !has(v.valueFrom) || !has(v.valueFrom.secretKeyRef)
          )
        )
        &&
        (
          !has(object.spec.envFrom) ||
          object.spec.envFrom.all(v,!has(v.secretKeyRef))
        )
        &&
        (
          !has(object.spec.patch) ||
          !has(object.spec.patch.env) ||
          object.spec.patch.env.all(v,!has(v.valueFrom) || !has(v.valueFrom.secretKeyRef))
        )
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicyBinding
metadata:
  name: "prevent-secret-referencing-clusterwide"
spec:
  policyName: "prevent-secret-referencing"
  validationActions: [Deny]
```

You can also use this to limit modification of the deployment created by the `Grafana` resource.

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
  ghcr.io/grafana/grafana-operator@$(oras resolve --platform linux/amd64 ghcr.io/grafana/grafana-operator:v5.19.1)
```

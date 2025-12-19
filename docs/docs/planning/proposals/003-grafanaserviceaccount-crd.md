---
title: "GrafanaServiceAccount CRD"
linkTitle: "GrafanaServiceAccount CRD"
---

## Summary

Add GrafanaServiceAccounts to the Grafana CRD so the operator can create Grafana Service Accounts automatically when deploying grafana instances.

This proposal outlines a new custom resource called `GrafanaServiceAccount` that manages the service account, it's role and associated tokens.

## Info

status: Implemented

## Motivation

The Grafana operator does not support management of service accounts in a declarative way.

We want to cover the following use cases:

* As an administrator of a Grafana instance, I want to create a service account for it
* As a developer requiring a Grafana service account, I want to create a service account on demand per application
* As a security conscious SRE, I want to ensure nobody can compromise a Grafana instance through the Grafana operator


## Verification

- The operator can create new Grafana service accounts
- The operator can rotate tokens when the expiration date changes
- The operator overrides manually set tokens

## Current solution

Currently you are only able to create these grafana service accounts using the grafana GUI or by using the HTTP-API after the grafana has already been deployed and is running.
When not using persistent storage, this service account is removed on reconciliation so there is no way to declaratively manage service accounts as code, using the operator.

## Proposal

To support this functionality, we propose the following changes to the Grafana operator.

### Create a new resource `GrafanaServiceAccount`

This resource controls the reconciliation of service accounts. An example could look like this:

```.yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaServiceAccount
metadata:
  name: grafana-sa
spec:
  instanceName: test-grafana-instance
  name: grafana-service-account
  role: Viewer # Valid options: Viewer, Editor, Admin
  tokens:
    - name: test-token
      expires: 2025-12-31T14:00:00+02:00 # optional / never expires if unset
      secretName: grafana-sa-token # optional / generated if unset
  permissions: # this controls who is allowed to customize the service account
    - user: <users in the cluster/root user etc>
      permission: [Edit/Admin]

```

Since reconciling lists is a complex operation to implement, both the permissions & tokens lists are seen as authoritative.
This means that, if defined, these lists are the full set of specified values and any customizations made through the Grafana UI are replaced/removed on reconciliation.

Service accounts reference an instance by resource name directly to ensure correct targeting and avoid accidentally creating accounts on instances which should not be targeted.
For now, service accounts can only exist in the same namespace as the Grafana resource as a security precaution.



### The handling of TTL of tokens

Grafana supports setting expiration of tokens. The operator should respect this and not automatically extend the TTL.
When the user updates the `expires` field of a token, the operator deletes the token and creates a new token under the same name with the updated expiration date.
This effectively rotates the secret.

### Security considerations

As service accounts are a sensitive topic when it comes to security and auditing, special attention is taken here to reflect on security implications of this resource.

Pointed out by @nissessenap in [the original proposal discussions](https://github.com/grafana/grafana-operator/pull/1413#issuecomment-1962404070), users need a way to restrict who can create service accounts for a specific Grafana instance.

By having a dedicated resource, the permission to create service accounts can be granted through standard kubernetes RBAC on a namespace level.
This works to ensure kubernetes users can only create Grafana service accounts when explicitly granted access to do so in a specific namespace.
Granting cluster-wide permissions to create service accounts is not advisable.
For now, namespaces are the finest granularity on which we grant access control.
This means, it is not possible to have multiple Grafana instances in one namespace with different access rules.
Future implementations could support creation of service accounts through the Grafana resource itself, solving for this situation as well.

### Scopes & Limitations

The service account API doesn't support UIDs so we'll have to do some kind of matching between existing resources and the desired state.
To resolve conflicts, we _always_ apply what's defined in the operator resources.
If an administrator manually creates or modifies service accounts, these changes will be overwritten by the operator.
This significantly reduces the implementation complexity as it avoids the need to store state in the `status` field.

Another limitation regards the creation of secrets.
For now, `GrafanaServiceAccount` resources only support creating secrets in the same namespace.
This ensures that we don't compromise the integrity of secrets in other namespaces to which the creator of the service account might not have access to.


## Related issues

- [Issue 1388](https://github.com/grafana/grafana-operator/issues/1388)
- [PR 1907](https://github.com/grafana/grafana-operator/pull/1907)
- [PR 2055](https://github.com/grafana/grafana-operator/pull/2055)

## Additional context

@ndk started implementing the original proposal which sparked a lot of discussions around the proposal and whether it makes sense to implement it as is.
We discussed different controller strategies, placement of resources and implementation complexity.
As an outcome, this proposal has been updated to reflect many, many sessions of discussing this topic so it can serve as a reference for implementing this functionality.

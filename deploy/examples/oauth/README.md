# Grafana OauthProxy Example

This example shows how to combine Grafana with the [OpenShift OAuthProxy](https://github.com/openshift/oauth-proxy). This will only work on OpenShift.

## Installation

1. Create the [session secret](./session-secret.yaml) in the same namespace as Grafana.
2. Create the additional [cluster role](./cluster_role.yaml) and [binding](./cluster_role_binding.yaml).
3. Create the [config map](./ocp-injected-certs.yaml) in the same namespace as Grafana.
4. Create Grafana from the [CR](./Grafana.yaml) in this example.

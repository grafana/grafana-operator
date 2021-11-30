# Grafana OauthProxy Example

This example shows how to combine Grafana with the [OpenShift OAuthProxy](https://github.com/openshift/oauth-proxy). This will only work on OpenShift.

## Installation
1. Create your desired namespace, install grafana operator there and change `example-namespace` in all yaml files to your namespace.
2. Create the [service account](./service_account.yaml) in the same namespace as Grafana.
3. Create the [session secret](./session-secret.yaml) in the same namespace as Grafana.
4. Create the additional [cluster role](./cluster_role.yaml) and [binding](./cluster_role_binding.yaml).
5. Create the [config map](./ocp-injected-certs.yaml) in the same namespace as Grafana.
6. Create Grafana from the [CR](./Grafana.yaml) in this example.
7. Create the [route](./grafana_route.yaml) in the same namespace as Grafana.

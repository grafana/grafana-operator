---
title: "Keycloak OAuth2 SSO"
linkTitle: "Auth: Keycloak OAuth2 SSO"
tags:
  - Auth config
---

A basic example of a Grafana Deployment that overrides generic oauth configuration, it's important to note that most configuration that is valid in the `grafana` container can be done with grafana-operator.

## Steps

### Create Keycloak Client for Grafana

Follow official Grafana guide in how to create a Keycloak client and role mappers for Grafana [here](https://grafana.com/docs/grafana/latest/setup-grafana/configure-security/configure-authentication/keycloak/#keycloak-configuration).

### Create a Kubernetes Secret

In order to safely store the client-id and client-secret for the Keycloak client you have created in the first step, we recommend you creating a Kubernetes secrets to store the client-id and client-secret that Keycloak will use.

The grafana-operator is agnostic to any secret management solution you might use to get this secret (Vault, external-secrets, vanilla K8s secrets, etc).

```yaml
apiVersion: v1
data:
  client-id: c29tZXJlYWxseWxvbmdzZWNyZXRqdXN0dG9jb3ZlcnN0dWZmCg==
  client-secret: c29tZXJlYWxseWxvbmdzZWNyZXRqdXN0dG9jb3ZlcnN0dWZmCg==
kind: Secret
metadata:
  name: grafana-oauth
type: Opaque
```

### Creating our Grafana Instance

Create a Grafana instance overriding the configuration for `auth.generic_oauth:`.

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  name: grafana
  labels:
    dashboards: "grafana"
spec:
  config:
    log:
      mode: "console"
    auth:
      disable_login_form: "false"
    auth.generic_oauth:
      # For variables see https://grafana.com/docs/grafana/latest/setup-grafana/configure-grafana/#env-provider
      enabled: "true"
      name: "Keycloak SSO"
      allow_sign_up: "true"
      client_id: ${AUTH_CLIENT_ID}
      client_secret: ${AUTH_CLIENT_SECRET}
      scopes: "openid email profile offline_access roles"
      email_attribute_path: email
      login_attribute_path: username
      name_attribute_path: full_name
      groups_attribute_path: groups
      auth_url: "https://<PROVIDER_DOMAIN>/realms/<REALM_NAME>/protocol/openid-connect/auth"
      token_url: "https://<PROVIDER_DOMAIN>/realms/<REALM_NAME>/protocol/openid-connect/token"
      api_url: "https://<PROVIDER_DOMAIN>/realms/<REALM_NAME>/protocol/openid-connect/userinfo"
      role_attribute_path: "contains(roles[*], 'admin') && 'Admin' || contains(roles[*], 'editor') && 'Editor' || 'Viewer'"
...
```
Now, we need to set `secretKeyRef` to the Grafana container to pass the values inside the secret you have created in the previous step as environment variables. Please make sure to point to the right secret name and key.

```yaml
...
  deployment:
    spec:
      template:
        spec:
          containers:
            - name: grafana
              env:
                - name: AUTH_CLIENT_ID
                  valueFrom:
                    secretKeyRef:
                      name: grafana-oauth
                      key: client-id
                - name: AUTH_CLIENT_SECRET
                  valueFrom:
                    secretKeyRef:
                      name: grafana-oauth
                      key: client-secret
...
```
Full configuration is below.

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

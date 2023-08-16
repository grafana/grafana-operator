---
title: "Grafana deployment with Google SSO configuration"
linkTitle: "Grafana deployment with Google SSO configuration"
---

A basic example of a Grafana Deployment that overrides SSO configuration, it's important to note that most configuration that is valid in the `grafana` container can be done with grafana-operator.

Make sure that you have DNS and HTTPS already configured with your Grafana instance, as Google SSO requires HTTPS to work with SSO applications.

## Steps

### Generate Google OAuth Keys

Follow official Grafana guide in how to create Google Oauth Keys [here](https://grafana.com/docs/grafana/latest/setup-grafana/configure-security/configure-authentication/google/).

### Create a Kubernetes Secret

In order to safely manage the OAuth keys/credentials we recommend you creating a Kubernetes secrets to store the clientId and clientSecret that Google will use.

The grafana-operator is agnostic to any secret management solution you might use to get this secret (Vault, external-secrets, vanilla K8s secrets, etc).

```yaml
apiVersion: v1
data:
  client-id: c29tZXJlYWxseWxvbmdzZWNyZXRqdXN0dG9jb3ZlcnN0dWZmCg==
  client-secret: c29tZXJlYWxseWxvbmdzZWNyZXRqdXN0dG9jb3ZlcnN0dWZmCg==
kind: Secret
metadata:
  name: grafana-admin-credentials
  namespace: monitoring
type: Opaque
```

### Creating our Grafana Instance

Create a Grafana instance overriding the configuration for `auth.google`.

Haven't tested this with other means of authentication (Github, Okta, etc) but configuration should be pretty similar in case you want to use any other solution.

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
    auth.google:
      enabled: "true"
      scopes: https://www.googleapis.com/auth/userinfo.profile https://www.googleapis.com/auth/userinfo.email
      auth_url: https://accounts.google.com/o/oauth2/auth
      token_url: https://accounts.google.com/o/oauth2/token
      allowed_domains: your-domain.com
      allow_sign_up: "true"
    server:
      root_url: https://grafana.your-domain.com
...
```

Now, we need to override some of the deployment spec, in order to add the secrets, make sure to point to the right secret name and key.

```yaml
...
deployment:
    spec:
      template:
        spec:
          containers:
            - name: grafana
              env:
                - name: GF_AUTH_GOOGLE_CLIENT_ID
                  valueFrom:
                    secretKeyRef:
                      name: grafana-admin-credentials
                      key: client-id
                - name: GF_AUTH_GOOGLE_CLIENT_SECRET
                  valueFrom:
                    secretKeyRef:
                      name: grafana-admin-credentials
                      key: client-secret
              image: grafana/grafana:10.0.3
...
```

Make sure that have ingress and tls already configured as they are a prerequisite to work with Google SSO.

```yaml
...
  ingress:
    metadata:
      annotations:
        kubernetes.io/ingress.class: nginx
        external-dns.alpha.kubernetes.io/hostname: grafana.your-domain.com
        cert-manager.io/cluster-issuer: letsencrypt-prod
    spec:
      ingressClassName: nginx
      rules:
        - host: grafana.your-domain.com
          http:
            paths:
              - backend:
                  service:
                    name: grafana-service
                    port:
                      number: 3000
                path: /
                pathType: Prefix
      tls:
        - hosts:
            - grafana.your-domain.com
          secretName: grafana-tls-secret
```

Full configuration is below.

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

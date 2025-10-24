---
title: "Kubernetes ServiceAccount Auth"
---

<!-- Intro -->

Starting from v5.20.0, the operator can authenticate to Grafana instances using the mounted Kubernetes ServiceAccount token:

<!-- Examples and options -->

{{< readfile file="jwt-example.yaml" code="true" lang="yaml" >}}

<!-- Result -->

![Account created via JWT authentication in Grafana](./jwt-account.png)

This makes it significantly easier to declaratively disable the default Grafana Admin account.

<!-- Mechanisms -->

Previously it was necessary to manually create Grafana Service/User Accounts.

By using `[auth.jwt]` you can now allow native Kubernetes ServiceAccount to authenticate and be assigned roles:

```ini
# grafana.ini
[auth.jwt]
auto_sign_up = "true"

# Validate JWT agains Kubernetes API server
jwk_set_url = https://${KUBERNETES_SERVICE_HOST}:${KUBERNETES_SERVICE_PORT_HTTPS}/openid/v1/jwks
jwk_set_bearer_token_file = /var/run/secrets/kubernetes.io/serviceaccount/token
tls_client_ca = /var/run/secrets/kubernetes.io/serviceaccount/ca.crt

# Assigns Admin to any ServiceAccount in the `grafana` namespace
role_attribute_path = "contains(\"kubernetes.io\".namespace, 'grafana') && 'GrafanaAdmin' || 'None'"
# If allow_assign_grafana_admin is enabled, GrafanaAdmin is assigned instead
# allow_assign_grafana_admin: "true"
```

The above example will assign Admin/GrafanaAdmin to any ServiceAccount within the `grafana` namespace.

Modifying `role_attribute_path` allows heavily modifying the conditions using any of the claims in the JWT body:

The below JWT has decoded a ServiceAccount token taken from a pod at `/var/run/secrets/kubernetes.io/serviceaccount/token`

Alternatively, a token can be issued with: `kubectl create token -n grafana test-token`

{{< readfile file="jwt-claims.json" code="true" lang="yaml" >}}


## Grafana versions lower than 12.2.0

Older versions of Grafana cannot authenticate with a JWKS endpoint, which is necessary when

```bash
kubectl create configmap kube-root-jwks --from-literal=jwks.json="$(kubectl get --raw /openid/v1/jwks)"
```

{{< readfile file="older-versions.yaml" code="true" lang="yaml" >}}

### Setup

```bash
# Create Grafana instance
# If your Grafana version does not support [auth.jwt].tls_client_ca yet,
# use mounted-jwt-grafana.yml example instead and create the ConfigMap first
kubectl apply -f jwt-grafana.yml

# Create serviceaccount token to curl Grafana instance with
kubectl create token -n grafana grafana-operator-controller-manager > token

# Expose a port
kubectl port-forward svc/jwt-grafana-ca-service 3000:3000 &>/dev/null &

# Curl the instance
curl http://127.0.0.1:3000/api/folders -H "Authorization: Bearer $(cat token)"

# `[]` is a successful response
```

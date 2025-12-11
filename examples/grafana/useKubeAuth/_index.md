---
title: "Kubernetes ServiceAccount JWT Auth"
linkTitle: "Auth: Kubernetes ServiceAccount"
tags:
  - Auth config
---

Among the [auth mechanisms supported by Grafana](https://grafana.com/docs/grafana/latest/setup-grafana/configure-access/configure-authentication/), `[auth.jwt]` stands out as it is uniquely compatible with Kubernetes!

By using this method, it is possible to use Kubernetes ServiceAccount tokens (JWTs) to authenticate with Grafana.

Since version `v5.21.0` of the Grafana-Operator, we support authentication using the projected Kubernetes ServiceAccount JWT when `[auth.jwt]` is configured.


## Configuration
Enable JWT auth for a Grafana instance with `.spec.client.useKubeAuth=true` and configure Grafana to trust JWTs issued by Kubernetes:

{{< readfile file="./resources.yaml" code="true" lang="yaml" >}}

{{% alert title="Note" color="primary" %}}
The example assumes the Grafana operator is installed in the `default` namespace and using the `grafana-operator` ServiceAccount.
Both being the default when installing via the official Helm chart.

Remember to update the `role_attribute_path` accordingly.
{{% /alert %}}

![Account created via JWT authentication in Grafana](./jwt-account.png)

`role_attribute_path` determines the assigned role by the claims in the JWT body:

The example assigns `Admin`, or `GrafanaAdmin` if `allow_assign_grafana_admin: "true"`, to the `grafana-operator` ServiceAccount in the `default` namespace.

If you require more flexibility, you can customize the `role_attribute_path`.
By default, you have the following claims available from the kubernetes service account token:

{{< readfile file="./jwt-claims.json" code="true" lang="yaml" >}}

The below configuration will assign `GrafanaAdmin` to the main ServiceAccount, but `Editor` to any Service account in the `grafana` namespace.

This can be used when other workloads in the cluster need access to Grafana through the same JWT authentication but with less permissions.

```yaml
role_attribute_path: "contains(sub, 'system:serviceaccount:default:grafana-operator') && 'GrafanaAdmin' || contains(\"kubernetes.io\".namespace, 'grafana') && 'Editor' || 'None'"
```

If you intend to use ServiceAccount tokens with the default audience (`aud`) claim, remember to remove the `expect_claims` config from the examples.

## Grafana versions prior to 12.2.0

Older versions of Grafana cannot authenticate with a JWKS endpoint, which is necessary to retrieve the `JWKSet` from Kubernetes.

Users instead need to mount in the `JWKSet` as a file from either a ConfigMap or Secret.

```bash
kubectl create configmap kube-root-jwks --from-literal=jwks.json="$(kubectl get --raw /openid/v1/jwks)"
```

{{< readfile file="./older-versions.yaml" code="true" lang="yaml" >}}


## Issuing Tokens for ServiceAccounts

{{% alert title="Warning" color="secondary" %}}
Always be vary of tokens being leaked, especially when using long-lived Kubernetes ServiceAccount tokens for other purposes than authenticating with the Kubernetes API.
{{% /alert %}}

Tokens can be issued for a service account ad hoc with kubectl.

This could be used for testing or just an easy way to create short lived JWTs for a ServiceAccount with access to Grafana

```bash
# Create serviceaccount JWT and store it in ./token
kubectl create token -n grafana grafana-operator-controller-manager --audience 'operator.grafana.com' --duration 1h >token

# Expose a port
kubectl port-forward svc/jwt-grafana-ca-service 3000:3000 &>/dev/null &

# curl the instance using the token
curl 'http://127.0.0.1:3000/api/folders' -H "Authorization: Bearer $(cat token)"

# An array, even empty `[]`, is a successful response!
```

### Custom token audience

If the default token at `/var/run/secrets/kubernetes.io/serviceaccount/token` is leaked, whatever permissions assigned to the ServiceAccount can be abused by whoever obtains it.

To prevent this, it's highly recommended to create tokens with custom audience claims (`"aud": ["..."]`) that invalidates the token from being used with the Kubernets API.

By default, the `grafana-operator` ServiceAccount can create, Update, and Delete various resources on the cluster level.
To avoid accidental exposure of this service account, the operator mounts a second token at `/var/run/secrets/grafana.com/serviceaccount/token` that cannot be used with the Kubernetes API.
This is the token used for JWT authentication.

```yaml
# The majority of the manifest is omitted for brevity
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana-operator
spec:
  replicas: 1
  template:
    spec:
      containers:
        - name: automation
          image: ghcr.io/grafana/grafana-operator:v5.21.0
          imagePullPolicy: IfNotPresent
          volumeMounts: # Where to mount the serviceaccount
            - name: kubeauth-token-volume
              mountPath: /var/run/secrets/grafana.com/serviceaccount
              readOnly: true
      volumes:
        - name: kubeauth-token-volume
          projected:
            sources:
            - serviceAccountToken:
                audience: operator.grafana.com # Ensures the token cannot authenticate with the Kubernetes API
                path: token
```

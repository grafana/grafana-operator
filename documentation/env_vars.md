# Exposing environment variables to the Grafana container

You can expose extra environment variables to the Grafana container from secrets or config maps in the same namespace as the Grafana CR:

```yaml
spec:
  deployment:
    skipCreateAdminAccount: true
    envFrom:
      - secretRef:
          name: external-credentials
      - configMapRef:
          name: extra-vars
```

This adds all the key value pairs from the secret `external-credentials` and the config map `extra-vars` to the Grafana container.

## Requirements when providing external admin credentials

Typically, this feature is used to provide the credentials for the Grafana admin account. In that case, the following needs to be assured:

1. Admin credentials must be provided via the `GF_SECURITY_ADMIN_USER` and `GF_SECURITY_ADMIN_PASSWORD` environment variables.
2. Set `skipCreateAdminAccount` to `true` to prevent the operator from creating an admin secret.

*NOTE*: The operator still requires an admin account to interact with Grafana. It will try to obtain the credentials from the provided secrets or config maps.
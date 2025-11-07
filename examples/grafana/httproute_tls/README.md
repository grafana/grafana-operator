# HTTPRoute with TLS Example

This example demonstrates HTTPRoute with TLS termination at the Gateway level.

## Prerequisites

- Kubernetes cluster with Gateway API CRDs installed
- A Gateway resource with TLS listener configured
- TLS certificate and key
- grafana-operator installed

## Configuration

The example shows:
- HTTPRoute referencing Gateway's HTTPS listener via `sectionName: https`
- Gateway handles TLS termination
- Backend communication remains HTTP

Note: Update the Secret with your actual TLS certificate and key.

## Deploy

```bash
# Update tls.crt and tls.key in resources.yaml first
kubectl apply -f resources.yaml
```

## Access

Grafana will be accessible at `https://grafana.example.com` once Gateway TLS is properly configured.

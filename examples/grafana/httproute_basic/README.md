# Basic HTTPRoute Example

This example demonstrates basic HTTPRoute configuration for Grafana using Gateway API.

## Prerequisites

- Kubernetes cluster with Gateway API CRDs installed
- A Gateway resource named `my-gateway` in the `default` namespace
- grafana-operator installed

## Install Gateway API CRDs

```bash
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.3.0/standard-install.yaml
```

## Deploy

```bash
kubectl apply -f resources.yaml
```

## Configuration

The example creates:
- Grafana instance with basic HTTP configuration
- HTTPRoute resource pointing to `my-gateway`
- Hostname: `grafana.example.com`

## Access

Once deployed, Grafana will be accessible at `http://grafana.example.com` (depending on your Gateway configuration).

See `resources.yaml` for admin credentials configuration.

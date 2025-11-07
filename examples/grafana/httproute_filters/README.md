# HTTPRoute with Filters Example

This example demonstrates HTTPRoute with request and response header modification filters.

## Prerequisites

- Kubernetes cluster with Gateway API CRDs installed
- A Gateway resource that supports HTTPRoute filters
- grafana-operator installed

## Deploy

```bash
kubectl apply -f resources.yaml
```

## What This Example Shows

- Adding custom request headers (`X-Custom-Header`, `X-Request-ID`)
- Removing unwanted request headers
- Adding response headers (`X-Frame-Options`, `X-Grafana-Version`)

For more information about available filter types and advanced configurations, see the [Gateway API HTTPRoute Filters documentation](https://gateway-api.sigs.k8s.io/api-types/httproute/#filters-optional).

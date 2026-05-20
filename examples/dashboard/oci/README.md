---
title: "Dashboard from OCI artifact"
linkTitle: "Dashboard from OCI artifact"
---

Shows how to load a dashboard JSON file stored inside an OCI artifact in a container registry (e.g. GHCR, ECR, GAR).
Bytes are fetched at reconcile time and never stored in etcd, making this the recommended source for dashboards larger than ~1 MiB.

The `reference` field must include either a tag (`:v1.4.7`) or a digest (`@sha256:...`):

- tag - mutable pointer; operator re-fetches on each reconcile interval and caches via `contentCacheDuration`.
- digest - immutable; guarantees bit-for-bit reproducibility across clusters.

For public registries, omit `pullSecretRef`. For private registries, create a `kubernetes.io/dockerconfigjson` Secret in the same namespace and reference it via `pullSecretRef`.

Push a dashboard artifact with [oras](https://oras.land/):

```bash
echo '{"title":"My Dashboard","panels":[]}' > board.json
oras push ghcr.io/team-a/dashboards:v1.4.7 \
  --artifact-type application/vnd.grafana.dashboard+json \
  board.json:application/json
```

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

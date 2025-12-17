---
title: "Dynamic Resource Patching"
---

## Summary

We often get issues and feature requests around supporting arbitrary modification of the resources before being applied to Grafana.
At the time of writing, we support limited replacement using `valuesFrom` in some resources but this is not very useable and inconsistent.
With the upcoming addition of generic manifests, we need a stable way to perform manipulation of the resource before submitting to the Grafana instance.

## Proposal

All resources that get applied to the Grafana instance should support a new section `.spec.patch` that contains:
* A `yq` script
* A list of `env` variables to dynamically fetch data.

This could look something like this:

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaLibraryPanel
metadata:
  name: grpc-server-success-rate
spec:
  instanceSelector:
    matchLabels:
      env: dev
      region: us-central1
  url: http://assets.example.com/library-panels/grpc-server-success-rate.json
  patch:
    script: '.title=env(PANEL_TITLE)'
    env:
      - name: PANEL_TITLE
        valueFrom:
          secretKeyRef:
            name: some-secret
            key: some-key
```

We can also extend the supported options in `valueFrom` to support information from the Grafana instance like a pseudo downward api:
```yaml
patch:
  script: '.title=env(PANEL_TITLE)'
  env:
    - name: PANEL_TITLE
      valueFrom:
        grafanaInstance:
          fieldPath: '.spec.version'
```

---
title: "Patch sourced dashboard content"
linkTitle: "Patch sourced dashboard content"
---

Shows how to use `spec.patch` to modify a dashboard's JSON model with `jq` scripts.

Unlike `spec.variables`/`spec.datasources`/`spec.envs`, which target specific, purpose-built substitution points (template variable defaults, `${...}` placeholders, and env var interpolation, respectively), `spec.patch` allows arbitrary modification of the dashboard model.
It runs last, after those other mechanisms have already resolved.
An example use case is to replace the tags for a dashboard whose JSON comes from an external source (`grafanaCom`, remote `url`, `oci`, etc...).

Each entry in `scripts` is a `jq` expression evaluated against the model in order, using the previous script's output.
The `spec.patch.env` array makes secret/configmap/Grafana instance values available to those scripts as `env.NAME`.
`spec.patch.env[].valueFrom.grafanaRef` selects a field off the matching `Grafana` instance (e.g. `metadata.name`).
Unlike `secretKeyRef`/`configMapKeyRef`, a `grafanaRef` value is resolved separately for each matching `Grafana` instance, so the patched dashboard content can vary across instances.

The operator protects the `uid` field it manages itself from patch scripts.
If a script changes it, the operator restores the original value.

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

---
title: "Patch sourced dashboard content"
linkTitle: "Patch sourced dashboard content"
---

Shows how to use `spec.patch` to modify a dashboard's JSON model with `jq` scripts.

Unlike `spec.variables`/`spec.datasources`/`spec.envs`, which target specific, purpose-built substitution points (template variable defaults, `${...}` placeholders, and env var interpolation, respectively), `spec.patch` allows arbitrary modification of the dashboard model.
It runs last, after those other mechanisms have already resolved.
An example use case is to replace the tags for a dashboard whose JSON comes from an external source (`grafanaCom`, remote `url`, `oci`, etc...).

Each entry in `scripts` is a `jq` expression evaluated against the model in order, using the previous script's output.
The `spec.patch.env` array makes secret/configmap values available to those scripts as `env.NAME`.
The `spec.patch.env[].valueFrom.grafanaRef` field is not supported here.
A dashboard's model is resolved once and shared across every matching `Grafana` instance, so there is no single instance for a `grafanaRef` to resolve against and declaring one is rejected at admission.

The operator protects the fields it manages itself (`id`, `uid`) from patch scripts.
If a script  changes either, the operator restores the original value and emits a `ProhibitedPatchDetected` warning event.

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

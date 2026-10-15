---
title: "Variable override"
linkTitle: "Variable override"
---

This example shows how to override the default (`current`) value of a dashboard's template
variables (`templating.list[]`) from the `GrafanaDashboard` resource.

Unlike `spec.datasources`, which does a `${...}` string replacement and requires the
dashboard author to embed placeholders, `spec.variables` operates on the parsed model. It
works on dashboards you do not author, such as those pulled from
`grafanaCom`, `url`, `oci`, or a shared `configMapRef`.

For each entry, the operator finds the matching variable by `name` and sets its
`current.text`/`current.value` (and reconciles `options[]` so the selected value is valid).
A `name` that is not present in the dashboard is ignored. Datasource-type variables are
overridden the same way.

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

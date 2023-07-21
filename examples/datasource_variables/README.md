---
title: "Datasource variable"
linkTitle: "Datasource variable"
---

This example shows how to reference values from a secret in data source fields.
Values can be assigned using `spec.valuesFrom`:

```yaml
  valuesFrom:
    - targetPath: "secureJsonData.httpHeaderValue1"
      valueFrom:
        secretKeyRef:
          name: "credentials"
          key: "PROMETHEUS_TOKEN"
```

The Operator will look for a key with the name `PROMETHEUS_TOKEN` in a secret with the name `credentials`.
It will then inject the value into `secureJsonData.httpHeaderValue1`:

```yaml
  datasource:
    secureJsonData:
      "httpHeaderValue1": "Bearer ${PROMETHEUS_TOKEN}"
```

The Operator expects a string to be present with the replacement pattern.

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

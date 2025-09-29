---
title: "Dashboard gzipped"
linkTitle: "Dashboard gzipped"
---

Shows how to store the dashboard definition (json) as gzip compressed.

```shell
cat dashboard.json | gzip | base64 -w0
```

{{< readfile file="dashboard.json" code="true" lang="json" >}}

Should provide

```txt
H4sIAAAAAAAAA4WQQU/DMAyF7/0VVc9MggMgcYV/AOKC0OQubmM1jSPH28Sm/XfSNJ1WcaA3f+/l+dXnqk5fQ6Z5qf3eubt5VlKHCTXvNAaH9RtE2zKI2fQnCgFNsxihj8n39V3mqD/zQwMyXE004ol95q3wMaIsEhpSaPMTlT0WasngK3sVdlN6By4uUi8Q7AezUwpJeig4gEe3ajItTfM5T5l0wuNUwfNx82RLg9nLhTeZXW4iAu2GVHcVNPEtByX2tyuzJtgJRrslrygHKJ3WsZhuCkq+X8c6ivrXDd6zwrLrX3vZP/3PY1yuHHcWR/hEiSlmutpzEQ5XdF+IIz+Uzpeq+gWtMMT1HwIAAA==
```

And simply add this output to your dashboard CR

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

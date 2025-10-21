---
title: "Image renderer via sidecar example"
---

This example configures the [Grafana image rendering service](https://grafana.com/docs/grafana/latest/setup-grafana/image-rendering/) for use in alerting & reporting.

For production use, ensure that the image renderer has enough resources. Refer to the [recommendations in the official documentation](https://grafana.com/docs/grafana/latest/setup-grafana/image-rendering/#memory-requirements) for more information.

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

---
title: "Custom Version and Images"
---

It's possible to overwrite the default Grafana image in multiple ways.

When applying a new `Grafana` CR, if `.spec.version` is absent, the operator will populate the field with the current default.

This prevents unplanned restarts of Grafana instances when the operator is upgraded and the default is updated.

{{< readfile file="version.yaml" code="true" lang="yaml" >}}


The `.spec.version` field supports setting image names allowing the use of private registries or use alternate images.

{{< readfile file="image.yaml" code="true" lang="yaml" >}}

Additionally, it's possible to lock images with `sha256` hashes to ensure the same OCI Artifact is retrieved.

{{< readfile file="sha-locked.yaml" code="true" lang="yaml" >}}

---
title: Configuration
description:
weight: 20
---

## Cache settings

To reduce memory usage, the operator does not cache and watch everything by default.
You should not notice this during regular operations but there is one important detail to call out: the operator is _not_ notified if secrets or config maps change and will wait for the scheduled resync before applying the changes.

If immediate updates are important to you, there are two ways to accomplish this:

* Set `--caching-level="off"` - This increases memory usage but makes every update available immediately.
* Label selected resources with `app.kubernetes.io/managed-by: grafana-operator`. The operator will watch these resources in the default cache level as well.

## All settings

All available options in the latest Grafana-Operator version.

Most, if not all options should be available with all installation methods.

{{< readfile file="help.txt" code="true" >}}

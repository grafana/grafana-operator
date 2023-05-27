---
author: "Edvin 'NissesSenap' Norling"
date: 2023-05-27
title: "v4 to v5 dashboard migration"
linkTitle: "v4 to v5 dashboard migration"
description: "How to migrate grafanaDashboard CRD from v4 to v5"
---

We are getting close to release version 5 of grafana-operator.
We have done a number of breaking changes from v5 + version updates so we thought we would supply a small script for inspiration to migration dashboards from v4 to v5.
Due to the complexity and the low amount of instance we saw no need to write a migration script for the other resources.

This script isn't meant to solve all potential use-cases but rather an idea on how you can solve it.
If write a better script feel free to share it with the rest of the community.

## Migration

The biggest difference from v4 to v5 is that the label selector isn't on the grafana instance, instead it's on the dashboard.
In short instead of the grafana instance choosing which dashboards to apply the dashboard chooses which grafana instances to apply to.

So we mush add a `instanceSelector` and update the API version.

Below you can see our v4 sample dashboard.

```yaml
apiVersion: integreatly.org/v1alpha1
kind: GrafanaDashboard
metadata:
  name: simple-dashboard
  labels:
    app: grafana
spec:
  json: >
    {
      "id": null,
      "title": "Simple Dashboard",
      "tags": [],
      "style": "dark",
    }
```

And here you can see a v5 dashboard and what we are aiming for.

```yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafanadashboard-sample
spec:
  resyncPeriod: 30s
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  json: >
    {
      "id": null,
      "title": "Simple Dashboard",
      "tags": [],
      "style": "dark",
    }
```

## Script

I'm using [yq](https://github.com/mikefarah/yq) to perform changes on my dashboard files, which is the jq equivalent but for yaml.

```sh
#!/bin/sh

# A script to migrate grafana-operator dashboard from v4 to v5

# Usage: ./v4-v5migration.sh file-name.yaml

filename=$1

# Update api version
yq -i '.apiVersion = "grafana.integreatly.org/v1beta1"' $filename

# remove labels.app
yq -i 'del(.metadata.labels.app)' $filename
# if metadata.label is empty remove metadata.labels
if [[ $(yq '.metadata.labels' $filename) == {} ]]; then
  yq -i 'del(.metadata.labels)' $filename
fi

# resyncPeriod is optional, if you don't want it just comment it.
yq -i '.spec.resyncPeriod = "30s"' $filename
# Add instanceSelector.matchLabels.dashbords: grafana
yq -i '.spec.instanceSelector.matchLabels.dashboards = "grafana"' $filename
# For some reason yq adds \n in the json data, so we need to remove it
cat $filename | tr -s '\n' '\n' > tmp.yaml && mv tmp.yaml $filename
```

So lets run this script over all my dashboard files located in a folder ending with `.yaml`
Don't forget to take a backup of your files before running the script, they will be changed in place.

```bash
for file in $(ls *.yaml); do ./v4-v5migration.sh $file; done
```

I hope this small script gave you some idea how your dashboards can be migrated to version 5 of the operator.

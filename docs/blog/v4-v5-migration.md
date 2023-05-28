---
author: "Edvin 'NissesSenap' Norling"
date: 2023-05-27
title: "v4 to v5 migration"
linkTitle: "v4 to v5 migration"
description: "How to migrate grafana-operator from v4 to v5"
---

We are getting close to release version 5 of grafana-operator.
As a part of version 5 we have remade the operator from both with code and API and thus contains a number of breaking changes.
We recommend that you read through the changes in our [intro blog]({{< ref "/blog/v5-intro.md" >}}).

The operator supports multiple installation solution like

- [Helm]({{< ref "/docs/installation/helm.md">}})
- [Kustomize]({{< ref "/docs/installation/kustomize.md">}})
- OCP OLM

Look [here]({{< ref "/docs/">}}) for documentation.
Just like earlier we have also created a big [example library]({{< ref "/docs/examples/">}}) on how to configure the operator.
For an extended overview of all possible configuration options look at our [API documentation]({{< ref "/docs/api.md">}})

As part of the migration we thought we would supply a small script for inspiration to migration dashboards from v4 to v5.
Due to the complexity and the low amount of instance we saw no need to write a migration script for the other resources.

This script isn't meant to solve all potential use-cases but rather an idea on how you can solve it.
If you write a better script feel free to share it with the rest of the community in our slack or through a PR.

## Dashboard migration

The biggest difference from v4 to v5 is that the label selector isn't on the grafana instance, instead it's on the dashboard.
In short instead of the grafana instance choosing which dashboards to apply the dashboard chooses which grafana instances to apply to.

So we mush add a `instanceSelector` and update the API version. We also most updated the `apiVersion` and optionally add the `resyncPerio`.

Below you can see a v4 sample dashboard.

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

And here is the same dashboard but adapted to v5.

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

The script uses [yq](https://github.com/mikefarah/yq) to perform changes on my dashboard files, which is the jq equivalent but for yaml.
In the script you will find comment that explain all the steps that we are taking.
Some of the changes are needed while others isn't.
Please adapt the values according to your needs, lets call the script `v4-v5migration.sh`.

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

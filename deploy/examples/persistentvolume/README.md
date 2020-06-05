# Grafana with persistent storage

This example demonstrates how to deploy Grafana with persistent storage.

Grafana uses sqlite database to store user data by default, we can use `PersistentVolume` to persist such user's data.

## Installation

1. Make sure the operator is running, then create the Grafana:

```shell script
$ kubectl apply -f deploy/examples/persistentvolume
```

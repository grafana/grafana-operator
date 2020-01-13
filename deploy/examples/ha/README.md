# High Availability Deployment

This example demonstrates how to deploy Grafana in high availability mode running multiple replicas and using Postgres as a database.

Grafana will by default also use the database for session storage. To keep user sessions separate, the `remote_cache` configuration can be used. 

## Installation

1. Make sure the operator is running, then create the templates in this directory:

```shell script
$ kubectl apply -f deploy/examples/ha
```

*NOTE:* This examples uses an `emptyDir` storage for Postgres and is only meant for demo purposes.
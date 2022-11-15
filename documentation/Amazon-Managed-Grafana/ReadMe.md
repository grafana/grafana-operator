# Design proposal 001 Amazon Managed Grafana Integration

## Summary

Introduce integration to Amazon Managed Grafana with Grafana Operator

This document contains the complete design required for integrating Amazon Managed Grafana with Grafana Operator.
This includes design elements to support the following with Grafana operator :
- Ability to read keys from external key store such as AWS Secrets manager for remote Grafana instance such as Amazon Managed Service for Prometheus.
- Ability to add AWS data sources such as Amazon Managed Grafana, Amazon CloudWatch.
- Ability to create Grafana dashboards on a remote Grafana Instance such as Amazon Managed Grafana.
- Ability to create alerting on a remote Grafana Instance such as Amazon Managed Grafana.

## Info

status: Draft

## Motivation

Cloud providers such Amazon Web Services (AWS) have started providing managed grafana services such as Amazon Managed Grafana which decouples to responsibilities of managing a grafana instanaces from ops personas and kubernetes environment. Amazon Managed Grafana is a fully managed service based on open-source Grafana that makes it easier for you to visualize and analyze your operational data at scale.

Currently Grafana operator has an integration to add Amazon Managed Service for Promethus (AMP) as a data source to a Grafana instances hosted in a kubernetes environment. As more customers starting to use managed grafana services such as Amazon Managed Grafana, expanding the Grafana operator to support remote grafana instances becomes ineviatable. With ability to integrate to managed grafana services such as Amazon Managed Grafana and create dashboards and alerting on a remote Grafana instances offloads responsibilities of managing a grafana instanaces from ops personas which helps them to focus developing the features required for their business. This helps the customer teams to move from self managed Grafana instance on their Kubernetes environments to Pay as you go model of Grafana instances provided by providers.

## Out of scope

In the design, integration to remote managed Grafana providers other than Amazon Managed Grafana is out of scope at this point. We may add additional integrations as part of future road map items.

## Verification

- Create integration tests for adding keys for remote Grafana instance such as Amazon Managed Grafana.
- Create integration tests to add AWS data sources such as Amazon Managed Service for Prometheus, Amazon CloudWatch.
- Create integration tests to create Grafana dashboards on a remote Grafana Instance such as Amazon Managed Grafana.
- Create integration tests to create alerting on a remote Grafana Instance such as Amazon Managed Grafana.

## Current

Currently the grafana operator supports the following for only self managed Grafana Instance :
- Adding remote data sources such as Amazon Managed Service for Prometheus.
- Creating Dashboards.
- Setting up alerting.
- And Many more.


## Proposal

In short the proposal in this document is about enhancing the Grafana Operator to support the integration of managed grafana services such as Amazon Managed Grafana. We would need to enhance the current version of Grafana Operator to support the following :

- Reading keys for accessing Amazon Managed Grafana from AWS Secrets Manager.
- Adding AWS data sources to Amazon Managed Grafana.
- Creating Grafana dashboards on Amazon Managed Grafana.
- Creating alerting on Amazon Managed Grafana.

### Reading keys for accessing Amazon Managed Grafana from AWS Secrets Manager

Today Grafana Operator supports Kubernetes role, rolebinding, service account mechanisms to access other kubernetes resources. We need to add a mechanism to allow Grafana Operator to read API keys for remote Grafana Instance from a secret store such as AWS Secrets Manager using [external-secrets](https://github.com/external-secrets/external-secrets). 

For a demonstration standpoing, after `external-secrets` helm is installed, step 1 is the create a secret for AWS credentials :

```.bash
kubectl create secret generic aws-secret \
  --from-literal=access-key=$ACCESS_KEY \
  --from-literal=secret=$SECRET_KEY
```

> Below shows how you can sync secrets between AWS Secrets Manager for Grafana Operator :

```.bash
cat <<EOF | kubectl apply -f - 
apiVersion: external-secrets.io/v1beta1
kind: ClusterSecretStore
metadata:
  name: grafana-secret-store
spec:
  provider:
    aws:  # set secretStore provider to AWS.
      service: SecretsManager # Configure service to be Secrets Manager
      region: us-west-2   # Region where the secret is.
      auth:
        secretRef:
          accessKeyIDSecretRef: 
            name: aws-secret # References the secret we created
            namespace: default
            key: access-key  
          secretAccessKeySecretRef:
            name: aws-secret
            namespace: default
            key: secret
EOF
```

> Finally, Grafana operator should be enhanced to read secrets like below :

```.bash

apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: grafana-token-secret
  namespace: grafana
spec:
  refreshInterval: 1m
  secretStoreRef:
    name: grafana-secret-store #The secret store name we have just created.
    kind: ClusterSecretStore
  target:
    name: grafana-secret # Secret name in k8s
  data:
  - secretKey: grafana-token # which key it's going to be stored
    remoteRef:
      key: grafana-api-key # Our secret-name goes here
```


### Adding AWS data sources to Amazon Managed Grafana

Today Grafana Operator supports [Amazon Managed Service for Prometheus](https://github.com/grafana-operator/grafana-operator/blob/master/deploy/examples/datasources/AWS-Prometheus.yaml) as a data source. Grafana operator solution should be enhanced to add following AWS data sources (minimally Amazon CloudWatch) :
- Amazon CloudWatch
- Amazon OpenSearch Service
- AWS IoT SiteWise
- AWS IoT TwinMaker
- Amazon Timestream
- Amazon Athena
- Amazon Redshift
- AWS X-Ray

This is not a road block feature but nice to have to support all AWS data sources.

In order to accomplish this, "Reading keys for accessing Amazon Managed Grafana from AWS Secrets Manager" feature above becomes a prerequisite. `GrafanaDataSource` CRD should be modified to use `grafana-secret` to add AWS data sources as `GrafanaDataSource` on Amazon Managed Grafana. This allows customers to use GitOps approach to add AWS data sources as `GrafanaDataSource` on remote Grafana instances such as Amazon Managed Grafana.

### Creating Grafana dashboards on Amazon Managed Grafana.

Today Grafana operator supports the creation of Grafana dashboards on self managed grafana instance in the cluster where Grafana operator is installed. Grafana operator should be enhanced to support creating dashboards on Amazon Managed Grafana. 

In order to accomplish this, "Reading keys for accessing Amazon Managed Grafana from AWS Secrets Manager" feature above becomes a prerequisite. `GrafanaDashboard` CRD should be modified to use `grafana-secret` to create `GrafanaDashboard` on Amazon Managed Grafana. This allows customers to use GitOps approach to create Grafana Dashboards on remote Grafana instances such as Amazon Managed Grafana.

### Creating Grafana alerting on Amazon Managed Grafana.

Today Grafana operator supports the creation of Grafana notification on self managed grafana instance in the cluster where Grafana operator is installed. Grafana operator should be enhanced to support creating notifications on Amazon Managed Grafana. 

In order to accomplish this, "Reading keys for accessing Amazon Managed Grafana from AWS Secrets Manager" feature above becomes a prerequisite. `GrafanaNotificationChannel` CRD should be modified to use `grafana-secret` to create `GrafanaNotificationChannel` on Amazon Managed Grafana. This allows customers to use GitOps approach to create Grafana alerting on remote Grafana instances such as Amazon Managed Grafana.

## Related issues

- [Issue 402](https://github.com/grafana-operator/grafana-operator/issues/402)

## References

- [Amazon Managed Grafana](https://docs.aws.amazon.com/grafana/latest/userguide/what-is-Amazon-Managed-Service-Grafana.html)
- [Amazon Managed Grafana Data Sources](https://docs.aws.amazon.com/grafana/latest/userguide/AMG-data-sources.html)
- [External Secrets - ClusterSecretStore](https://external-secrets.io/v0.4.4/api-clustersecretstore/)
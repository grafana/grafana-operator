# Design proposal 001 External Grafana Instance Integration

## Summary

Introduce integration to External Grafana Instances with Grafana Operator.

This document contains the complete design required for integrating external Grafana instances such as Amazon Managed Grafana with Grafana Operator.
This includes design elements to support the following with Grafana operator :
- Ability to define an external Grafana instance as Grafana source to Grafana operator
- Ability to create Grafana dashboards on a remote Grafana Instance.
- Ability to add external cloud data sources to an external Grafana instance.
- Ability to create alerting on a remote Grafana Instance.

## Info

status: Draft

## Motivation

Cloud providers such Amazon Web Services (AWS) have started providing managed remote grafana services such as Amazon Managed Grafana which decouples the responsibilities of managing a grafana instanaces from ops personas and kubernetes environment. 

Currently Grafana operator has an integration to add Amazon Managed Service for Promethus (AMP) as a data source to Grafana instances hosted in a kubernetes environment. As more customers starting to use external grafana services, expanding the Grafana operator to support remote Grafana instances becomes inevitable. Adding ability to integrate with external grafana services, adding data sources, creating dashboards and alerting on a remote Grafana instances offloads responsibilities of managing a grafana instanaces from ops personas which helps them to focus on developing the features required for their business. This helps the customer teams to move from self managed Grafana instance on their Kubernetes environments to Pay as you go model on Grafana instances provided by providers.

## Out of scope

In the design, integration to remote managed Grafana providers other than Amazon Managed Grafana is out of scope at this point. We may add additional integrations as part of future road map items.

## Verification

- Create integration tests for adding keys for remote Grafana instance.
- Create integration tests to create Grafana dashboards on a remote Grafana Instance.
- Create integration tests to add cloud data sources such as Amazon Managed Service for Prometheus, Amazon CloudWatch to remote Grafana instance.
- Create integration tests to create alerting on a remote Grafana Instance.

## Current

Currently the grafana operator supports the following for only self managed Grafana Instance :
- Adding remote data sources such as Amazon Managed Service for Prometheus.
- Creating Dashboards.
- Setting up alerting.
- And Many more.

## Proposal

In short the proposal in this document is about enhancing the Grafana Operator to support the integration to managed grafana services. We would need to enhance the current version of Grafana Operator to support the following :

- Defining external Grafana instance as Grafana source to Grafana operator.
- Creating Grafana dashboards on external Grafana instance.
- Adding data sources to external Grafana instance.
- Creating alerting on external Grafana instance.

### Defining external Grafana instance as Grafana source to Grafana operator.

Today Grafana operator supports self managed grafana instance as a Grafana source to Grafana operator. With this new feature, we would like to add a new CRD by name `GrafanaInstance` which basically creates an identity to an external Grafana instance. 

> CRD design of how a `GrafanaInstance` will look like. In this sample design, `url` should be preloaded in a ConfigMap and `grafana_api_key` should be loaded to a Secret. Choice of using external secrets or loading secrets manually is end user responsibility :

```.yaml
apiVersion: integreatly.org/v1alpha1
kind: GrafanaInstance
metadata:
  name: amazon-managed-grafana
spec:
  external:
    url: <type ConfigMapSelector>
    grafana_api_key: <type SecretKeySelector>
```

> CRD design for `GrafanaInstance` with a default setup for an external Grafana Instance. With this all Grafana Operator operations such as `GrafanaDataSource`, `GrafanaDashboard` will point to this default instance :

```.yaml
apiVersion: integreatly.org/v1alpha1
kind: GrafanaInstance
metadata:
  name: amazon-managed-grafana
  annotations:
    grafanainstance.kubernetes.io/is-default-grafana: "true"
spec:
  external:
    url: <type ConfigMapSelector>
    grafana_api_key: <type SecretKeySelector>
```

### Creating Grafana dashboards on Amazon Managed Grafana.

Today Grafana operator supports the creation of Grafana dashboards on self managed grafana instance in the cluster where the Grafana operator is installed. Grafana operator should be enhanced to support creating dashboards on external Grafana instance. 

> In order to accomplish this,  `GrafanaDashboard` CRD should be modified to use `GrafanaInstance` as an input parameter to point to external Grafana instance. This allows customers to use GitOps approach to create Grafana Dashboards on remote Grafana instances.

```.yaml
apiVersion: integreatly.org/v1alpha1
kind: GrafanaDashboard
metadata: 
  name: grafana-dashboard-from-url
  labels: 
    app: grafana
spec:
  grafanainstance: "amazon-managed-grafana"
  url: https://raw.githubusercontent.com/integr8ly/grafana-operator/master/deploy/examples/remote/grafana-dashboard.json
```

### Adding data sources to external Grafana instance.

Today Grafana operator supports [Amazon Managed Service for Prometheus](https://github.com/grafana-operator/grafana-operator/blob/master/deploy/examples/datasources/AWS-Prometheus.yaml) as a data source. Grafana operator solution should be enhanced to add following data sources (minimally Amazon CloudWatch) :
- Amazon CloudWatch
- Amazon OpenSearch Service
- AWS IoT SiteWise
- AWS IoT TwinMaker
- Amazon Timestream
- Amazon Athena
- Amazon Redshift
- AWS X-Ray

One example per below for each of the above listed data source in `grafana-operator` github is desirable for our users.

> CRD design for `GrafanaInstance` to add external cloud based data sources for an external Grafana Instance. With this all Grafana Operator operations such as `GrafanaDataSource`, `GrafanaDashboard` will point to this default instance :

```.yaml
apiVersion: integreatly.org/v1alpha1
kind: GrafanaDataSource
metadata:
  name: aws-cloudwatch-datasource
spec:
  grafanainstance: "amazon-managed-grafana"
  datasources:
  - access: proxy
    editable: true
    isDefault: true
    name: cloudwatch
    type: cloudwatch
    region: <type ConfigMapSelector>
    iamRole: <type ConfigMapSelector>
```

Note: `grafanainstance` is not mandatory if its setup as default external Grafana instance in `GrafanaInstance` CRD. 

This is not a road block feature but nice to have to support all cloud based data sources.

### Creating Grafana alerting on Amazon Managed Grafana.

We need a new CRD `GrafanaAlerting` to support alerting for external Grafana instance and to internal Grafana instance. 

> The new CRD can look like below :

```.yaml
apiVersion: integreatly.org/v1alpha1
kind: GrafanaAlerting
metadata:
  name: pager-duty-channel
  labels:
    app: grafana
spec:
  grafanainstance: "amazon-managed-grafana"
  name: pager-duty-channel.json
  json: >
    {
      "uid": "pager-duty-alert-notification",
      "name": "Pager Duty alert notification",
      "type":  "pagerduty",
      "isDefault": true,
      "sendReminder": true,
      "frequency": "15m",
      "disableResolveMessage": true,
      "settings": {
        "integrationKey": "put key here",
        "autoResolve": true,
        "uploadImage": true
     }
    }
```

## Related issues

- [Issue 402](https://github.com/grafana-operator/grafana-operator/issues/402)

## References

- [Amazon Managed Grafana](https://docs.aws.amazon.com/grafana/latest/userguide/what-is-Amazon-Managed-Service-Grafana.html)
- [Amazon Managed Grafana Data Sources](https://docs.aws.amazon.com/grafana/latest/userguide/AMG-data-sources.html)
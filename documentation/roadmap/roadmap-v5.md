# Grafana Operator v5.0 - Roadmap

A list of features and changes for the 5.0 release of the Grafana Operator. The purpose of this document is to allow users to discuss and extend the proposed changes.

## Multi Namespace Support

The Operator should be capable of managing multiple Grafana instances in multiple namespaces. It is no longer a requirement that the Grafana instance is deployed to the same namespace as the Operator.

Currently being worked on? Yes: [#599](https://github.com/grafana-operator/grafana-operator/pull/599)

## Reconciler Update

We want to switch to use `controllerutils.CreateOrUpdate` instead of manually checking the current cluster state in every reconciliation. This will reduce both, code size and Kubernetes API requests.
Issue [362](https://github.com/grafana-operator/grafana-operator/issues/362)

Currently being worked on? Yes. No PR yet.

## grafana CRD changes

Update the grafana CRD to be easier to use and provide more customization opportunities.

We will also change a number of defaults to make the operator and the grafana instances more secure by default.

Currently being worked on? Yes. Design document [684](https://github.com/grafana-operator/grafana-operator/pull/684).

## CRD version

The CRD version will be updated to v1beta1. The group will change from `integreatly.org` to `grafana-operator`.

Currently being worked on? No.

## Updated handling of folders

Currently we are not deleting empty folders to account for unmanaged dashboards. This policy will change in 5.x and we will delete empty folders assuming that all dashboards are managed.

Currently being worked on? Yes. [#657](https://github.com/grafana-operator/grafana-operator/pull/657)

## Align Routes and Ingresses

Currently Routes and Ingresses support different features (e.g. no TLS options exposed for Routes). This should be streamlined so that both support the same features.

Currently being worked on? No.

## Dashboard discovery

We want to flip the dashboard discovery logic. Instead of putting label selectors on the `Grafana` CR, they will be put on the `GrafanaDashboard` CR.

Dashboards will select the Grafana instances that should import them.

Currently being worked on? No.

## grafanadatasource use grafana API

Today the operator adds grafana datasources by mounting a configmap in to your grafana deployment.
We want to change this and use the grafana API just like we do in the rest of the controllers.

This will most likely create breaking changes in the grafanadatasource CRD.

Currently being worked on? No.

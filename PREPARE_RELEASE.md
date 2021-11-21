# Prepare for a new release

In this repo you will need to perfrom the following tasks manually

There are allot of information on what is needed to manage OLM [compatible operators](https://redhat-connect.gitbook.io/certified-operator-guide/ocp-deployment/operator-metadata/creating-the-csv).

- Update `version/version.go`
- Update the `Makefile`
- Update `replaces` field in `config/manifests/bases/grafana-operator.clusterserviceversion.yaml`
- Update `CreatedAt` field in `config/manifests/bases/grafana-operator.clusterserviceversion.yaml`
  You will have to asses when it's going to get merged and you will be able to do a release.
  You should make sure it's the same date. If not you will have to change it
  manually when creating PR:s to OLM.

      # This is how the time syntax should look.
      $ docker inspect quay.io/grafana-operator/grafana-operator:v4.0.2 |jq '.[0].Created'
      "2021-11-22T10:34:12.173861869Z"
      # 2021-11-22T10:34:12Z is enough
- Run `make bundle`
- Create a PR and get it merged
- Create a new release with the new tag

To update the OLM channels you will need to create a PR in the following repos:

- [community operators](https://github.com/k8s-operatorhub/community-operators)
- [RedHat operators](https://github.com/redhat-openshift-ecosystem/community-operators-prod/tree/main/operators)

## Community operators

Create a new version of the operator under
[https://github.com/k8s-operatorhub/community-operators/tree/main/operators/grafana-operator](https://github.com/k8s-operatorhub/community-operators/tree/main/operators/grafana-operator)
that matches the new tag.

Copy the content of `bundle/manifests/` in the grafana-operator repo from the taged version.

Update `operators/grafana-operator/grafana-operator.package.yaml` with the new tag.

## RedHat operators

Create a new version of the operator under
[https://github.com/redhat-openshift-ecosystem/community-operators-prod/tree/main/operators/grafana-operator](https://github.com/redhat-openshift-ecosystem/community-operators-prod/tree/main/operators/grafana-operator)
that matches the new tag.

Copy the content of `bundle/manifests/` in the grafana-operator repo from the taged version.

Update `grafana-operator/grafana-operator.package.yaml` with the new tag.

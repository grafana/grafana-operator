# Prepare for a new release

In this repo you will need to perfrom the following tasks manually

There is a lot of information on what is needed to manage OLM [compatible operators](https://redhat-connect.gitbook.io/certified-operator-guide/ocp-deployment/operator-metadata/creating-the-csv).

- Update `version/version.go` bumping the minor or patch version. Major versions shouldn't be released without discussion and approval of the maintainers!
- Update the `Makefile`
- Update `containerImage` field in `config/manifests/bases/grafana-operator.clusterserviceversion.yaml`
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
- Run `make bundle-kustomization KUSTOMIZE_TAG=v4.0.2` the KUSTOMIZE_TAG should match the new tag
- Update `deploy/manifests/latest/deployment.yaml` so it matches with the new image version
- Create a PR and get it merged
- Create a new release with the new tag, make sure to compile release notes (github has an option to do this for you)

To update the OLM channels you will need to create a PR in the following repos:
You will need to sign your commits, and make sure they are squashed before submitting the PR, be aware that these repos also require you to sign certain open-source agreement documents as part of the CI-checks.

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

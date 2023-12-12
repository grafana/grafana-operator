# Prepare for a new release

In this repo you will need to perform the following tasks manually

## New release

Hurray, time for a new release.
Follow the instructions below.

Currently our **documentation** needs to be updated in two spots.

You need to change the version in [hugo/config.toml](hugo/config.toml).
You also need to change the version for helm in [deploy/helm/grafana-operator/Chart.yaml](deploy/helm/grafana-operator/Chart.yaml).
After that you need to run `make helm/docs` which will generate the changes to become visible on our homepage.

- Update the `Makefile` version
- `Helm` look if any rbac rules have been changed in the last release, if so verify that the rbac rules for the helm chart is correct. This should be done in those PRs but it don't hurt take an extra look.
- Create a PR and get it merged
- Create a new release with the new tag, make sure to compile release notes (github has an option to do this for you)

## OLM

After version v5.4.1, we no longer update the image version in this repo, but only upstream in the OLM repos.
This to support disconnected mode, for more information see [PR 1234](https://github.com/grafana/grafana-operator/pull/1234).

After cutting a new release according to the instructions above, run the below instructions in this repo and create a PR to the different upstream repos, there is no need to create a PR to this repo.

There is a lot of information on what is needed to manage OLM [compatible operators](https://redhat-connect.gitbook.io/certified-operator-guide/ocp-deployment/operator-metadata/creating-the-csv).

- Run `make generate` & `make manifests`
- Update the following fields under `metadata.annotations` in `config/manifests/bases/grafana-operator.clusterserviceversion.yaml`:
  - `containerImage`
  - `replaces`
  - `createdAt`: Make sure that createdAt matches when the image was published. If not you will have to change it manually when creating PR:s to OLM.
    ```
    # This is how the time syntax should look.
    $ docker inspect ghcr.io/grafana/grafana-operator:v5.0.0 |jq '.[0].Created'
    "2023-11-22T10:34:12.173861869Z"
    # 2023-11-22T10:34:12Z is enough
    ```
- Run `make bundle/redhat`

To update the OLM channels you will need to create a PR in the following repos:
You will need to sign your commits, and make sure they are squashed before submitting the PR, be aware that these repos also require you to sign certain open-source agreement documents as part of the CI-checks.

- [community operators](https://github.com/k8s-operatorhub/community-operators)
- [RedHat operators](https://github.com/redhat-openshift-ecosystem/community-operators-prod/tree/main/operators)

### Community operators


Create a new version of the operator under
[https://github.com/k8s-operatorhub/community-operators/tree/main/operators/grafana-operator](https://github.com/k8s-operatorhub/community-operators/tree/main/operators/grafana-operator)
that matches the new tag.

Copy the content of `bundle/manifests/` in the grafana-operator repo from the taged version.

Update `operators/grafana-operator/grafana-operator.package.yaml` with the new tag.

### RedHat operators

Create a new version of the operator under
[https://github.com/redhat-openshift-ecosystem/community-operators-prod/tree/main/operators/grafana-operator](https://github.com/redhat-openshift-ecosystem/community-operators-prod/tree/main/operators/grafana-operator)
that matches the new tag.

Copy the content of `bundle/manifests/` in the grafana-operator repo from the taged version.

Update `grafana-operator/grafana-operator.package.yaml` with the new tag.

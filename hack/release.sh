#!/bin/bash
set -ex

# This script is used to generate a tagged version if the grafana-operator yaml
# This way users can easily use the pre-built kustomization file or setup there own but be able to point on these artifacts.

#TAG="v4.0.2"
TAG="latest"
if [[ $1 != "" ]]; then
  TAG=$1
fi
echo "TAG is: $TAG"

# Get kustomize
which kustomize
if [[ $? != 0 ]]; then
  make kustomize
fi

BASE_PATH=./deploy/manifests/$TAG

if [[ -f $BASE_PATH/kustomization.yaml ]]; then
    rm $BASE_PATH/kustomization.yaml
fi


mkdir -p $BASE_PATH
kustomize build ./config/crd > $BASE_PATH/crds.yaml
kustomize build ./config/manager > $BASE_PATH/deployment.yaml
kustomize build ./config/rbac > $BASE_PATH/rbac.yaml

cd $BASE_PATH
kustomize create --autodetect

#!/bin/bash
set -ex

# If run localy remember to not commit in the changes this e2e script does
# If used together with kind it assumes that you have pre-loaded kind with the image that you define.
# For more information read develop.md in our documentation.
# You can define any image you want when calling the scipt by using:
# sh hack/e2e.sh <img-name>
# I cleanup the port-forward in the end of the script but if it errors out before it will still remain, don't forget to delete it.

# TMP, should use config/install in the future
INSTALL_PATH="config/manager"
NAMESPACE="grafana-operator-system"
PATH=$PATH:$PWD/bin
HEADER='-H Accept:application/json -H Content-Type:application/json'
DEBUG_FILE="/tmp/grafana_e2e_debug.txt"

IMG=quay.io/grafana-operator/grafana-operator:latest
if [[ $1 != "" ]]; then
  IMG=$1
fi
echo $IMG

# Get kustomize
which kustomize
if [[ $? != 0 ]]; then
  make kustomize
fi

# Prepare for kind e2e test
cd $INSTALL_PATH && kustomize edit set image controller=$IMG
cd -

set +ex

# Check if imagePullPolicy is set
cat $INSTALL_PATH/kustomization.yaml |grep "/spec/template/spec/containers/0/imagePullPolicy"
if [[ $? != 0 ]]; then
cat <<EOF >> $INSTALL_PATH/kustomization.yaml

patchesJson6902:
  - target:
      version: v1
      kind: Deployment
      name: controller-manager
    patch: |-
      - op: add
        path: /spec/template/spec/containers/0/imagePullPolicy
        value: Never
EOF
fi

set -ex

# For loop for debugging
for i in {1..80}; do kubectl get all -n $NAMESPACE >> $DEBUG_FILE; echo "Output: $i "$(date) >> $DEBUG_FILE; sleep 1; done &
FORPID=$!

# Deploy the operator
kubectl apply -k config/default
sleep 5
kubectl rollout status -w --timeout=60s deployment grafana-operator-controller-manager -n $NAMESPACE

kubectl apply -f deploy/examples/Grafana.yaml -n $NAMESPACE
sleep 20
# Takes some time for the operator to create the deployment
kubectl rollout status -w --timeout=60s deployment grafana-deployment -n $NAMESPACE

# Get the admin password
PASSWORD=$(kubectl -n $NAMESPACE get secrets grafana-admin-credentials --template={{.data.GF_SECURITY_ADMIN_PASSWORD}} | base64 -d)

# Create some base dashboard & datasource
kubectl apply -f deploy/examples/dashboards/SimpleDashboard.yaml -n $NAMESPACE
kubectl apply -f deploy/examples/datasources/Prometheus.yaml -n $NAMESPACE

# Verify that the grafana dashboard exist
# This sleep is needed for the operator to apply the dashboards and the datasource
sleep 30

# port-forward
kubectl port-forward -n $NAMESPACE service/grafana-service 3000:3000 &
FPID=$!

sleep 5
curl localhost:3000/api/health
sleep 5
DASHBOARDOUTPUT=$(curl $HEADER "http://admin:$PASSWORD@localhost:3000/api/search?folderIds=0&query=&starred=false")
sleep 1
GRAFANAUID=$(echo $DASHBOARDOUTPUT |jq -r '.[0].uid')
sleep 1
GRAFANA_DASHBOARD=$(curl $HEADER "http://admin:$PASSWORD@localhost:3000/api/dashboards/uid/$GRAFANAUID")
sleep 1
FOLDER_ID=$(echo $GRAFANA_DASHBOARD |jq -r .meta.folderId)
if [[ $FOLDER_ID != 0 ]]; then
  echo "Unable to get grafana dashboard"
  exit 1
fi

# Clean up
# Delete the port-forward pid
set +ex
kill $FPID
kill $FORPID

# Don't care if the FPID don't get killed, happy with the exit anyway.
# The set +ex handels that logic.
exit 0

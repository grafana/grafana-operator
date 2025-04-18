#!/usr/bin/env bash

KUBECONFIG=${KUBECONFIG:-~/.kube/kind-grafana-operator}
CRD_NS=${CRD_NS:-grafana-crds}

set -eu

# Find the script directory
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

kubectl --kubeconfig "${KUBECONFIG}" \
    label ns default grafana=grafana

# Install ingress-nginx
echo ""
echo "###############################"
echo "# 2. Install ingress-nginx    #"
echo "###############################"
kubectl --kubeconfig="${KUBECONFIG}" \
    apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
kubectl --kubeconfig="${KUBECONFIG}" \
    -n ingress-nginx \
    wait deploy ingress-nginx-controller \
    --for condition=Available \
    --timeout=90s

# Will install the CRD:s
echo ""
echo "###############################"
echo "# 3. Install CRDs             #"
echo "###############################"
pushd "${SCRIPT_DIR}/../.."
KUBECONFIG="${KUBECONFIG}" make install
sleep 2

# Setup a grafana objects in default namespace
echo ""
echo "###############################"
echo "# 4. Install grafana objects  #"
echo "###############################"
# Create an extra namespace for CRDs
kubectl --kubeconfig "${KUBECONFIG}" \
    create ns "${CRD_NS}"
kubectl --kubeconfig "${KUBECONFIG}" \
    label ns "${CRD_NS}" grafanacrd=grafana --overwrite

kubectl --kubeconfig="${KUBECONFIG}" \
    apply -k "${SCRIPT_DIR}/resources/default/"

# Setup a grafana objects in specific ns
echo ""
echo "##########################################"
echo "# 5. Install grafana objects in ${CRD_NS}"
echo "##########################################"
kubectl -n "${CRD_NS}" --kubeconfig="${KUBECONFIG}" \
    apply -k "${SCRIPT_DIR}/resources/crd-ns/"

echo ""
echo "##########################################"
echo "# All done!"
echo "##########################################"
echo "To access the cluster instance, configure KUBECONFIG:"
echo "export KUBECONFIG=${KUBECONFIG}"
echo ""
echo "To run the operator locally against the new cluster, use 'make run'"

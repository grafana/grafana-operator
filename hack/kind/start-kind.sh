#!/usr/bin/env bash

KIND=${KIND:-kind}
KIND_CLUSTER_NAME=${KIND_CLUSTER_NAME:-kind-grafana}
KUBECONFIG=${KUBECONFIG:-~/.kube/kind-grafana-operator}
RECREATE_CLUSTER=${RECREATE_CLUSTER:-yes}
set -eu

# Find the script directory
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

# Check if named kind cluster already exists
if [[ "$($KIND get clusters)" =~ "$KIND_CLUSTER_NAME" ]]; then
    # Unless otherwise specified, delete it and create a new cluster
    if [[ "$RECREATE_CLUSTER" == "yes" ]]; then
        echo "Delete existing cluster"
        ${KIND} --kubeconfig="${KUBECONFIG}" delete cluster --name "${KIND_CLUSTER_NAME}" ||
            echo "There was no existing cluster"
    else
        exit 0
    fi
fi

NODE_VERSION_ARG=""
if [[ -n "${KIND_NODE_VERSION:-}" ]]; then
    NODE_VERSION_ARG="--image=kindest/node:v${KIND_NODE_VERSION}"
fi

# Start kind cluster
echo ""
echo "###############################"
echo "# 1. Start kind cluster       #"
echo "###############################"
${KIND} --kubeconfig "${KUBECONFIG}" create cluster \
    --name "${KIND_CLUSTER_NAME}" \
    --wait 120s \
    --config="${SCRIPT_DIR}/resources/cluster.yaml" \
    $NODE_VERSION_ARG

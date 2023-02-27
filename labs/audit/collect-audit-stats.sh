#!/usr/bin/env bash
set -euo pipefail

docker cp kind-control-plane:/var/log/kubernetes/kube-apiserver-audit.log .
docker exec kind-control-plane truncate -s 0 /var/log/kubernetes/kube-apiserver-audit.log
grep RequestReceived kube-apiserver-audit.log | grep "system:serviceaccount:grafana-operator-system:grafana-operator-controller-manager" | jq '. | { verb, requestURI } | join(" - ")' | sort | uniq -c

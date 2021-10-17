#!/bin/bash
set -x

NAMESPACE="grafana-operator-system"

kubectl delete -f deploy/examples/Grafana.yaml -n $NAMESPACE

kubectl delete -f deploy/examples/dashboards/SimpleDashboard.yaml -n $NAMESPACE
kubectl delete -f deploy/examples/datasources/Prometheus.yaml -n $NAMESPACE

sleep 2
kubectl delete deployments.apps grafana-operator-controller-manager -n $NAMESPACE

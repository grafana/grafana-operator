#!/usr/bin/env bash
KIND_CLUSTER_NAME=${KIND_CLUSTER_NAME:-kind}
KUBECONFIG=${KUBECONFIG:-~/.kube/kind-grafana-operator}

set -eu

# Make sure there is no current cluster
echo "Delete existing cluster"
kind --kubeconfig="${KUBECONFIG}" delete cluster --name "${KIND_CLUSTER_NAME}" \
  || echo "There was no existing cluster"

# Start kind cluster
echo ""
echo "###############################"
echo "# 1. Start kind cluster       #"
echo "###############################"
cat <<EOF | kind create cluster \
  --name "${KIND_CLUSTER_NAME}" \
  --wait 120s \
  --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF

kubectl label ns default grafana=grafana

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
make install
sleep 2

# Setup a grafana object
echo ""
echo "###############################"
echo "# 4. Install a grafana object #"
echo "###############################"
cat << EOF | kubectl --kubeconfig="${KUBECONFIG}" \
  apply -f -
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  name: grafana
  labels:
    dashboards: "grafana"
spec:
  config:
    log:
      mode: "console"
    auth:
      disable_login_form: "false"
    security:
      admin_user: root
      admin_password: secret
  ingress:
    spec:
      ingressClassName: nginx
      rules:
        - host: grafana.127.0.0.1.nip.io
          http:
            paths:
              - backend:
                  service:
                    name: grafana-service
                    port:
                      number: 3000
                path: /
                pathType: Prefix
EOF

# Setup a grafana dashboard in standard ns
echo ""
echo "#####################################"
echo "# 5. Install a dashboard in default #"
echo "####################################"
cat << EOF | kubectl --kubeconfig="${KUBECONFIG}" \
  apply -f -
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafana
  labels:
    dashboards: "grafana"
spec:
  folder: my-folder-name
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  json: |
    {
      "id": null,
      "title": "Simple Dashboard",
      "tags": [],
      "style": "dark",
      "timezone": "browser",
      "editable": true,
      "hideControls": false,
      "graphTooltip": 1,
      "panels": [],
      "time": {
        "from": "now-6h",
        "to": "now"
      },
      "timepicker": {
        "time_options": [],
        "refresh_intervals": []
      },
      "templating": {
        "list": []
      },
      "annotations": {
        "list": []
      },
      "refresh": "5s",
      "schemaVersion": 17,
      "version": 0,
      "links": []
    }
EOF

# Setup a grafana datasource
echo ""
echo "###################################"
echo "# 6. Install a datasource in default #"
echo "###################################"
cat << EOF | kubectl --kubeconfig="${KUBECONFIG}" \
  apply -f -
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDatasource
metadata:
  name: example-grafanadatasource
spec:
  datasource:
    access: proxy
    type: prometheus
    jsonData:
      timeInterval: 5s
      tlsSkipVerify: true
    name: Prometheus
    url: http://prometheus-service:9090
  instanceSelector:
    matchLabels:
      dashboards: grafana
  plugins:
    - name: grafana-clock-panel
      version: 1.3.0
EOF

# Create an extra namespace for CRDs
CRD_NS=grafana-crds
kubectl create ns "${CRD_NS}"
kubectl label ns "${CRD_NS}" grafanacrd=grafana --overwrite

# Setup a grafana dashboard in specific ns
echo ""
echo "############################"
echo "# 7. Install a dashboard in ${CRD_NS}"
echo "############################"
cat << EOF | kubectl -n "${CRD_NS}" --kubeconfig="${KUBECONFIG}" \
  apply -f -
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafana
  labels:
    dashboards: "grafana"
spec:
  folder: my-folder-name
  instanceSelector:
    matchLabels:
      dashboards: "grafana"
  json: |
    {
      "id": null,
      "title": "Simple Dashboard in CRD NS",
      "tags": [],
      "style": "dark",
      "timezone": "browser",
      "editable": true,
      "hideControls": false,
      "graphTooltip": 1,
      "panels": [],
      "time": {
        "from": "now-6h",
        "to": "now"
      },
      "timepicker": {
        "time_options": [],
        "refresh_intervals": []
      },
      "templating": {
        "list": []
      },
      "annotations": {
        "list": []
      },
      "refresh": "5s",
      "schemaVersion": 17,
      "version": 0,
      "links": []
    }
EOF

# Setup a grafana datasource
echo ""
echo "############################"
echo "# 8. Install a datasource in ${CRD_NS}"
echo "############################"
cat << EOF | kubectl -n "${CRD_NS}" --kubeconfig="${KUBECONFIG}" \
  apply -f -
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDatasource
metadata:
  name: example-grafanadatasource
spec:
  datasource:
    access: proxy
    type: prometheus
    jsonData:
      timeInterval: 5s
      tlsSkipVerify: true
    name: Prometheus-in-crd-nd
    url: http://prometheus-service:9090
  instanceSelector:
    matchLabels:
      dashboards: grafana
  plugins:
    - name: grafana-clock-panel
      version: 1.3.0
EOF

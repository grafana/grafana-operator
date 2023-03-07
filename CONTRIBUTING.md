# Contributing to grafana-operator

Thank you for investing your time in contributing to our project.

## Development

The operator uses unit tests and [Kuttl](https://kuttl.dev/) for e2e tests to make sure that the operator is working as intended, we use make to generate a number of docs and scripts for us.

### Local development using make run

Some of us use kind some use crc, below you can find an example on how to integrate with a kind cluster.
When adding a grafanadashboard to our grafana instances through the operator and using `make test` to run the operator we need a way to send data in to the grafana instance.

There are multiple ways of doing so but this is one of them using [kind](https://kind.sigs.k8s.io/docs/user/ingress/#create-cluster).

```shell
cat <<EOF | kind create cluster --config=-
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
```

When the kind cluster is up and running setup your ingress.

```shell
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
```

Get your kind IP.
In this example lets assume that you are using test.io as your ingress endpoint.

```shell
KIND_IP=$(docker container inspect kind-control-plane \
  --format '{{ .NetworkSettings.Networks.kind.IPAddress }}')
sudo sh -c "echo $KIND_IP test.io >> /etc/hosts"
```

To get the operator to create a dashboard through the ingress object we have added a feature in the operator.

Notice the `spec.client.preferIngress: true`
This should only be used during development.

```.yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: Grafana
metadata:
  name: grafana
  labels:
    dashboards: "grafana"
spec:
  client:
    preferIngress: true
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
        - host: test.io
          http:
            paths:
              - backend:
                  service:
                    name: grafana-service
                    port:
                      number: 3000
                path: /
                pathType: Prefix
```

This makes the grafana client in the operator to use the ingress address instead of the service name.

```shell
# Will install the CRD:s
make install
# Will run the operator from your console
make run
```

Now you should be ready to develop the operator.

### E2e tests using Kuttl

As mentioned above we use Kuttl to run e2e tests for the operator, we normally run Kuttl on [Kind](https://kind.sigs.k8s.io/)

The `make e2e` command will

```shell
# Build the container
VERSION=latest make docker-build
# Using kind load the locally built image to the kind cluster
kind load docker-image quay.io/grafana-operator/grafana-operator:v5.0.0
# Create grafana-operator-system namespace
kubectl create ns grafana-operator-system
# Run the Kuttl tests
VERSION=latest make e2e
```

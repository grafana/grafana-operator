# Contributing to grafana-operator

Thank you for investing your time in contributing to our project.

## Development

The operator uses unit tests and [Kuttl](https://kuttl.dev/) for e2e tests to make sure that the operator is working as intended, we use make to generate a number of docs and scripts for us.

The operator use a submodule for [grafonnet-lib](https://github.com/grafana/grafonnet-lib),
one of the first things you have to do is to run `make submodule`.

### Code standards

We use a number of code standards in the project that we apply using a number of different tools.
As a part of the CI solution these settings will be validated, but all of them can be tested using the Makefile before pushing.

- [golanci-lint](https://golangci-lint.run/)
- [gofumpt](https://github.com/mvdan/gofumpt)

Before pushing any code we recommend that you run the following make commands.

```shell
make submodule
make test
make code/golangci-lint
```

Depending on what you have changed these commands will update a number of different files.

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

To get the operator to create a dashboard through the ingress object we have added a feature in the operator.

Notice the `spec.client.preferIngress: true`
This should only be used during development.

In this example we are using [nip.io](https://nip.io/) which will steer traffic to your local deployment through a DNS response (e.g. `nslookup grafana.127.0.0.1.nip.io` will respond with `127.0.0.1`).

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
kind load docker-image ghcr.io/grafana-operator/grafana-operator:v5.0.0
# Create grafana-operator-system namespace
kubectl create ns grafana-operator-system
# Run the Kuttl tests
VERSION=latest make e2e
```

### Helm

We support helm as a deployment solution and it can be found under [deploy/helm](deploy/helm/grafana-operator/README.md).

The grafana-operator helm chart is currently manually created.
When CRD:s is upgraded the helm chart will also get an update.

But if you generate new RBAC rules or create new deployment options for the operator you will need to add them manually.

Chart.yaml `appVersion` follows the grafana-operator version but the helm chart is versioned separately.

If you add update the chart don't forget to run `make helm-docs`, which will update the helm specific README file.

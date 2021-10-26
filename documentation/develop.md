# Develop

The grafana-operator is currently built on operator-sdk version
[1.3.0](https://github.com/operator-framework/operator-sdk/releases/tag/v1.3.0).

To our knowledge the grafana-operator works on all kubernetes deployments,
for local development many of us use is [kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
or [crc](https://developers.redhat.com/products/codeready-containers/overview).

## Read the Makefile

We strongly recommend that you read through the Makefile,
we are heavily relying on make commands to help you getting started quicker.

## Local deployment using go

```shell
make generate
make install
# Set environment variables, adjust the WATCH_NAMESPACE to the namespace you want to watch.
# This variable have to be set
export WATCH_NAMESPACE=default
make run
```

You can of course run the deployment using a debugger or similar tools.

## Local deployment using docker

If you want a solution closer to reality you can instead build your own containers.
We will go through two sections.

- Using kind load docker-image.
- Using your own quay repo to upload a manually built image.

### Kind load docker-image

This solution assumes that you are using kind in your development environment.

```shell
make generate
make install
make docker-build
```

We will pre-load the container image to kind. To make sure that we only use our locally built container
edit the kustomize file to never pull the image from a external source.

Remember to not commit these changes.

```shell
cat <<EOF >> config/manager/kustomization.yaml

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
```

Load the image in to kind and start the deployment.

```shell
kind load docker-image quay.io/integreatly/grafana-operator:latest
make deploy
```

### Remote repo

If you want to build and upload your container image to your own remote repo you can follow these instructions.

```shell
make generate
make install

# Login to remote repo
export QUAY_USER=username1
export QUAY_PASSWD=super-secret-password
# In this example we use quay.io, you can use any provider you see fit.
# This is one way of many on how to login using docker, perform the one that works for you.
echo $QUAY_PASSWD | docker login -u $QUAY_USER --password-stdin quay.io

# If you don't want to add the IMG= all the time you can also edit the IMG variable in the Makefile
make docker-build IMG=quay.io/$QUAY_USER/grafana-operator:latest
make docker-push IMG=quay.io/$QUAY_USER/grafana-operator:latest
make deploy IMG=quay.io/$QUAY_USER/grafana-operator:latest
```

## e2e script

Running the e2e script locally assumes that you have made the container image available to your cluster.
For example if you are using kind it should be pre-loaded.

It assume that you are not running any other grafana operator instance for example through go.

To run it:

```hack/e2e.sh
sh hack/e2e.sh
```

If you want to clean-up a few of the resources that hack/e2e.sh creates use clean_e2e.sh.
It will remove the grafana instances and operator but it won't delete the port-forward.
It will also remove the debug output file /tmp/grafana_e2e_debug.txt after reading the file.

```hack/clean_e2e.sh
sh hack/clean_e2e.sh
# To kill the potentially remaining port-forward to the grafana service:
kill $(lsof -t -i:3000)
```

ORG=pb82
NAMESPACE=application-monitoring
PROJECT=grafana-operator
REG=docker.io
SHELL=/bin/bash
TAG=latest
PKG=github.com/integr8ly/grafana-operator
TEST_DIRS?=$(shell sh -c "find $(TOP_SRC_DIRS) -name \\*_test.go -exec dirname {} \\; | sort | uniq")
TEST_POD_NAME=grafana-operator-test
COMPILE_TARGET=./tmp/_output/bin/$(PROJECT)

.PHONY: setup/dep
setup/dep:
	@echo Installing dep
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
	@echo setup complete

.PHONY: setup/travis
setup/travis:
	@echo Installing Operator SDK
	@curl -Lo operator-sdk https://github.com/operator-framework/operator-sdk/releases/download/v0.2.1/operator-sdk-v0.2.1-x86_64-linux-gnu && chmod +x operator-sdk && sudo mv operator-sdk /usr/local/bin/

.PHONY: code/run
code/run:
	@operator-sdk up local --namespace=${NAMESPACE}

.PHONY: code/compile
code/compile:
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o=$(COMPILE_TARGET) ./cmd/manager

.PHONY: code/gen
code/gen:
	operator-sdk generate k8s

.PHONY: code/check
code/check:
	@diff -u <(echo -n) <(gofmt -d `find . -type f -name '*.go' -not -path "./vendor/*"`)

.PHONY: code/fix
code/fix:
	@gofmt -w `find . -type f -name '*.go' -not -path "./vendor/*"`

.PHONY: image/build
image/build: code/compile
	@operator-sdk build ${REG}/${ORG}/${PROJECT}:${TAG}

.PHONY: image/push
image/push:
	docker push ${REG}/${ORG}/${PROJECT}:${TAG}

.PHONY: image/build/push
image/build/push: image/build image/push

.PHONY: image/build/test
image/build/test:
	operator-sdk build --enable-tests docker.io/${ORG}/${PROJECT}:${TAG}

.PHONY: test/unit
test/unit:
	@echo Running tests:
	go test -v -race -cover ./pkg/...

.PHONY: test/e2e
test/e2e:
	kubectl apply -f deploy/test-e2e-pod.yaml -n ${PROJECT}
	${SHELL} ./scripts/stream-pod ${TEST_POD_NAME} ${PROJECT}

.PHONY: cluster/prepare
cluster/prepare:
	-kubectl apply -f deploy/crds/
	-oc new-project $(NAMESPACE)
	-kubectl create --insecure-skip-tls-verify -f deploy/rbac.yaml -n $(NAMESPACE)

.PHONY: cluster/clean
cluster/clean:
	-kubectl delete role grafana-operator -n $(NAMESPACE)
	-kubectl delete rolebinding grafana-operator -n $(NAMESPACE)
	-kubectl delete crd grafanas.integreatly.org
	-kubectl delete namespace $(NAMESPACE)

.PHONY: cluster/create/examples
cluster/create/examples:
		-kubectl create -f deploy/examples/Grafana.yaml -n $(NAMESPACE)

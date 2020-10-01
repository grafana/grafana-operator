ORG?=integreatly
NAMESPACE?=grafana
PROJECT=grafana-operator
REG?=quay.io
SHELL=/bin/bash
TAG?=latest
PKG=github.com/integr8ly/grafana-operator
COMPILE_TARGET=./tmp/_output/bin/$(PROJECT)

.PHONY: setup/travis
setup/travis:
	@echo Installing Operator SDK
	@curl -Lo operator-sdk https://github.com/operator-framework/operator-sdk/releases/download/v0.12.0/operator-sdk-v0.12.0-x86_64-linux-gnu && chmod +x operator-sdk && sudo mv operator-sdk /usr/local/bin/

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
	@diff -u <(echo -n) <(gofmt -d .)

.PHONY: code/fix
code/fix:
	@gofmt -w .

.PHONY: image/build
image/build: code/compile
	@operator-sdk build ${REG}/${ORG}/${PROJECT}:${TAG}

.PHONY: image/push
image/push:
	docker push ${REG}/${ORG}/${PROJECT}:${TAG}

.PHONY: image/build/push
image/build/push: image/build image/push

.PHONY: test/unit
test/unit:
	@echo Running tests:
	go test -v -race -cover ./pkg/...

.PHONY: test/e2e
test/e2e:
	@operator-sdk --verbose test local ./test/e2e --watch-namespace="grafana-test-e2e" --operator-namespace="grafana-test-e2e" --debug --up-local

.PHONY: cluster/prepare/local/file
cluster/prepare/local/file:
	@sed -i "s/__NAMESPACE__/${NAMESPACE}/g" deploy/cluster_roles/cluster_role_binding_grafana_operator.yaml

.PHONY: cluster/prepare/local
cluster/prepare/local: cluster/prepare/local/file
	-kubectl create namespace ${NAMESPACE}
	kubectl apply -f deploy/crds
	kubectl apply -f deploy/roles -n ${NAMESPACE}
	kubectl apply -f deploy/cluster_roles
	kubectl apply -f deploy/examples/Grafana.yaml -n ${NAMESPACE}

.PHONY: cluster/cleanup
cluster/cleanup: operator/stop
	-kubectl delete deployment grafana-deployment -n ${NAMESPACE}
	-kubectl delete namespace ${NAMESPACE}

## Deploy the latest tagged release
.PHONY: operator/deploy
operator/deploy: cluster/prepare/local
	kubectl apply -f deploy/operator.yaml -n ${NAMESPACE}
	@git checkout -- deploy/cluster_roles/cluster_role_binding_grafana_operator.yaml

## Deploy the latest master image
.PHONY: operator/deploy/master
operator/deploy/master: cluster/prepare/local
	kubectl apply -f deploy/operatorMasterImage.yaml -n ${NAMESPACE}
	@git checkout -- deploy/cluster_roles/cluster_role_binding_grafana_operator.yaml

.PHONY: operator/stop
operator/stop:
	-kubectl delete deployment grafana-operator -n ${NAMESPACE}

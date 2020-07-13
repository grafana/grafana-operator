ORG?=integreatly
NAMESPACE=grafana
PROJECT=grafana-operator
REG?=quay.io
SHELL=/bin/bash
TAG?=latest
PKG=github.com/integr8ly/grafana-operator
COMPILE_TARGET=./tmp/_output/bin/$(PROJECT)

.PHONY: setup/dep
setup/dep:
	@echo Installing dep
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
	@echo setup complete

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

.PHONY: test/unit
test/unit:
	@echo Running tests:
	go test -v -race -cover ./pkg/...

.PHONY: minikube/prepare/local
minikube/prepare/local:
	-kubectl create namespace ${NAMESPACE}
	kubectl apply -f deploy/crds
	kubectl apply -f deploy/roles -n ${NAMESPACE}
	kubectl apply -f deploy/cluster_roles
	kubectl apply -f deploy/examples/Grafana.yaml -n ${NAMESPACE}


.PHONY: minikube/deploy
minikube/deploy: minikube/prepare/local
	kubectl apply -f deploy/operator.yaml -n ${NAMESPACE}

.PHONY: minikube/stop
minikube/stop:
	-kubectl delete deployment grafana-operator -n ${NAMESPACE}

.PHONY: minikube/cleanup
minikube/cleanup: minikube/stop
	minikube stop
	minikube delete
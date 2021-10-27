# Current Operator version
VERSION ?= 4.0.1
# Default bundle image tag
BUNDLE_IMG ?= quay.io/grafana-operator/controller-bundle:v$(VERSION)
# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# Image URL to use all building/pushing image targets8
IMG ?= quay.io/grafana-operator/grafana-operator:v$(VERSION)
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
test: generate fmt vet manifests api-docs
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.7.0/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Get submodules
submodule:
	git submodule update --init --recursive

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

# UnDeploy controller from the configured Kubernetes cluster in ~/.kube/config
undeploy:
	$(KUSTOMIZE) build config/default | kubectl delete -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Generate API reference documentation
api-docs: gen-crd-api-reference-docs kustomize
	@{ \
	set -e ;\
	TMP_DIR=$$(mktemp -d) ; \
	$(KUSTOMIZE) build config/crd -o $$TMP_DIR/crd-output.yaml ;\
	$(API_REF_GEN) crdoc --resources $$TMP_DIR/crd-output.yaml --output documentation/api.md ;\
	}

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build a single-architecture docker image
docker-build: test
	DOCKER_BUILDKIT=1 docker build -t ${IMG} .

# Push the single-architecture docker image
docker-push:
	docker push ${IMG}

# Build and push a multi-architecture docker image
docker-buildx: test
	docker buildx build --platform linux/amd64,linux/arm64,linux/s390x --push -t ${IMG} .

# Download controller-gen locally if necessary
CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen:
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1)

# Download kustomize locally if necessary
KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize:
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

# Generate bundle manifests and metadata, then validate generated files.
.PHONY: bundle
bundle: manifests kustomize
	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	operator-sdk bundle validate ./bundle

# Build the bundle image.
.PHONY: bundle-build
bundle-build:
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .


.PHONY: code/check
code/check: fmt vet
	golint ./...

.PHONY: code/golangci-lint
code/golangci-lint:
	golangci-lint run ./...

.PHONY: cluster/prepare/local/file
cluster/prepare/local/file:
	@sed -i "s/__NAMESPACE__/${NAMESPACE}/g" deploy/cluster_roles/cluster_role_binding_grafana_operator.yaml

.PHONY: cluster/prepare/local
cluster/prepare/local: cluster/prepare/local/file
	-kubectl create namespace ${NAMESPACE}
	kubectl apply -f deploy/roles -n ${NAMESPACE}
	kubectl apply -f deploy/cluster_roles
	kubectl apply -f deploy/examples/Grafana.yaml -n ${NAMESPACE}

# Find or download gen-crd-api-reference-docs
gen-crd-api-reference-docs:
ifeq (, $(shell which crdoc))
	@{ \
	set -e ;\
	API_REF_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$API_REF_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get fybrik.io/crdoc@v0.5.2 ;\
	rm -rf $$API_REF_GEN_TMP_DIR ;\
	}
API_REF_GEN=$(GOBIN)/crdoc
else
API_REF_GEN=$(shell which crdoc)
endif

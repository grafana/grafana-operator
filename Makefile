# Current Operator version
VERSION ?= 5.15.1

# BUNDLE_GEN_FLAGS are the flags passed to the operator-sdk generate bundle command
BUNDLE_GEN_FLAGS ?= -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)

# USE_IMAGE_DIGESTS defines if images are resolved via tags or digests
# You can enable this value if you would like to use SHA Based Digests
# To enable set flag to true
USE_IMAGE_DIGESTS ?= false
ifeq ($(USE_IMAGE_DIGESTS), true)
	BUNDLE_GEN_FLAGS += --use-image-digests
endif

# Read Grafana Image and Version from go code
GRAFANA_IMAGE := $(shell grep 'GrafanaImage' controllers/config/operator_constants.go | sed 's/.*"\(.*\)".*/\1/')
GRAFANA_VERSION := $(shell grep 'GrafanaVersion' controllers/config/operator_constants.go | sed 's/.*"\(.*\)".*/\1/')

# Image URL to use all building/pushing image targets
REGISTRY ?= ghcr.io
ORG ?= grafana
IMG ?= $(REGISTRY)/$(ORG)/grafana-operator:v$(VERSION)
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.28.0

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: manifests test api-docs helm/docs

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: yq controller-gen kustomize ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role webhook paths="./..." crd output:crd:artifacts:config=config/crd/bases
	$(CONTROLLER_GEN) rbac:roleName=manager-role webhook paths="./..." crd output:crd:artifacts:config=deploy/helm/grafana-operator/crds
	$(YQ) -i '(select(.kind == "Deployment") | .spec.template.spec.containers[0].env[] | select (.name == "RELATED_IMAGE_GRAFANA")).value="$(GRAFANA_IMAGE):$(GRAFANA_VERSION)"' config/manager/manager.yaml

	# NOTE: As we publish the whole kustomize folder structure (deploy/kustomize) as an OCI arfifact via flux, in kustomization.yaml, we cannot reference files that reside outside of deploy/kustomize. Thus, we need to maintain an additional copy of CRDs and the ClusterRole
	$(KUSTOMIZE) build config/crd -o deploy/kustomize/base/crds.yaml
	cp config/rbac/role.yaml deploy/kustomize/base/role.yaml

	# Sync role definitions to helm chart
	mkdir -p deploy/helm/grafana-operator/files
	cat config/rbac/role.yaml | $(YQ) -r 'del(.rules[] | select (.apiGroups | contains(["route.openshift.io"])))' > deploy/helm/grafana-operator/files/rbac.yaml
	cat config/rbac/role.yaml | $(YQ) -r 'del(.rules[] | select (.apiGroups | contains(["route.openshift.io"]) | not))'  > deploy/helm/grafana-operator/files/rbac-openshift.yaml

# Generate API reference documentation
api-docs: manifests gen-crd-api-reference-docs
	$(API_REF_GEN) crdoc --resources config/crd/bases --output docs/docs/api.md --template hugo/templates/frontmatter-grafana-operator.tmpl

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate code/gofumpt code/golangci-lint api-docs vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile cover.out

##@ Build

.PHONY: build
build: generate code/gofumpt vet ## Build manager binary.
	go build -o bin/manager main.go

.PHONY: run
run: manifests generate code/gofumpt vet ## Run a controller from your host.
	go run ./main.go --zap-devel=true

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl replace --force=true -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply --server-side --force-conflicts -f -

.PHONY: deploy-chainsaw
deploy-chainsaw: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/chainsaw-overlay | kubectl apply --server-side --force-conflicts -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: start-kind
start-kind: kind ## Start kind cluster locally
	@hack/kind/start-kind.sh

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
OPERATOR_SDK ?= $(LOCALBIN)/operator-sdk
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
YQ = $(LOCALBIN)/yq
KIND = $(LOCALBIN)/kind

## Tool Versions
# Set the Operator SDK version to use. By default, what is installed on the system is used.
# This is useful for CI or a project to utilize a specific version of the operator-sdk toolkit.
OPERATOR_SDK_VERSION ?= v1.32.0
KUSTOMIZE_VERSION ?= v5.1.1
CONTROLLER_TOOLS_VERSION ?= v0.16.3
OPM_VERSION ?= v1.23.2
YQ_VERSION ?= v4.35.2
KO_VERSION ?= v0.16.0
KIND_VERSION ?= v0.24.0
CHAINSAW_VERSION ?= v0.2.10

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || { curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: operator-sdk
operator-sdk: ## Download operator-sdk locally if necessary.
ifeq (,$(wildcard $(OPERATOR_SDK)))
ifeq (, $(shell which operator-sdk 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPERATOR_SDK)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPERATOR_SDK) https://github.com/operator-framework/operator-sdk/releases/download/$(OPERATOR_SDK_VERSION)/operator-sdk_$${OS}_$${ARCH} ;\
	chmod +x $(OPERATOR_SDK) ;\
	}
else
OPERATOR_SDK = $(shell which operator-sdk)
endif
endif

.PHONY: yq
yq: ## Download yq locally if necessary.
ifeq (,$(wildcard $(YQ)))
ifeq (,$(shell which yq 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(YQ)) ;\
	OSTYPE=$(shell uname | awk '{print tolower($$0)}') && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(YQ) https://github.com/mikefarah/yq/releases/download/$(YQ_VERSION)/yq_$${OSTYPE}_$${ARCH} ;\
	chmod +x $(YQ) ;\
	}
else
YQ = $(shell which yq)
endif
endif

.PHONY: kind
kind: ## Download kind locally if necessary.
ifeq (,$(wildcard $(KIND)))
ifeq (,$(shell which kind 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(KIND)) ;\
	OSTYPE=$(shell uname | awk '{print tolower($$0)}') && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(KIND) https://github.com/kubernetes-sigs/kind/releases/download/$(KIND_VERSION)/kind-$${OSTYPE}-$${ARCH} ;\
	chmod +x $(KIND) ;\
	}
else
KIND = $(shell which kind)
endif
endif

# CHANNELS define the bundle channels used in the bundle.
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "candidate,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=candidate,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="candidate,fast,stable")
CHANNELS=v5
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif

# DEFAULT_CHANNEL defines the default channel used in the bundle.
# Add a new line here if you would like to change its default config. (E.g DEFAULT_CHANNEL = "stable")
# To re-generate a bundle for any other default channel without changing the default setup, you can:
# - use the DEFAULT_CHANNEL as arg of the bundle target (e.g make bundle DEFAULT_CHANNEL=stable)
# - use environment variables to overwrite this value (e.g export DEFAULT_CHANNEL="stable")
DEFAULT_CHANNEL="v5"
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

.PHONY: bundle
bundle: manifests kustomize operator-sdk ## Generate bundle manifests and metadata, then validate generated files.
	$(OPERATOR_SDK) generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle $(BUNDLE_GEN_FLAGS)
	./hack/add-openshift-annotations.sh
	$(OPERATOR_SDK) bundle validate ./bundle

.PHONY: bundle/redhat
bundle/redhat: BUNDLE_GEN_FLAGS += --use-image-digests
bundle/redhat: bundle

# e2e
.PHONY: e2e-kind
e2e-kind:
ifeq (,$(shell kind get clusters ))
	$(KIND) --kubeconfig="${KUBECONFIG}" create cluster --image=kindest/node:v$(ENVTEST_K8S_VERSION) --config tests/e2e/kind.yaml
endif

.PHONY: e2e-local-gh-actions
e2e-local-gh-actions: e2e-kind ko-build-kind e2e

.PHONY: e2e
e2e: chainsaw install deploy-chainsaw ## Run e2e tests using chainsaw.
	$(CHAINSAW) test --test-dir ./tests/e2e/$(TESTS)

# Find or download chainsaw
chainsaw:
ifeq (, $(shell which chainsaw))
	@{ \
	set -e ;\
	go install github.com/kyverno/chainsaw@$(CHAINSAW_VERSION) ;\
	}
CHAINSAW=$(GOBIN)/chainsaw
else
CHAINSAW=$(shell which chainsaw)
endif

golangci:
ifeq (, $(shell which golangci-lint))
	@{ \
	set -e ;\
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.0 ;\
	}
GOLANGCI=$(GOBIN)/golangci-lint
else
GOLANGCI=$(shell which golangci-lint)
endif

.PHONY: code/golangci-lint
code/golangci-lint: golangci
	$(GOLANGCI) run --allow-parallel-runners ./...

gofumpt:
ifeq (, $(shell which gofumpt))
	@{ \
	set -e ;\
	go install mvdan.cc/gofumpt@v0.6.0 ;\
	}
GOFUMPT=$(GOBIN)/gofumpt
else
GOFUMPT=$(shell which gofumpt)
endif

.PHONY: code/gofumpt
code/gofumpt: gofumpt
	$(GOFUMPT) -l -w .

ko:
ifeq (, $(shell which ko))
	@{ \
	set -e ;\
	go install github.com/google/ko@${KO_VERSION} ;\
	}
KO=$(GOBIN)/ko
else
KO=$(shell which ko)
endif

export KO_DOCKER_REPO ?= ko.local/grafana/grafana-operator
export KIND_CLUSTER_NAME ?= kind-grafana
export KUBECONFIG        ?= ${HOME}/.kube/kind-grafana-operator

.PHONY: ko-build-local
ko-build-local: ko ## Build Docker image with KO
	$(KO) build --sbom=none --bare

.PHONY: ko-build-kind
ko-build-kind: ko-build-local ## Build and Load Docker image into kind cluster
	$(KIND) load docker-image $(KO_DOCKER_REPO) --name $(KIND_CLUSTER_NAME)

helm-docs:
ifeq (, $(shell which helm-docs))
	@{ \
	set -e ;\
	go install github.com/norwoodj/helm-docs/cmd/helm-docs@v1.11.0 ;\
	}
HELM_DOCS=$(GOBIN)/helm-docs
else
HELM_DOCS=$(shell which helm-docs)
endif

.PHONY: helm/docs
helm/docs: helm-docs
	$(HELM_DOCS)

BUNDLE_IMG ?= $(REGISTRY)/$(ORG)/grafana-operator-bundle:v$(VERSION)

.PHONY: bundle-build
bundle-build: ## Build the bundle image.
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: bundle-push
bundle-push: ## Push the bundle image.
	docker push $(BUNDLE_IMG)

.PHONY: opm
OPM = ./bin/opm
opm: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/$(OPM_VERSION)/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}
else
OPM = $(shell which opm)
endif
endif

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0,example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMGS ?= $(BUNDLE_IMG)

# The image tag given to the resulting catalog image (e.g. make catalog-build CATALOG_IMG=example.com/operator-catalog:v0.2.0).
CATALOG_IMG ?= $(REGISTRY)/$(REPO)/grafana-operator-catalog:v$(VERSION)

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

# Build a catalog image by adding bundle images to an empty catalog using the operator package manager tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: catalog-build
catalog-build: opm ## Build a catalog image.
	$(OPM) index add --container-tool docker --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)

# Push the catalog image.
.PHONY: catalog-push
catalog-push: ## Push a catalog image.
	docker push $(CATALOG_IMG)

# Find or download gen-crd-api-reference-docs
gen-crd-api-reference-docs:
ifeq (, $(shell which crdoc))
	@{ \
	set -e ;\
	API_REF_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$API_REF_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go install fybrik.io/crdoc@ad5ba1e62f8db46cb5a9282dfedfc3d8f3d45065 ;\
	rm -rf $$API_REF_GEN_TMP_DIR ;\
	}
API_REF_GEN=$(GOBIN)/crdoc
else
API_REF_GEN=$(shell which crdoc)
endif

.PHONY: prep-release
prep-release: yq
	$(YQ) -i '.version="v$(VERSION)"' deploy/helm/grafana-operator/Chart.yaml
	$(YQ) -i '.appVersion="v$(VERSION)"' deploy/helm/grafana-operator/Chart.yaml
	$(YQ) -i '.params.version="v$(VERSION)"' hugo/config.yaml
	sed -i 's/--version v5.*/--version v$(VERSION)/g' README.md
	sed -i 's/VERSION ?= 5.*/VERSION ?= $(VERSION)/g' Makefile
	make helm/docs

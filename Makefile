include Toolchain.mk

.DEFAULT_GOAL := all

# Current Operator version
VERSION ?= 5.20.0

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
ENVTEST_K8S_VERSION = 1.34.0

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
all: manifests test api-docs helm-docs

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
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: $(CONTROLLER_GEN) $(KUSTOMIZE) $(YQ) ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(info $(M) running $@)
	$(CONTROLLER_GEN) rbac:roleName=manager-role webhook paths="./..." crd output:crd:artifacts:config=config/crd/bases
	$(CONTROLLER_GEN) rbac:roleName=manager-role webhook paths="./..." crd output:crd:artifacts:config=deploy/helm/grafana-operator/files/crds
	# Remove CRD descriptions under Grafana#.spec.deployment
	$(YQ) -i '.spec.versions[] |= del(.schema.openAPIV3Schema.properties.spec.properties.deployment.properties | .. | select(has("description")).description)' config/crd/bases/grafana.integreatly.org_grafanas.yaml
	$(YQ) -i '.spec.versions[] |= del(.schema.openAPIV3Schema.properties.spec.properties.deployment.properties | .. | select(has("description")).description)' deploy/helm/grafana-operator/files/crds/grafana.integreatly.org_grafanas.yaml
	$(YQ) -i '.spec.versions[0].schema.openAPIV3Schema.properties.spec.properties.version.description += "\ndefault: $(GRAFANA_VERSION)"' config/crd/bases/grafana.integreatly.org_grafanas.yaml
	$(YQ) -i '.spec.versions[0].schema.openAPIV3Schema.properties.spec.properties.version.description += "\ndefault: $(GRAFANA_VERSION)"' deploy/helm/grafana-operator/files/crds/grafana.integreatly.org_grafanas.yaml

	$(YQ) -i '(select(.kind == "Deployment") | .spec.template.spec.containers[0].env[] | select (.name == "RELATED_IMAGE_GRAFANA")).value="$(GRAFANA_IMAGE):$(GRAFANA_VERSION)"' config/manager/manager.yaml

	@# NOTE: As we publish the whole kustomize folder structure (deploy/kustomize) as an OCI arfifact via flux, in kustomization.yaml, we cannot reference files that reside outside of deploy/kustomize. Thus, we need to maintain an additional copy of CRDs and the ClusterRole
	$(KUSTOMIZE) build config/crd -o deploy/kustomize/base/crds.yaml
	cp config/rbac/role.yaml deploy/kustomize/base/role.yaml

	@# Sync role definitions to helm chart
	mkdir -p deploy/helm/grafana-operator/files
	cat config/rbac/role.yaml | $(YQ) -r 'del(.rules[] | select (.apiGroups | contains(["route.openshift.io"])))' > deploy/helm/grafana-operator/files/rbac.yaml
	cat config/rbac/role.yaml | $(YQ) -r 'del(.rules[] | select (.apiGroups | contains(["route.openshift.io"]) | not))'  > deploy/helm/grafana-operator/files/rbac-openshift.yaml

# Generate API reference documentation
.PHONY: api-docs
api-docs: $(CRDOC) manifests
	$(info $(M) running $@)
	$(CRDOC) --resources config/crd/bases --output docs/docs/api.md --template hugo/templates/frontmatter-grafana-operator.tmpl

.PHONY: generate
generate: $(CONTROLLER_GEN) ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(info $(M) running $@)
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Run golangci-lint checks.
	$(info $(M) running $@)
	$(GOLANGCI_LINT) config verify
	$(GOLANGCI_LINT) fmt ./...
	$(GOLANGCI_LINT) run --allow-parallel-runners ./...

.PHONY: helm-docs
helm-docs: $(HELM_DOCS) ## Generate helm docs.
	$(info $(M) running $@)
	$(HELM_DOCS) deploy/helm

.PHONY: helm-lint
helm-lint: $(HELM) ## Validate helm chart.
	$(info $(M) running $@)
	$(HELM) template deploy/helm/grafana-operator/ > /dev/null
	$(HELM) lint deploy/helm/grafana-operator/

.PHONY: hugo
hugo: $(DART_SASS) $(HUGO) ## Prepare production build for hugo docs.
	$(info $(M) running $@)
	@echo -- Checking presence of dart-sass
	@cd hugo && $(HUGO) env | grep dart-sass
	@echo -- Building artifacts
	@cd hugo && HUGO_ENVIRONMENT=production HUGO_ENV=production $(HUGO) --gc --minify

.PHONY: hugo-dev
hugo-dev: $(DART_SASS) $(HUGO) ## Start development server for hugo.
	$(info $(M) running $@)
	@echo -- Checking presence of dart-sass
	@cd hugo && $(HUGO) env | grep dart-sass
	@echo -- Starting dev server
	@cd hugo && $(HUGO) env && $(HUGO) server --baseURL http://127.0.0.1/

.PHONY: kustomize-lint
kustomize-lint: $(KUSTOMIZE) ## Lint kustomize overlays.
	$(info $(M) running $@)
	@for d in deploy/kustomize/overlays/*/ ; do \
		$(KUSTOMIZE) build "$${d}" --load-restrictor LoadRestrictionsNone > /dev/null ;\
	done

.PHONY: kustomize-set-image
kustomize-set-image: $(KUSTOMIZE) ## Sets release image.
	$(info $(M) running $@)
	cd deploy/kustomize/base && $(KUSTOMIZE) edit set image ghcr.io/${GITHUB_REPOSITORY}=${GHCR_REPO}:${RELEASE_NAME} && cd -

.PHONY: kustomize-github-assets
kustomize-github-assets: $(KUSTOMIZE) ## Generates GitHub assets.
	$(info $(M) running $@)
	@for d in deploy/kustomize/overlays/*/ ; do \
		echo "$${d}" ;\
		$(KUSTOMIZE) build "$${d}" --load-restrictor LoadRestrictionsNone > kustomize-$$(basename "$${d}").yaml ;\
	done
	$(KUSTOMIZE) build config/crd > crds.yaml

.PHONY:
muffet-dev: $(MUFFET) ## Detect broken internal links in docs.
	$(MUFFET) --include=http://localhost:1313 http://localhost:1313

.PHONY:
test-image-pre-pull: ## Pre-pulls Grafana image used in tests to speed up CI
	docker pull $(GRAFANA_IMAGE):$(GRAFANA_VERSION) > /dev/null 2>&1 &

.PHONY: test
test: $(ENVTEST) manifests generate vet golangci-lint api-docs kustomize-lint helm-docs helm-lint ## Run tests.
	$(info $(M) running $@)
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(BIN) -p path)" go test ./... -coverprofile cover.out

.PHONY: test-short
test-short: ## Skips slow integration tests
	$(info $(M) running $@)
	go test ./... -short -coverprofile cover.out

.PHONY: vet
vet: ## Run go vet against code.
	$(info $(M) running $@)
	go vet ./...

##@ Build

.PHONY: build
build: $(GOLANGCI_LINT) generate vet ## Build manager binary.
	$(info $(M) running $@)
	$(GOLANGCI_LINT) fmt ./...
	go build -o bin/manager main.go

.PHONY: run
run: $(GOLANGCI_LINT) manifests generate vet ## Run a controller from your host.
	$(info $(M) running $@)
	$(GOLANGCI_LINT) fmt ./...
	go run ./main.go --zap-devel=true

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: $(KUSTOMIZE) manifests ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(info $(M) running $@)
	$(KUSTOMIZE) build config/crd | kubectl replace --force=true -f -

.PHONY: uninstall
uninstall: $(KUSTOMIZE) manifests ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(info $(M) running $@)
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: $(KUSTOMIZE) manifests ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	$(info $(M) running $@)
	cd deploy/kustomize/overlays/cluster_scoped && $(KUSTOMIZE) edit set image ghcr.io/grafana/grafana-operator=${IMG}
	$(KUSTOMIZE) build deploy/kustomize/overlays/cluster_scoped | kubectl apply --server-side --force-conflicts -f -

.PHONY: deploy-chainsaw
deploy-chainsaw: $(KUSTOMIZE) manifests ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	$(info $(M) running $@)
	$(KUSTOMIZE) build deploy/kustomize/overlays/chainsaw | kubectl apply --server-side --force-conflicts -f -

.PHONY: deploy-chainsaw-debug
deploy-chainsaw-debug: $(KUSTOMIZE) manifests ## Deploy debug controller (http-echo) to the K8s cluster for mirrord (https://github.com/metalbear-co/mirrord) debugging.
	$(info $(M) running $@)
	$(KUSTOMIZE) build deploy/kustomize/overlays/chainsaw-debug | kubectl apply --server-side --force-conflicts -f -

.PHONY: undeploy
undeploy: $(KUSTOMIZE) ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(info $(M) running $@)
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: start-kind
start-kind: $(KIND) ## Start kind cluster locally
	$(info $(M) running $@)
	@KIND=$(KIND) hack/kind/start-kind.sh
	@KIND=$(KIND) hack/kind/populate-kind-cluster.sh

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
bundle: $(KUSTOMIZE) $(OPERATOR_SDK) manifests ## Generate bundle manifests and metadata, then validate generated files.
	$(info $(M) running $@)
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle $(BUNDLE_GEN_FLAGS)
	./hack/add-openshift-annotations.sh
	$(OPERATOR_SDK) bundle validate ./bundle

.PHONY: bundle/redhat
bundle/redhat: BUNDLE_GEN_FLAGS += --use-image-digests
bundle/redhat: bundle

# e2e
.PHONY: e2e-kind
e2e-kind: $(KIND)
	$(info $(M) running $@)
	KIND=$(KIND) hack/kind/start-kind.sh

.PHONY: e2e-local-gh-actions
e2e-local-gh-actions: e2e-kind ko-build-kind e2e

.PHONY: e2e
e2e: $(CHAINSAW) install deploy-chainsaw ## Run e2e tests using chainsaw.
	$(info $(M) running $@)
	$(CHAINSAW) test --test-dir ./tests/e2e/$(TESTS)

export KO_DOCKER_REPO ?= ko.local/grafana/grafana-operator
export KIND_CLUSTER_NAME ?= kind-grafana
export KUBECONFIG        ?= ${HOME}/.kube/kind-grafana-operator

.PHONY: ko-build-local
ko-build-local: $(KO) ## Build Docker image with KO
	$(info $(M) running $@)
	$(KO) build --sbom=none --bare

.PHONY: ko-build-kind
ko-build-kind: $(KIND) ko-build-local ## Build and Load Docker image into kind cluster
	$(info $(M) running $@)
	$(KIND) load docker-image $(KO_DOCKER_REPO) --name $(KIND_CLUSTER_NAME)

BUNDLE_IMG ?= $(REGISTRY)/$(ORG)/grafana-operator-bundle:v$(VERSION)

.PHONY: bundle-build
bundle-build: ## Build the bundle image.
	$(info $(M) running $@)
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: bundle-push
bundle-push: ## Push the bundle image.
	$(info $(M) running $@)
	docker push $(BUNDLE_IMG)

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
catalog-build: $(OPM) ## Build a catalog image.
	$(info $(M) running $@)
	$(OPM) index add --container-tool docker --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)

# Push the catalog image.
.PHONY: catalog-push
catalog-push: ## Push a catalog image.
	$(info $(M) running $@)
	docker push $(CATALOG_IMG)

.PHONY: prep-release
prep-release: $(YQ)
	$(info $(M) running $@)
	$(YQ) -i '.version="v$(VERSION)"' deploy/helm/grafana-operator/Chart.yaml
	$(YQ) -i '.appVersion="v$(VERSION)"' deploy/helm/grafana-operator/Chart.yaml
	$(YQ) -i '.params.version="v$(VERSION)"' hugo/config.yaml
	sed -i 's/--version v5.*/--version v$(VERSION)/g' README.md
	sed -i 's/^VERSION ?= 5.*/VERSION ?= $(VERSION)/g' Makefile
	grep -q "$(GRAFANA_VERSION)" docs/docs/versioning.md || sed -Ei 's/\|-\|-\|/|-|-|\n| \`v$(VERSION)\` | \`$(GRAFANA_VERSION)\` |/' docs/docs/versioning.md
	$(YQ) -i '.images[0].newTag="v$(VERSION)"' deploy/kustomize/base/kustomization.yaml
	make helm-docs

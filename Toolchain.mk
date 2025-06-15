BIN = $(CURDIR)/bin
$(BIN):
	mkdir -p $(BIN)

PATH := $(BIN):$(PATH)

M = $(shell printf "\033[34;1m▶\033[0m")

CHAINSAW_VERSION = v0.2.12
CONTROLLER_GEN_VERSION = v0.17.3
CRDOC_VERSION = v0.6.4
ENVTEST_VERSION = v0.21.0
GOLANGCI_LINT_VERSION = v2.1.6
HELM_DOCS_VERSION = v1.14.2
HELM_VERSION = v3.17.3
HUGO_VERSION = 0.134.3
KIND_VERSION = v0.29.0
KO_VERSION = v0.18.0
KUSTOMIZE_VERSION = v5.6.0
OPERATOR_SDK_VERSION = v1.32.0
OPM_VERSION = v1.23.2
YQ_VERSION = v4.45.4

CHAINSAW := $(BIN)/chainsaw-$(CHAINSAW_VERSION)
$(CHAINSAW): | $(BIN)
	$(info $(M) installing chainsaw)
	@{ \
	set -e ;\
	OSTYPE=$(shell uname | awk '{print tolower($$0)}') && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(CHAINSAW).tar.gz https://github.com/kyverno/chainsaw/releases/download/$(CHAINSAW_VERSION)/chainsaw_$${OSTYPE}_$${ARCH}.tar.gz && \
	tar -zxvf $(CHAINSAW).tar.gz chainsaw && \
	chmod +x chainsaw && \
	mv chainsaw $(CHAINSAW) && \
	rm $(CHAINSAW).tar.gz ;\
	}

CONTROLLER_GEN := $(BIN)/controller-gen-$(CONTROLLER_GEN_VERSION)
$(CONTROLLER_GEN): | $(BIN)
	$(info $(M) installing controller-gen)
	@{ \
	set -e ;\
	OSTYPE=$(shell uname | awk '{print tolower($$0)}') && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(CONTROLLER_GEN) https://github.com/kubernetes-sigs/controller-tools/releases/download/$(CONTROLLER_GEN_VERSION)/controller-gen-$${OSTYPE}-$${ARCH} ;\
	chmod +x $(CONTROLLER_GEN) ;\
	}

CRDOC := $(BIN)/crdoc-$(CRDOC_VERSION)
$(CRDOC): | $(BIN)
	$(info $(M) installing crdoc)
	@{ \
	set -e ;\
	GOBIN=$(BIN) go install fybrik.io/crdoc@$(CRDOC_VERSION) ;\
	mv $(BIN)/crdoc $(CRDOC) ;\
	}

ENVTEST := $(BIN)/setup-envtest-$(ENVTEST_VERSION)
$(ENVTEST): | $(BIN)
	$(info $(M) installing setup-envtest)
	@{ \
	set -e ;\
	OSTYPE=$(shell uname | awk '{print tolower($$0)}') && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(ENVTEST) https://github.com/kubernetes-sigs/controller-runtime/releases/download/$(ENVTEST_VERSION)/setup-envtest-$${OSTYPE}-$${ARCH} ;\
	chmod +x $(ENVTEST) ;\
	}

GOLANGCI_LINT := $(BIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)
$(GOLANGCI_LINT): | $(BIN)
	$(info $(M) installing golangci-lint)
	@{ \
	set -e ;\
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s $(GOLANGCI_LINT_VERSION) ;\
	mv $(BIN)/golangci-lint $(GOLANGCI_LINT) ;\
	}

HELM := $(BIN)/helm-$(HELM_VERSION)
$(HELM): | $(BIN)
	$(info $(M) installing helm)
	@{ \
	set -e ;\
	curl -sSfL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | HELM_INSTALL_DIR=$(BIN) bash -s -- -v $(HELM_VERSION) --no-sudo ;\
	mv $(BIN)/helm $(HELM) ;\
	}

HELM_DOCS := $(BIN)/helm-docs-$(HELM_DOCS_VERSION)
$(HELM_DOCS): | $(BIN)
	$(info $(M) installing helm-docs)
	@{ \
	set -e ;\
	GOBIN=$(BIN) go install github.com/norwoodj/helm-docs/cmd/helm-docs@$(HELM_DOCS_VERSION) ;\
	mv $(BIN)/helm-docs $(HELM_DOCS) ;\
	}

HUGO := $(BIN)/hugo-$(HUGO_VERSION)
$(HUGO): | $(BIN)
	$(info $(M) installing hugo)
	@{ \
	set -e ;\
	OSTYPE=$(shell uname | awk '{print tolower($$0)}') && ARCH=$(shell go env GOARCH) && \
	if [ "`uname`" = "Darwin" ]; then ARCH="universal"; fi && \
	curl -sSLo $(HUGO).tar.gz https://github.com/gohugoio/hugo/releases/download/v$(HUGO_VERSION)/hugo_extended_$(HUGO_VERSION)_$${OSTYPE}-$${ARCH}.tar.gz && \
	tar -zxvf $(HUGO).tar.gz -C $(BIN) hugo && \
	mv $(BIN)/hugo $(HUGO) && \
	chmod +x $(HUGO) && \
	rm $(HUGO).tar.gz ;\
	}

KIND := $(BIN)/kind-$(KIND_VERSION)
$(KIND): | $(BIN)
	$(info $(M) installing kind)
	@{ \
	set -e ;\
	OSTYPE=$(shell uname | awk '{print tolower($$0)}') && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(KIND) https://github.com/kubernetes-sigs/kind/releases/download/$(KIND_VERSION)/kind-$${OSTYPE}-$${ARCH} ;\
	chmod +x $(KIND) ;\
	}

KO := $(BIN)/ko-$(KO_VERSION)
$(KO): | $(BIN)
	$(info $(M) installing ko)
	@{ \
	set -e ;\
	GOBIN=$(BIN) go install github.com/google/ko@$(KO_VERSION) ;\
	mv $(BIN)/ko $(KO) ;\
	}

KUSTOMIZE := $(BIN)/kustomize-$(KUSTOMIZE_VERSION)
$(KUSTOMIZE): | $(BIN)
	$(info $(M) installing kustomize)
	@{ \
	set -e ;\
	curl -sSfL "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(BIN) ;\
	mv $(BIN)/kustomize $(KUSTOMIZE) ;\
	}

OPERATOR_SDK := $(BIN)/operator-sdk-$(OPERATOR_SDK_VERSION)
$(OPERATOR_SDK): | $(BIN)
	$(info $(M) installing operator-sdk)
	@{ \
	set -e ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPERATOR_SDK) https://github.com/operator-framework/operator-sdk/releases/download/$(OPERATOR_SDK_VERSION)/operator-sdk_$${OS}_$${ARCH} ;\
	chmod +x $(OPERATOR_SDK);\
	}

OPM := $(BIN)/opm-$(OPM_VERSION)
$(OPM): | $(BIN)
	$(info $(M) installing opm)
	@{ \
	set -e ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/$(OPM_VERSION)/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}

YQ := $(BIN)/yq-$(YQ_VERSION)
$(YQ): | $(BIN)
	$(info $(M) installing yq)
	@{ \
	set -e ;\
	OSTYPE=$(shell uname | awk '{print tolower($$0)}') && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(YQ) https://github.com/mikefarah/yq/releases/download/$(YQ_VERSION)/yq_$${OSTYPE}_$${ARCH} ;\
	chmod +x $(YQ) ;\
	}

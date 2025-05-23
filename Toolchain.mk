BIN = $(CURDIR)/bin
$(BIN):
	mkdir -p $(BIN)

PATH := $(BIN):$(PATH)

CHAINSAW_VERSION = v0.2.10
CONTROLLER_GEN_VERSION = v0.16.3
CRDOC_VERSION = v0.6.4
ENVTEST_VERSION = 0.20
GOLANGCI_LINT_VERSION = v2.1.6
HELM_DOCS_VERSION = v1.11.0
HELM_VERSION = v3.16.2
KIND_VERSION = v0.27.0
KO_VERSION = v0.16.0
KUSTOMIZE_VERSION = v5.1.1
OPERATOR_SDK_VERSION = v1.32.0
OPM_VERSION = v1.23.2
YQ_VERSION = v4.35.2

CHAINSAW := $(BIN)/chainsaw-$(CHAINSAW_VERSION)
$(CHAINSAW): $(BIN)
ifeq (, $(shell which $(CHAINSAW)))
	@{ \
	set -e ;\
	GOBIN=$(BIN) go install github.com/kyverno/chainsaw@$(CHAINSAW_VERSION) ;\
	mv $(BIN)/chainsaw $(CHAINSAW) ;\
	}
endif

CONTROLLER_GEN := $(BIN)/controller-gen-$(CONTROLLER_GEN_VERSION)
$(CONTROLLER_GEN): $(BIN)
ifeq (, $(shell which $(CONTROLLER_GEN)))
	@{ \
	set -e ;\
	GOBIN=$(BIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION) ;\
	mv $(BIN)/controller-gen $(CONTROLLER_GEN) ;\
	}
endif

CRDOC := $(BIN)/crdoc-$(CRDOC_VERSION)
$(CRDOC): $(BIN)
ifeq (, $(shell which $(CRDOC)))
	@{ \
	set -e ;\
	GOBIN=$(BIN) go install fybrik.io/crdoc@$(CRDOC_VERSION) ;\
	mv $(BIN)/crdoc $(CRDOC) ;\
	}
endif

ENVTEST := $(BIN)/setup-envtest-$(ENVTEST_VERSION)
$(ENVTEST): $(BIN)
ifeq (, $(shell which $(ENVTEST)))
	@{ \
	set -e ;\
	GOBIN=$(BIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@release-$(ENVTEST_VERSION) ;\
	mv $(BIN)/setup-envtest $(ENVTEST) ;\
	}
endif

GOLANGCI_LINT := $(BIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)
$(GOLANGCI_LINT): $(BIN)
ifeq (, $(shell which $(GOLANGCI_LINT)))
	@{ \
	set -e ;\
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s $(GOLANGCI_LINT_VERSION) ;\
	mv $(BIN)/golangci-lint $(GOLANGCI_LINT) ;\
	}
endif

HELM := $(BIN)/helm-$(HELM_VERSION)
$(HELM): $(BIN)
ifeq (, $(shell which $(HELM)))
	@{ \
	set -e ;\
	curl -sSfL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | HELM_INSTALL_DIR=$(BIN) bash -s -- -v $(HELM_VERSION) --no-sudo ;\
	mv $(BIN)/helm $(HELM) ;\
	}
endif

HELM_DOCS := $(BIN)/helm-docs-$(HELM_DOCS_VERSION)
$(HELM_DOCS): $(BIN)
ifeq (, $(shell which $(HELM_DOCS)))
	@{ \
	set -e ;\
	GOBIN=$(BIN) go install github.com/norwoodj/helm-docs/cmd/helm-docs@$(HELM_DOCS_VERSION) ;\
	mv $(BIN)/helm-docs $(HELM_DOCS) ;\
	}
endif

KIND := $(BIN)/kind-$(KIND_VERSION)
$(KIND): $(BIN)
ifeq (, $(shell which $(KIND)))
	@{ \
	set -e ;\
	OSTYPE=$(shell uname | awk '{print tolower($$0)}') && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(KIND) https://github.com/kubernetes-sigs/kind/releases/download/$(KIND_VERSION)/kind-$${OSTYPE}-$${ARCH} ;\
	chmod +x $(KIND) ;\
	}
endif

KO := $(BIN)/ko-$(KO_VERSION)
$(KO): $(BIN)
ifeq (, $(shell which $(KO)))
	@{ \
	set -e ;\
	GOBIN=$(BIN) go install github.com/google/ko@$(KO_VERSION) ;\
	mv $(BIN)/ko $(KO) ;\
	}
endif

KUSTOMIZE := $(BIN)/kustomize-$(KUSTOMIZE_VERSION)
$(KUSTOMIZE): $(BIN)
ifeq (, $(shell which $(KUSTOMIZE)))
	@{ \
	set -e ;\
	curl -sSfL "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(BIN) ;\
	mv $(BIN)/kustomize $(KUSTOMIZE) ;\
	}
endif

OPERATOR_SDK := $(BIN)/operator-sdk-$(OPERATOR_SDK_VERSION)
$(OPERATOR_SDK): $(BIN)
ifeq (, $(shell which $(OPERATOR_SDK)))
	@{ \
	set -e ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPERATOR_SDK) https://github.com/operator-framework/operator-sdk/releases/download/$(OPERATOR_SDK_VERSION)/operator-sdk_$${OS}_$${ARCH} ;\
	chmod +x $(OPERATOR_SDK);\
	}
endif

OPM := $(BIN)/opm-$(OPM_VERSION)
$(OPM): $(BIN)
ifeq (, $(shell which $(OPM)))
	@{ \
	set -e ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/$(OPM_VERSION)/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}
endif

YQ := $(BIN)/yq-$(YQ_VERSION)
$(YQ): $(BIN)
ifeq (, $(shell which $(YQ)))
	@{ \
	set -e ;\
	OSTYPE=$(shell uname | awk '{print tolower($$0)}') && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(YQ) https://github.com/mikefarah/yq/releases/download/$(YQ_VERSION)/yq_$${OSTYPE}_$${ARCH} ;\
	chmod +x $(YQ) ;\
	}
endif

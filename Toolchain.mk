BIN = $(CURDIR)/bin
$(BIN):
	mkdir -p $(BIN)

PATH := $(BIN):$(PATH)

GOLANGCI_LINT_VERSION = v2.1.6
KUSTOMIZE_VERSION = v5.1.1
OPERATOR_SDK_VERSION = v1.32.0

GOLANGCI_LINT := $(BIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)
$(GOLANGCI_LINT): $(BIN)
ifeq (, $(shell which $(GOLANGCI_LINT)))
	@{ \
	set -e ;\
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s $(GOLANGCI_LINT_VERSION) ;\
	mv $(BIN)/golangci-lint $(GOLANGCI_LINT) ;\
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

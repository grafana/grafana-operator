BIN = $(CURDIR)/bin
$(BIN):
	mkdir -p $(BIN)

PATH := $(BIN):$(PATH)

GOLANGCI_LINT_VERSION = v2.1.6
KUSTOMIZE_VERSION = v5.1.1

GOLANGCI_LINT := $(BIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)
$(GOLANGCI_LINT):
ifeq (, $(shell which $(GOLANGCI_LINT)))
	@{ \
	set -e ;\
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s $(GOLANGCI_LINT_VERSION) ;\
	mv $(BIN)/golangci-lint $(GOLANGCI_LINT) ;\
	}
endif

KUSTOMIZE := $(BIN)/kustomize-$(KUSTOMIZE_VERSION)
$(KUSTOMIZE):
ifeq (, $(shell which $(KUSTOMIZE)))
	@{ \
	set -e ;\
	curl -sSfL "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(BIN) ;\
	mv $(BIN)/kustomize $(KUSTOMIZE) ;\
	}
endif

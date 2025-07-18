BIN = $(CURDIR)/bin
$(BIN):
	mkdir -p $(BIN)

M = $(shell printf "\033[34;1m▶\033[0m")

CHAINSAW_VERSION = v0.2.12
CONTROLLER_GEN_VERSION = v0.17.3
CRDOC_VERSION = v0.6.4
DART_SASS_VERSION = 1.86.0
ENVTEST_VERSION = v0.21.0
GOLANGCI_LINT_VERSION = v2.1.6
HELM_DOCS_VERSION = 1.14.2
HELM_VERSION = v3.17.3
HUGO_VERSION = 0.134.3
KIND_VERSION = v0.29.0
KO_VERSION = 0.18.0
KUSTOMIZE_VERSION = v5.6.0
MUFFET_VERSION = v2.10.9
OPERATOR_SDK_VERSION = v1.32.0
OPM_VERSION = v1.23.2
YQ_VERSION = v4.45.4

ifdef GITHUB_TOKEN
	CURL_GH_AUTH=-H 'Authorization: Bearer $(GITHUB_TOKEN)'
endif

CHAINSAW := $(BIN)/chainsaw-$(CHAINSAW_VERSION)
$(CHAINSAW): | $(BIN)
	$(info $(M) installing chainsaw)
	@{ \
	set -e ;\
	OSTYPE=$(shell uname | awk '{print tolower($$0)}') && ARCH=$(shell go env GOARCH) && \
	curl -sSLfo $(CHAINSAW).tar.gz $(CURL_GH_AUTH) https://github.com/kyverno/chainsaw/releases/download/$(CHAINSAW_VERSION)/chainsaw_$${OSTYPE}_$${ARCH}.tar.gz && \
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
	curl -sSLfo $(CONTROLLER_GEN) $(CURL_GH_AUTH) https://github.com/kubernetes-sigs/controller-tools/releases/download/$(CONTROLLER_GEN_VERSION)/controller-gen-$${OSTYPE}-$${ARCH} ;\
	chmod +x $(CONTROLLER_GEN) ;\
	}

CRDOC := $(BIN)/crdoc-$(CRDOC_VERSION)
$(CRDOC): | $(BIN)
	$(info $(M) installing crdoc)
	@{ \
	set -e ;\
	OSTYPE=$(shell uname | awk '{print tolower($$0)}') && ARCH=$(shell go env GOARCH) && \
	if [ "`go env GOARCH`" = "amd64" ]; then ARCH="x86_64"; fi && \
	curl -sSLfo $(CRDOC).tar.gz $(CURL_GH_AUTH) https://github.com/fybrik/crdoc/releases/download/$(CRDOC_VERSION)/crdoc_$${OSTYPE}_$${ARCH}.tar.gz && \
	tar -zxvf $(CRDOC).tar.gz -C $(BIN) crdoc && \
	mv $(BIN)/crdoc $(CRDOC) && \
	chmod +x $(CRDOC) && \
	rm $(CRDOC).tar.gz ;\
	}

DART_SASS := $(BIN)/dart-sass-$(DART_SASS_VERSION)
$(DART_SASS): | $(BIN)
	$(info $(M) installing dart-sass)
	@{ \
	set -e ;\
	OSTYPE=$(shell uname | awk '{print tolower($$0)}') && ARCH=$(shell go env GOARCH) && \
	if [ "`uname`" = "Darwin" ]; then OSTYPE="macos"; fi && \
	if [ "`go env GOARCH`" = "amd64" ]; then ARCH="x64"; fi && \
	curl -sSLfo $(DART_SASS).tar.gz $(CURL_GH_AUTH) https://github.com/sass/dart-sass/releases/download/$(DART_SASS_VERSION)/dart-sass-$(DART_SASS_VERSION)-$${OSTYPE}-$${ARCH}.tar.gz && \
	mkdir -p $(DART_SASS) && \
	tar -zxvf $(DART_SASS).tar.gz -C $(DART_SASS) --strip-components=1 && \
	rm $(DART_SASS).tar.gz ;\
	}

ENVTEST := $(BIN)/setup-envtest-$(ENVTEST_VERSION)
$(ENVTEST): | $(BIN)
	$(info $(M) installing setup-envtest)
	@{ \
	set -e ;\
	OSTYPE=$(shell uname | awk '{print tolower($$0)}') && ARCH=$(shell go env GOARCH) && \
	curl -sSLfo $(ENVTEST) $(CURL_GH_AUTH) https://github.com/kubernetes-sigs/controller-runtime/releases/download/$(ENVTEST_VERSION)/setup-envtest-$${OSTYPE}-$${ARCH} ;\
	chmod +x $(ENVTEST) ;\
	}

GOLANGCI_LINT := $(BIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)
$(GOLANGCI_LINT): | $(BIN)
	$(info $(M) installing golangci-lint)
	@{ \
	set -e ;\
	curl -sSfL $(CURL_GH_AUTH) https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s $(GOLANGCI_LINT_VERSION) ;\
	mv $(BIN)/golangci-lint $(GOLANGCI_LINT) ;\
	}

HELM := $(BIN)/helm-$(HELM_VERSION)
$(HELM): | $(BIN)
	$(info $(M) installing helm)
	@{ \
	set -e ;\
	curl -sSfL $(CURL_GH_AUTH) https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | HELM_INSTALL_DIR=$(BIN) bash -s -- -v $(HELM_VERSION) --no-sudo ;\
	mv $(BIN)/helm $(HELM) ;\
	}

HELM_DOCS := $(BIN)/helm-docs-v$(HELM_DOCS_VERSION)
$(HELM_DOCS): | $(BIN)
	$(info $(M) installing helm-docs)
	@{ \
	set -e ;\
	OSTYPE=$(shell uname | awk '{print tolower($$0)}') && ARCH=$(shell go env GOARCH) && \
	if [ "`go env GOARCH`" = "amd64" ]; then ARCH="x86_64"; fi && \
	curl -sSLfo $(HELM_DOCS).tar.gz $(CURL_GH_AUTH) https://github.com/norwoodj/helm-docs/releases/download/v$(HELM_DOCS_VERSION)/helm-docs_$(HELM_DOCS_VERSION)_$${OSTYPE}_$${ARCH}.tar.gz && \
	tar -zxvf $(HELM_DOCS).tar.gz -C $(BIN) helm-docs && \
	mv $(BIN)/helm-docs $(HELM_DOCS) && \
	chmod +x $(HELM_DOCS) && \
	rm $(HELM_DOCS).tar.gz ;\
	}

HUGO := $(BIN)/hugo-$(HUGO_VERSION)
$(HUGO): | $(BIN)
	$(info $(M) installing hugo)
	@{ \
	set -e ;\
	OSTYPE=$(shell uname | awk '{print tolower($$0)}') && ARCH=$(shell go env GOARCH) && \
	if [ "`uname`" = "Darwin" ]; then ARCH="universal"; fi && \
	curl -sSLfo $(HUGO).tar.gz $(CURL_GH_AUTH) https://github.com/gohugoio/hugo/releases/download/v$(HUGO_VERSION)/hugo_extended_$(HUGO_VERSION)_$${OSTYPE}-$${ARCH}.tar.gz && \
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
	curl -sSLfo $(KIND) $(CURL_GH_AUTH) https://github.com/kubernetes-sigs/kind/releases/download/$(KIND_VERSION)/kind-$${OSTYPE}-$${ARCH} ;\
	chmod +x $(KIND) ;\
	}

KO := $(BIN)/ko-v$(KO_VERSION)
$(KO): | $(BIN)
	$(info $(M) installing ko)
	@{ \
	set -e ;\
	OSTYPE=$(shell uname | awk '{print tolower($$0)}') && ARCH=$(shell go env GOARCH) && \
	if [ "`go env GOARCH`" = "amd64" ]; then ARCH="x86_64"; fi && \
	curl -sSLfo $(KO).tar.gz $(CURL_GH_AUTH) https://github.com/ko-build/ko/releases/download/v$(KO_VERSION)/ko_$(KO_VERSION)_$${OSTYPE}_$${ARCH}.tar.gz && \
	tar -zxvf $(KO).tar.gz -C $(BIN) ko && \
	mv $(BIN)/ko $(KO) && \
	chmod +x $(KO) && \
	rm $(KO).tar.gz ;\
	}

KUSTOMIZE := $(BIN)/kustomize-$(KUSTOMIZE_VERSION)
$(KUSTOMIZE): | $(BIN)
	$(info $(M) installing kustomize)
	@{ \
	set -e ;\
	curl -sSfL $(CURL_GH_AUTH) "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(BIN) ;\
	mv $(BIN)/kustomize $(KUSTOMIZE) ;\
	}

MUFFET := $(BIN)/muffet-$(MUFFET_VERSION)
$(MUFFET): | $(BIN)
	$(info $(M) installing muffet)
	@{ \
	set -e ;\
	OSTYPE=$(shell uname | awk '{print tolower($$0)}') && ARCH=$(shell go env GOARCH) && \
	curl -sSLfo $(MUFFET).tar.gz $(CURL_GH_AUTH) https://github.com/raviqqe/muffet/releases/download/$(MUFFET_VERSION)/muffet_$${OSTYPE}_$${ARCH}.tar.gz && \
	tar -zxvf $(MUFFET).tar.gz muffet && \
	chmod +x muffet && \
	mv muffet $(MUFFET) && \
	rm $(MUFFET).tar.gz ;\
	}

OPERATOR_SDK := $(BIN)/operator-sdk-$(OPERATOR_SDK_VERSION)
$(OPERATOR_SDK): | $(BIN)
	$(info $(M) installing operator-sdk)
	@{ \
	set -e ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLfo $(OPERATOR_SDK) $(CURL_GH_AUTH) https://github.com/operator-framework/operator-sdk/releases/download/$(OPERATOR_SDK_VERSION)/operator-sdk_$${OS}_$${ARCH} ;\
	chmod +x $(OPERATOR_SDK);\
	}

OPM := $(BIN)/opm-$(OPM_VERSION)
$(OPM): | $(BIN)
	$(info $(M) installing opm)
	@{ \
	set -e ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLfo $(OPM) $(CURL_GH_AUTH) https://github.com/operator-framework/operator-registry/releases/download/$(OPM_VERSION)/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}

YQ := $(BIN)/yq-$(YQ_VERSION)
$(YQ): | $(BIN)
	$(info $(M) installing yq)
	@{ \
	set -e ;\
	OSTYPE=$(shell uname | awk '{print tolower($$0)}') && ARCH=$(shell go env GOARCH) && \
	curl -sSLfo $(YQ) $(CURL_GH_AUTH) https://github.com/mikefarah/yq/releases/download/$(YQ_VERSION)/yq_$${OSTYPE}_$${ARCH} ;\
	chmod +x $(YQ) ;\
	}

PATH := $(DART_SASS):$(BIN):$(PATH)

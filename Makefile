ORG?=sapcc
NAMESPACE=grafana-operator
PROJECT=grafana-operator
REG?=quay.io
SHELL=/bin/bash
TAG?=latest
PKG=github.com/sapcc/grafana-operator
COMPILE_TARGET=./tmp/_output/bin/$(PROJECT)

GOOS     ?= $(shell go env GOOS)
ifeq ($(GOOS),darwin)
export CGO_ENABLED=0
endif
BINARIES := apiserver manager
IMAGE   ?= hub.global.cloud.sap/monsoon/grafana-operator
VERSION = $(shell git rev-parse --verify HEAD | head -c 8)

LDFLAGS := -X github.com/sapcc/kubernikus/pkg/version.GitCommit=$(VERSION)
GOFLAGS := -ldflags "$(LDFLAGS) -s -w"

SRCDIRS  := pkg cmd
PACKAGES := $(shell find $(SRCDIRS) -type d)
GOFILES  := $(addsuffix /*.go,$(PACKAGES))
GOFILES  := $(wildcard $(GOFILES))

BUILD_ARGS = --build-arg VERSION=$(VERSION)
BUILD_ARGS = --build-arg VERSION=$(VERSION)

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




all: $(BINARIES:%=bin/$(GOOS)/%)

bin/%: $(GOFILES) Makefile
	GOOS=$(*D) GOARCH=amd64 go build $(GOFLAGS) -v -i -o $(@D)/$(@F) ./cmd/$(basename $(@F))

.PHONY: build
build:
	$(info .....)
	docker build $(BUILD_ARGS) -t hub.global.cloud.sap/monsoon/grafana-operator-binaries:$(VERSION) -f Dockerfile.binaries .
	docker build $(BUILD_ARGS) -t hub.global.cloud.sap/monsoon/grafana-operator:$(VERSION) -f Dockerfile .
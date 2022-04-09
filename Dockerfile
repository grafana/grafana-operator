ARG BUILDER_IMAGE=golang:1.16
ARG UBI_MINIMAL_IMAGE=registry.access.redhat.com/ubi8/ubi-minimal:8.4
ARG UBI_MICRO_IMAGE=registry.access.redhat.com/ubi8/ubi-micro:8.4

# Build the manager binary
# hadolint ignore=DL3006
FROM --platform=${BUILDPLATFORM} ${BUILDER_IMAGE} as builder

ARG TARGETARCH
ARG TARGETOS
ARG GOPROXY
ARG GOPRIVATE

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod ./
COPY go.sum ./
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN GOPROXY=${GOPROXY} \
    GOPRIVATE=${GOPRIVATE} \
    go mod download

# Copy the go source
COPY main.go ./main.go
COPY api/ api/
COPY controllers/ controllers/
COPY internal/ internal/
COPY version/ version/

# Build
RUN GOPROXY=${GOPROXY} \
    GOPRIVATE=${GOPRIVATE} \
    CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build -a -o manager main.go

# hadolint ignore=DL3006
FROM --platform=${TARGETPLATFORM} ${UBI_MINIMAL_IMAGE} as ubi-minimal

# hadolint ignore=DL3006
FROM --platform=${TARGETPLATFORM} ${UBI_MICRO_IMAGE}

# copy Root CA bundle from ubi-minimal
COPY --from=ubi-minimal /etc/pki/tls/certs/ca-bundle.crt /etc/pki/tls/certs/ca-bundle.crt

WORKDIR /
COPY --from=builder /workspace/manager .

RUN mkdir -p /opt/jsonnet && chown nobody /opt/jsonnet

USER nobody

COPY grafonnet-lib/grafonnet/ /opt/jsonnet/grafonnet

ENTRYPOINT ["./manager"]

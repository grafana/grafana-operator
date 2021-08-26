# Build the manager binary
FROM registry.access.redhat.com/ubi8/go-toolset:1.15.14 as builder

USER root

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY version/ version/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o grafana-operator cmd/manager/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM registry.access.redhat.com/ubi8/ubi-minimal:8.4

WORKDIR /
COPY --from=builder /workspace/grafana-operator .

ADD grafonnet-lib/grafonnet/ /opt/jsonnet/grafonnet

USER 65532:65532

ENTRYPOINT ["/grafana-operator"]
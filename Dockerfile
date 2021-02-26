# Build the manager binary
FROM golang:1.15 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/

# Build
FROM registry.access.redhat.com/ubi8/ubi-minimal:8.3

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go
RUN mkdir -p /opt/jsonnet && chown nobody /opt/jsonnet

USER nobody

ADD grafonnet-lib/grafonnet/ /opt/jsonnet/grafonnet
ADD build/_output/bin/grafana-operator /usr/local/bin/grafana-operator

WORKDIR /
COPY --from=builder /workspace/manager .



ENTRYPOINT ["/manager"]

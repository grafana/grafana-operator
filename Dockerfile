# Build the manager binary
FROM --platform=${BUILDPLATFORM} golang:1.16 as builder

ARG TARGETARCH
ARG TARGETOS

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod ./
COPY go.sum ./
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go ./main.go
COPY api/ api/
COPY controllers/ controllers/
COPY internal/ internal/
COPY version/ version/

# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -o manager main.go

FROM --platform=${TARGETPLATFORM} registry.access.redhat.com/ubi8/ubi-micro:8.4
WORKDIR /
COPY --from=builder /workspace/manager .

RUN mkdir -p /opt/jsonnet && chown nobody /opt/jsonnet

USER 65532:65532

COPY grafonnet-lib/grafonnet/ /opt/jsonnet/grafonnet

ENTRYPOINT ["./manager"]

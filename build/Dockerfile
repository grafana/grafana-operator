FROM registry.access.redhat.com/ubi8/ubi-minimal:8.3

RUN mkdir -p /opt/jsonnet && chown nobody /opt/jsonnet

USER nobody

ADD grafonnet-lib/grafonnet/ /opt/jsonnet/grafonnet

ARG TARGETARCH
ADD build/_output/bin/grafana-operator-${TARGETARCH:-amd64} /usr/local/bin/grafana-operator

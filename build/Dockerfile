FROM registry.access.redhat.com/ubi8/ubi:8.0

RUN mkdir -p /opt/jsonnet && chown nobody /opt/jsonnet

USER nobody

ADD grafonnet-lib/grafonnet/ /opt/jsonnet/grafonnet

ADD build/_output/bin/grafana-operator /usr/local/bin/grafana-operator

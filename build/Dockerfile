FROM registry.access.redhat.com/ubi8/ubi-minimal:8.3 AS symlinker

ADD build/_output/bin/grafana-operator /usr/local/bin/grafana-operator
RUN cd /usr/local/bin/ && ln -sf grafana-operator stat

FROM gcr.io/distroless/static

USER nobody

ADD grafonnet-lib/grafonnet/ /opt/jsonnet/grafonnet

COPY --from=symlinker /usr/local/bin/ /usr/local/bin/

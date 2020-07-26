FROM alpine:3.8

RUN apk upgrade --update --no-cache

RUN mkdir -p /opt/jsonnet/grafonnet-lib

USER nobody

ADD grafonnet-lib/grafonnet /opt/jsonnet/grafonnet

ADD build/_output/bin/grafana-operator /usr/local/bin/grafana-operator

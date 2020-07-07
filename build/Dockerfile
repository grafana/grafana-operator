FROM alpine:3.8

RUN apk upgrade --update --no-cache

RUN mkdir -p /opt/grafonnet-lib

USER nobody

ADD grafonnet-lib /opt/grafonnet-lib

ADD build/_output/bin/grafana-operator /usr/local/bin/grafana-operator

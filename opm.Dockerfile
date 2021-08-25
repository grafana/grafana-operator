FROM quay.io/operator-framework/upstream-registry-builder AS builder

FROM scratch
LABEL operators.operatorframework.io.index.database.v1=./bundles.db
COPY bundles.db ./
COPY --from=builder /bin/opm /opm
COPY --from=builder /bin/grpc_health_probe /bin/grpc_health_probe
EXPOSE 50051
ENTRYPOINT ["/opm"]
CMD ["registry", "serve", "--database", "bundles.db"]
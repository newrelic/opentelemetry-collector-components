FROM golang:1.20 as builder

WORKDIR /build

COPY . ./

WORKDIR /build/examples/newrelicapm/builder

RUN CGO_ENABLED=0 GOOS=linux make

WORKDIR /dist
RUN cp /build/examples/newrelicapm/builder/distributions/nr-example-collector/_build/nr_example_collector ./nrotelcomponents
COPY examples/newrelicapm/otel-collector-config.yaml ./

FROM scratch

ARG USER_UID=10001
USER ${USER_UID}

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /dist /

ENTRYPOINT ["/nrotelcomponents"]
EXPOSE 4317

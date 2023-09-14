FROM golang:1.20 as builder

WORKDIR /build

COPY go.mod go.sum ./
COPY connector/apmconnector ./connector/apmconnector
COPY processor/apmprocessor ./processor/apmprocessor
COPY receiver/nopreceiver ./receiver/nopreceiver
RUN go mod download -x

COPY . ./

WORKDIR /build/cmd/nrotelcomponents
RUN CGO_ENABLED=0 GOOS=linux go build -v -a .

WORKDIR /dist
RUN cp /build/cmd/nrotelcomponents/nrotelcomponents ./nrotelcomponents
COPY examples/newrelicapm/otel-collector-config.yaml ./

FROM scratch

ARG USER_UID=10001
USER ${USER_UID}

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /dist /

ENTRYPOINT ["/nrotelcomponents"]
EXPOSE 4317

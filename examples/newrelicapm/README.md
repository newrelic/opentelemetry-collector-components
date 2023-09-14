# New Relic APM Components for the OpenTelemetry Collector

Contained in this repository are components for the OpenTelemetry Collector
that  generate metric data that drive New Relic’s APM experience. The generated
metric data is derived from spans received by the collector. Provided here are
simple steps to help you get started using these components.

## Quick start

Run the example from this directory as follows:

```shell
export NEW_RELIC_API_KEY=<your_api_key>
docker compose up --build
```

The example starts an instance of the OpenTelemetry Collector configured with
the New Relic APM [processor](../../processor/apmprocessor) and
[connector](../../connector/apmconnector) components. View the full
configuration of the collector [here](./otel-collector-config.yaml).

The example also starts a simple Java application instrumented with the
OpenTelemetry Java agent to enable you quickly get data reporting to New Relic.
Navigate to http://localhost:8080 to exercise the application.

In New Relic the service will be named `OpenTelemetry-NewRelic-APM-Demo`.

### Instrumenting your own applications

With the example running, if you have instrumented your own application with
OpenTelemetry, you can configure it with an OTLP exporter and export data to
http://localhost:4317.

**NOTE:** In order to get accurate metric data, it is important that you also
configure your application to sample and export all span data to the collector.

## Installing and deploying the collector in your environment

There are two ways you can install and deploy the collector configured with the
New Relic APM components:

1. Use the preview version of NRDOT that includes the components.
2. Build your own distribution of the collector.

### Use the preview version of NRDOT

**TODO:** Add details for how to acquire the preview version.

### Build your own distribution

The OpenTelemetry Collector community provides the [OpenTelemetry Collector
Builder](https://github.com/open-telemetry/opentelemetry-collector/tree/main/cmd/builder)
for easily building your own distribution of the collector.

If you're already familiar with building and running your own distribution of
the OpenTelemetry Collector, you can review and customize this
[manifest.yaml](builder/distributions/nr-example-collector/manifest.yaml) for
your needs.

If you have not used the OpenTelemetry Collector Builder before, it's easy to
get started. The manifest file is used to describe all the components for the
OpenTelemetry Collector that you want to include in your distribution. When the
builder is run it fetches all the components and compiles your distribution of
the collector. The collector is developed in Golang, so running the builder
requires that you have Golang tooling installed.

The [builder](./builder) directory contains some helpful scripts that make
downloading and running the builder easy. After modifying the manifest file
for your needs, run:

```shell
cd builder
make
```

The built collector binary can be found in the 
`builder/distributions/nr-example-collector/_build` directory. The binary is 
compiled for the architecture of the machine you run it on.

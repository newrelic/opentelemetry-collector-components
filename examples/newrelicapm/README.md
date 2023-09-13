# New Relic APM Components for the OpenTelemetry Collector

New Relic APM offers insights into the performance and health of your services.
Traditionally, the curated experiences New Relic APM provides are driven by
data sent using one of our language agents.

OpenTelemetry is an emerging standard that standardizes the format and
conventions of telemetry data emitted by services. New Relic’s goal is to power
the same experiences our customers have come to depend on whether running one
of our own language agents or an OpenTelemetry agent.

The OpenTelemetry Collector is a powerful tool that enables sophisticated
processing and transformation of telemetry data and is a key component for
achieving our goal. It is highly customizable and extensible.

Contained in this repo are components for the OpenTelemetry Collector that
generate metric data that drive New Relic’s APM experience. The generated
metric data is derived from spans received by the collector.

## Quick start

Provided is a simple example to help you get started using these components.
Run the example from this directory as follows:

```shell
export NEW_RELIC_API_KEY=<your_api_key>
docker compose up --build
```

The example includes a simple Java application instrumented with the
OpenTelemetry Java agent to enable you quickly get data reporting to New Relic.
Navigate to http://localhost:8080 to exercise the application.

In New Relic the service will be named `OpenTelemetry-NewRelic-APM-Demo`.

**NOTE:** The application currently shows up under "Services - OpenTelemetry".
When viewing the list of services, note the value of the `Provider` column. The
provider indicates which experience you'll see when viewing the service. For
services reporting data through a collector running New Relic's components the
provider value will be `newrelic-opentelemetry` and indicates the NewRelic APM
experience will be used when viewing the service.

## Instrumenting your own applications

**TODO: information here about instrumenting your own application. Including information
about configuring it to sample 100% spans and configuring it to send data over OTLP to
the collector.**

You can just run the collector and configure your own application to send data
over OTLP to http://localhost:4317:

```shell
export NEW_RELIC_API_KEY=<your_api_key>
docker compose up otel-collector
```

## Building your own collector with the New Relic components

**TODO:** add instructions for using the OpenTelemetry Collector Builder.

## Work in progress

The components for the collector that drive the New Relic APM experience are a
work-in-progress. 

**TODO: elaborate on what you can expect from the components today and expand on
the limitations below.**

There are some limitations and a number of APM features that
are not yet available:

* Transaction traces
* Slow SQL traces
* All spans for a transaction must arrive at the collector in a single batch

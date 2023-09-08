# New Relic APM

This example demonstrates New Relic components for the OpenTelemetry Collector
that drive New Relic's APM experience.

Run the collector from this directory as follows:

```shell
export NEW_RELIC_API_KEY=<your_api_key>
docker compose up --build
```

The collector is configured with an OTLP receiver. Once running, data can be
sent to it from a locally running application instrumented with an
OpenTelemetry agent or SDK.

The New Relic components generate metrics from trace data emitted by the
instrumented application. These generated metrics drive New Relic's APM
experience. As such, it is important that the application is configured to
sample all spans.

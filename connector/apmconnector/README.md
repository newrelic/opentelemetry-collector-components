# APM Connector

The APM Connector receives traces from the language agents and generates telemetry that will 
light up the New Relic APM UI.

## Configuration

This connector can act as a trace receiver and a metric and logs exporter. It is best used
together with the [apm processor](../../processor/apmprocessor).

Here is an example configuration:

```yaml
receivers:
  otlp:
    protocols:
      grpc:

processors:
  newrelicapm:
  batch:

exporters:
  otlp:
    endpoint: <endpoint>

connectors:
  newrelicapm:

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [newrelicapm, batch]
      exporters: [newrelicapm, otlp]
    metrics:
      receivers: [newrelicapm]
      processors: [batch]
      exporters: [otlp]
    logs:
      receivers: [newrelicapm]
      processors: [batch]
      exporters: [otlp]
```

## Data

### Metrics

Based on the traces, transaction duration, error count and apdex are generated.

### Logs

Server spans generate Logs which are converted to Transaction events in the New Relic backend.

### Spans

For database spans, the `db.sql.table` is added when it can be extracted from the `db.statement` attribute.
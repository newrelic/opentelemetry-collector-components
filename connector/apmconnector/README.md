# APM Connector

The APM Connector receives traces from the language agents and generates telemetry that will 
light up the New Relic APM UI.

## Configuration

This connector can act as a trace receiver and a metric, traces and logs exporter. 
Here is an example configuration:

```yaml
receivers:
  otlp:
    protocols:
      grpc:

processors:
  batch:

exporters:
  otlp:
    endpoint: <endpoint>

connectors:
  apm:

service:
  pipelines:
    traces/in:
      receivers: [otlp]
      processors: [batch]
      exporters: [apmconnector]
    metrics/out:
      receivers: [apmconnector]
      processors: [batch]
      exporters: [otlp]
    logs/out:
      receivers: [apmconnector]
      processors: [batch]
      exporters: [otlp]
    traces/out:
      receivers: [apmconnector]
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
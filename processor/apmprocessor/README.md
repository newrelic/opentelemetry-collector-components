# APM Processor

This processor inspects traces and enhances them with attributes that wil
light up the New Relic APM UI.

It is best used together with the [apm connector](../../connector/apmconnector).

It can be used on its own as such:

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

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [newrelicapm, batch]
      exporters: [otlp]
```
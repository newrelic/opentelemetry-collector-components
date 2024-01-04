# New Relic APM metrics

This document describes rules for deriving metrics which drive the
New Relic APM UI. The derived metrics are sourced from metrics defined by the
[OpenTelemetry semantic conventions](https://github.com/open-telemetry/semantic-conventions).

## Metrics derived from OpenTelemetry spans

The following metric is generated upon the first span in each scope.

* [`apm.service.instance.count`](#metric-apmserviceinstancecount)
  * This metric is only generated if the `host.name` attribute is present.
  * Attribute mapping:
    * `instanceName` -> `host.name`
    * `hostname.displayName` -> `host.name`

For each span, exactly one of the following metrics is generated depending on whether the span represents either a DB operation, external call, or some other duration.

* [`apm.service.datastore.operation.duration`](#metric-apmservicedatastoreoperationduration)
  * Generated if `db.system` and `db.operation` is present.
  * If `db.system` is `redis` and `db.operation` is not present then it is added with the value equal to the span name.
  * The following attributes, if present, are copied from the span:
    * `db.system`
    * `db.name`
    * `db.operation`
    * `db.sql.table` (if it does not exist, it will be parsed from `db.statement` if possible, otherwise it is set to `unknown`)
    * `server.address`
    * `server.port`
    * `net.peer.name`
  * Attribute mapping:
    * `transactionType` -> see [deriving transaction type from span data](#deriving-transaction-type-from-span-data)
    * `scope` -> see [deriving transaction name from span data](#deriving-transaction-name-from-span-data)
    * `metricTimesliceName` -> Datastore/statement/{db.system}/{db.sql.table | 'unknown'}/{db.operation}

* [`apm.service.transaction.external.host.duration`](#metric-apmservicetransactionexternalhostduration)
  * Generated if it does not represent a DB call and `server.address` or `net.peer.name` is present.
  * Attribute mapping:
    * `transactionType` -> see [deriving transaction type from span data](#deriving-transaction-type-from-span-data)
    * `scope` -> see [deriving transaction name from span data](#deriving-transaction-name-from-span-data)
    * `server.address` -> {`server.address` | `net.peer.name`}
    * `external.host` -> {`server.address` | `net.peer.name`}
    * `metricTimesliceName` -> `External/{`server.address` | `net.peer.name`}/all`

* [`newrelic.timeslice.value`](#metric-newrelictimeslicevalue)
  * Generated if the span does not represent a DB or external call.
  * Attribute mapping:
    * `transactionType` -> see [deriving transaction type from span data](#deriving-transaction-type-from-span-data)
    * `scope` -> see [deriving transaction name from span data](#deriving-transaction-name-from-span-data)
    * `metricTimesliceName` -> `Custom/{span name}`

For each span, a number of overview metrics are generated.

* [`apm.service.transaction.overview`](#metric-apmservicetransactionoverview)
* [`apm.service.overview.web`](#metric-apmserviceoverviewweb)
* [`apm.service.overview.other`](#metric-apmserviceoverviewother)

For a group of spans representing a single transaction, a sampled duration metric is generated.

* [`apm.service.transaction.sampled_duration`](#metric-apmservicetransactionsampled_duration)

## Metrics derived from OpenTelemetry metrics

This section describes the specific metrics from OpenTelemetry which are used
to derive New Relic APM metrics and the rules for deriving them. For all
derived metrics, certain [resource attributes](#resource-attributes) are also
applied when present.

### HTTP Server

The following metrics are derived from the [`http.server.request.duration`](https://github.com/open-telemetry/semantic-conventions) metric (formerly `http.server.duration`).

| Name | Condition |
| ---- | --------- |
| [`apm.service.transaction.duration`](#metric-apmservicetransactionduration) | Always |
| [`apm.service.error.count`](#metric-apmserviceerrorcount) | `http.response.status_code` >= 500 |
| [`apm.service.transaction.error.count`](#metric-apmservicetransactionerrorcount) | `http.response.status_code` >= 500 |
| [`apm.service.apdex`](#metric-apmserviceapdex) | Always |
| [`apm.service.transaction.apdex`](#metric-apmservicetransactionapdex) | Always |

The attributes of the derived metrics are mapped as follows:

* `transactionType` is always `WebTransaction`
* `transactionName` is derived in order of preference:
  * `WebTransaction/http.route{http.route} ({http.request.method})`
  * `WebTransaction/http.route{http.route}`
  * `WebTransaction/Uri{uri.path | http.target} ({http.request.method})`
  * `WebTransaction/Uri{uri.path | http.target}`
  * `WebTransaction/http.method/{http.request.method}`
  * `WebTransaction/Other/Unknown`
* For `apdex.value` and `apdex.zone` see [computing Apdex](#computing-apdex).

### RPC Server

The following metrics are derived from the [`rpc.server.duration`](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/rpc/rpc-metrics.md#metric-rpcserverduration)
metric.

| Name | Condition |
| ---- | --------- |
| [`apm.service.transaction.duration`](#metric-apmservicetransactionduration) | Always |
| [`apm.service.apdex`](#metric-apmserviceapdex) | Always |
| [`apm.service.transaction.apdex`](#metric-apmservicetransactionapdex) | Always |

The attributes of the derived metrics are mapped as follows:

* `transactionType` is always `WebTransaction`
* `transactionName` is derived in order or preference:
  * `WebTransaction/rpc/{rpc.service}/{rpc.method}`
  * `WebTransaction/rpc/{rpc.service}`
  * `WebTransaction/Other/Unknown`
* For `apdex.value` and `apdex.zone` see [computing Apdex](#computing-apdex).

### HTTP Client

The following metrics are derived from the [`http.client.request.duration`](https://github.com/open-telemetry/semantic-conventions) metric (formerly `http.client.duration`).

| Name | Condition |
| ---- | --------- |
| [`apm.service.external.host.duration`](#metric-apmserviceexternalhostduration) | Always |

The attributes of the derived metrics are mapped as follows:

* `server.address` mapped from `server.address` or `net.peer.name`.
* `external.host` mapped from `server.address` or `net.peer.name`.
* `metricTimesliceName` is `External/{server.address}/all`

## Semantic Conventions for APM Metrics

### Metric: `apm.service.instance.count`

| Name     | Instrument Type | Unit (UCUM) | Description    |
| -------- | --------------- | ----------- | -------------- |
| `apm.service.instance.count` | Counter | `{instance}` | Instance of an APM service. |

| Attribute  | Type | Description  | Examples  | Requirement Level |
|---|---|---|---|---|
| `instanceName` | string | Enables the host dropdown in APM UI | `SomeHost` | Required |
| `hostname.displayName` | string | Enables the host dropdown in APM UI | `SomeHost` | Required |

### Metric: `apm.service.datastore.operation.duration`

| Name     | Instrument Type | Unit (UCUM) | Description    |
| -------- | --------------- | ----------- | -------------- |
| `apm.service.datastore.operation.duration` | Histogram | `s` | The duration of a datastore operation. |

| Attribute  | Type | Description  | Examples  | Requirement Level |
|---|---|---|---|---|
| `transactionType` | string | TODO | TODO | Required |
| `scope` | string | TODO | TODO | Required |
| `db.system` | string | TODO | TODO | Required |
| `db.name` | string | TODO | TODO | Required |
| `db.operation` | string | TODO | TODO | Required |
| `db.sql.table` | string | TODO | TODO | Required |
| `server.address` | string | TODO | TODO | Required |
| `server.port` | string | TODO | TODO | Required |
| `net.peer.name` | string | TODO | TODO | Required |
| `metricTimesliceName` | string | TODO | TODO | Required |

### Metric: `apm.service.transaction.external.host.duration`

| Name     | Instrument Type | Unit (UCUM) | Description    |
| -------- | --------------- | ----------- | -------------- |
| `apm.service.transaction.external.host.duration` | Histogram | `s` | The duration of an external call. |

| Attribute  | Type | Description  | Examples  | Requirement Level |
|---|---|---|---|---|
| `transactionType` | string | TODO | TODO | Required |
| `scope` | string | TODO | TODO | Required |
| `server.address` | string | TODO | TODO | Required |
| `external.host` | string | TODO | TODO | Required |
| `metricTimesliceName` | string | TODO | TODO | Required |

### Metric: `newrelic.timeslice.value`

| Name     | Instrument Type | Unit (UCUM) | Description    |
| -------- | --------------- | ----------- | -------------- |
| `newrelic.timeslice.value` | Histogram | `s` | The duration of a generic segment. |

| Attribute  | Type | Description  | Examples  | Requirement Level |
|---|---|---|---|---|
| `transactionType` | string | TODO | TODO | Required |
| `scope` | string | TODO | TODO | Required |
| `metricTimesliceName` | string | TODO | TODO | Required |

### Metric: `apm.service.transaction.overview`

| Name     | Instrument Type | Unit (UCUM) | Description    |
| -------- | --------------- | ----------- | -------------- |
| `apm.service.transaction.overview` | Histogram | `s` | The duration for a portion of a transaction. |

Attributes for a portion of the total duration which is associated with a particular span:

| Attribute  | Type | Description  | Examples  | Requirement Level |
|---|---|---|---|---|
| `transactionName` | string | TODO | TODO | Required |

Attributes for any remaining portion of the duration not associated with a particular span:

| Attribute  | Type | Description  | Examples  | Requirement Level |
|---|---|---|---|---|
| `transactionType` | string | TODO | TODO | Required |
| `transactionName` | string | TODO | TODO | Required |
| `metricTimesliceName` | string | TODO | TODO | Required |

### Metric: `apm.service.overview.web`

| Name     | Instrument Type | Unit (UCUM) | Description    |
| -------- | --------------- | ----------- | -------------- |
| `apm.service.overview.web` | Histogram | `s` | The duration of web transactions. |

| Attribute  | Type | Description  | Examples  | Requirement Level |
|---|---|---|---|---|
| `segmentName` | string | TODO | TODO | Required |

### Metric: `apm.service.overview.other`

| Name     | Instrument Type | Unit (UCUM) | Description    |
| -------- | --------------- | ----------- | -------------- |
| `apm.service.overview.other` | Histogram | `s` | The duration of other transactions. |

| Attribute  | Type | Description  | Examples  | Requirement Level |
|---|---|---|---|---|
| `segmentName` | string | TODO | TODO | Required |

### Metric: `apm.service.transaction.sampled_duration`

| Name     | Instrument Type | Unit (UCUM) | Description    |
| -------- | --------------- | ----------- | -------------- |
| `apm.service.transaction.sampled_duration` | Histogram | `s` | Duration of a transaction sourced from sampled span data. |

| Attribute  | Type | Description  | Examples  | Requirement Level |
|---|---|---|---|---|
| `transactionType` | string | The type of the transaction. | `WebTransaction`; `OtherTransaction` | Required |
| `transactionName` | string | The name of the transaction. | `WebTransaction/http.route/vets`; `WebTransaction/http.request.method/POST` | Required |
| `metricTimesliceName` | string | Same as `transactionName`. | Same as `transactionName`. | Required |

### Metric: `apm.service.transaction.duration`

| Name     | Instrument Type | Unit (UCUM) | Description    |
| -------- | --------------- | ----------- | -------------- |
| `apm.service.transaction.duration` | Histogram | `s` | Duration of a transaction. |

| Attribute  | Type | Description  | Examples  | Requirement Level |
|---|---|---|---|---|
| `transactionType` | string | The type of the transaction. | `WebTransaction`; `OtherTransaction` | Required |
| `transactionName` | string | The name of the transaction. | `WebTransaction/http.route/vets`; `WebTransaction/http.request.method/POST` | Required |
| `metricTimesliceName` | string | Same as `transactionName`. | Same as `transactionName`. | Required |

### Metric: `apm.service.error.count`

| Name     | Instrument Type | Unit (UCUM) | Description    |
| -------- | --------------- | ----------- | -------------- |
| `apm.service.error.count` | Counter | `{error}` | Number of errors. |

| Attribute  | Type | Description  | Examples  | Requirement Level |
|---|---|---|---|---|
| `transactionType` | string | The type of the transaction. | `WebTransaction`; `OtherTransaction` | Required |

### Metric: `apm.service.transaction.error.count`

| Name     | Instrument Type | Unit (UCUM) | Description    |
| -------- | --------------- | ----------- | -------------- |
| `apm.service.transaction.error.count` | Counter | `{error}` | Number of errors for a transaction. |

| Attribute  | Type | Description  | Examples  | Requirement Level |
|---|---|---|---|---|
| `transactionType` | string | The type of the transaction. | `WebTransaction`; `OtherTransaction` | Required |
| `transactionName` | string | The name of the transaction. | `WebTransaction/http.route/vets`; `WebTransaction/http.request.method/POST` | Required |

### Metric: `apm.service.apdex`

| Name     | Instrument Type | Unit (UCUM) | Description    |
| -------- | --------------- | ----------- | -------------- |
| `apm.service.apdex` | Counter | `{apdex}` | Apdex. |

| Attribute  | Type | Description  | Examples  | Requirement Level |
|---|---|---|---|---|
| `transactionType` | string | The type of the transaction. | `WebTransaction`; `OtherTransaction` | Required |
| `apdex.value` | double | The satisfying Apdex value. | `0.5` | Required |
| `apdex.zone` | double | The Apdex zone. | `S`; `T`; `F` | Required |

### Metric: `apm.service.transaction.apdex`

| Name     | Instrument Type | Unit (UCUM) | Description    |
| -------- | --------------- | ----------- | -------------- |
| `apm.service.transaction.apdex` | Counter | `{apdex}` | Apdex. |

| Attribute  | Type | Description  | Examples  | Requirement Level |
|---|---|---|---|---|
| `transactionType` | string | The type of the transaction. | `WebTransaction`; `OtherTransaction` | Required |
| `transactionName` | string | The name of the transaction. | `WebTransaction/http.route/vets`; `WebTransaction/http.request.method/POST` | Required |
| `apdex.value` | double | The satisfying Apdex value. | `0.5` | Required |
| `apdex.zone` | double | The Apdex zone. | `S`; `T`; `F` | Required |

### Metric: `apm.service.external.host.duration`

| Name     | Instrument Type | Unit (UCUM) | Description    |
| -------- | --------------- | ----------- | -------------- |
| `apm.service.external.host.duration` | Histogram | `s` | Duration of an external call. |

| Attribute  | Type | Description  | Examples  | Requirement Level |
|---|---|---|---|---|
| `server.address` | string | Host identifier of the ["URI origin"](https://www.rfc-editor.org/rfc/rfc9110.html#name-uri-origin) HTTP request is sent to. | `WebTransaction`; `OtherTransaction` | Required |
| `external.host` | string | Same as `server.address` | `example.com`; `10.1.2.80`; `/tmp/my.sock` | Required |
| `metricTimesliceName` | string | Equivalent timeslice metric name. | `External/{server.address}/all` | Required |

## Appendix

### Computing Apdex

TODO: describe computing Apdex from a histogram.

### Resource attributes

Metrics generated for driving the New Relic APM UI inherit the following
resource attributes defined by the OpenTelemetry semantic conventions
if present on the source metric.

* `host.arch`
* `host.name`
* `os.description`
* `os.type`
* `process.runtime.description`
* `process.runtime.version`
* `telemetry.auto.version`
* `telemetry.sdk.language`
* `telemetry.sdk.name`
* `telemetry.sdk.version`
* `service.name`
* `service.instance.id` (If not present, the APM Processor sets this from `host.name` if present)

Additionally, the following New Relic specific resource attributes are applied.

* `host` (The APM Processor duplicates this from `host.name` if present)
* `instrumentation.provider` (This is set to the value `newrelic-opentelemetry`)

### Deriving transaction type from span data

TODO

### Deriving transaction name from span data

TODO

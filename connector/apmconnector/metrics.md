# New Relic synthesized metrics

## Resource attributes

The following attributes are applied as resource attributes.

* `instrumentation.provider`
* `os.description`
* `telemetry.auto.version`
* `telemetry.sdk.language`
* `host.name`
* `os.type`
* `telemetry.sdk.name`
* `process.runtime.description`
* `process.runtime.version`
* `telemetry.sdk.version`
* `host.arch`
* `service.name`
* `service.instance.id` (If not present, the APM Processor sets this from `host.name` if present)
* `host` (The APM Processor duplicates this from `host.name` if present)

## OpenTelemetry metrics

### Metric: `http.server.request.duration`

Formerly `http.server.duration`.

Derived metrics:

| Name | Condition |
| ---- | --------- |
| [`apm.service.transaction.duration`](#metric-apmservicetransactionduration) | Always |
| [`apm.service.error.count`](#metric-apmserviceerrorcount) | `http.response.status_code` >= 500 |
| [`apm.service.transaction.error.count`](#metric-apmservicetransaactionerrorcount) | `http.response.status_code` >= 500 |
| [`apm.service.apdex`](#metric-apmserviceapdex) | Always |
| [`apm.service.transaction.apdex`](#metric-apmservicetransactionapdex) | Always |

Attribute mapping:

* `transactionType` is always `WebTransaction`
* `transactionName` is derived in order or preference:
  * `WebTransaction/http.route{http.route} ({http.request.method})`
  * `WebTransaction/http.route{http.route}`
  * `WebTransaction/Uri{uri.path | http.target} ({http.request.method})`
  * `WebTransaction/Uri{uri.path | http.target}`
  * `WebTransaction/http.method/{http.request.method}`
  * `WebTransaction/Other/Unknown`
* `apdex.value`
* `apdex.zone`

### Metric: `rpc.server.duration`

Derived metrics:

| Name | Condition |
| ---- | --------- |
| [`apm.service.transaction.duration`](#metric-apmservicetransactionduration) | Always |
| [`apm.service.apdex`](#metric-apmserviceapdex) | Always |
| [`apm.service.transaction.apdex`](#metric-apmservicetransactionapdex) | Always |

Attribute mapping:

* `transactionType` is always `WebTransaction`
* `transactionName` is derived in order or preference:
  * `WebTransaction/rpc/{rpc.service}/{rpc.method}`
  * `WebTransaction/rpc/{rpc.service}`
  * `WebTransaction/Other/Unknown`
* `apdex.value`
* `apdex.zone`

### Metric: `http.client.request.duration`

Formerly `http.client.duration`.

Derived metrics:

| Name | Condition |
| ---- | --------- |
| [`apm.service.external.host.duration`](#metric-apmserviceexternalhostduration) | Always |

Attribute mapping:

* `server.address` mapped from `server.address` or `net.peer.name`.
* `external.host` mapped from `server.address` or `net.peer.name`.
* `metricTimesliceName` is `External/{server.address}/all`

## APM

### Metric: `apm.service.transaction.duration`

| Name     | Instrument Type | Unit (UCUM) | Description    |
| -------- | --------------- | ----------- | -------------- |
| `apm.service.transaction.duration` | Histogram | `s` | Duration of a transaction. |

| Attribute  | Type | Description  | Examples  | Requirement Level |
|---|---|---|---|---|
| [`transactionType`](../attributes-registry/error.md) | string | The type of the transaction. [1] | `WebTransaction`; `OtherTransaction` | Required |
| [`transactionName`](../attributes-registry/http.md) | string | The name of the transaction. [2] | `WebTransaction/http.route/vets`; `WebTransaction/http.request.method/POST` | Required |
| [`metricTimesliceName`](../attributes-registry/http.md) | string | Same as `transactionName`. | Same as `transactionName` | Required |

### Metric: `apm.service.error.count`

| Name     | Instrument Type | Unit (UCUM) | Description    |
| -------- | --------------- | ----------- | -------------- |
| `apm.service.error.count` | Counter | `{error}` | Number of errors. |

| Attribute  | Type | Description  | Examples  | Requirement Level |
|---|---|---|---|---|
| [`transactionType`](../attributes-registry/error.md) | string | The type of the transaction. [1] | `WebTransaction`; `OtherTransaction` | Required |

### Metric: `apm.service.transaction.error.count`

| Name     | Instrument Type | Unit (UCUM) | Description    |
| -------- | --------------- | ----------- | -------------- |
| `apm.service.transaction.error.count` | Counter | `{error}` | Number of errors for a transaction. |

| Attribute  | Type | Description  | Examples  | Requirement Level |
|---|---|---|---|---|
| [`transactionType`](../attributes-registry/error.md) | string | The type of the transaction. [1] | `WebTransaction`; `OtherTransaction` | Required |
| [`transactionName`](../attributes-registry/http.md) | string | The name of the transaction. [2] | `WebTransaction/http.route/vets`; `WebTransaction/http.request.method/POST` | Required |

### Metric: `apm.service.apdex`

| Name     | Instrument Type | Unit (UCUM) | Description    |
| -------- | --------------- | ----------- | -------------- |
| `apm.service.apdex` | Counter | `{apdex}` | Apdex. |

| Attribute  | Type | Description  | Examples  | Requirement Level |
|---|---|---|---|---|
| [`transactionType`](../attributes-registry/error.md) | string | The type of the transaction. [1] | `WebTransaction`; `OtherTransaction` | Required |
| [`apdex.value`](../attributes-registry/http.md) | double | The satisfying Apdex value. [2] | `0.5` | Required |
| [`apdex.zone`](../attributes-registry/http.md) | double | The Apdex zone. [2] | `S`; `T`; `F` | Required |

### Metric: `apm.service.transaction.apdex`

| Name     | Instrument Type | Unit (UCUM) | Description    |
| -------- | --------------- | ----------- | -------------- |
| `apm.service.transaction.apdex` | Counter | `{apdex}` | Apdex. |

| Attribute  | Type | Description  | Examples  | Requirement Level |
|---|---|---|---|---|
| [`transactionType`](../attributes-registry/error.md) | string | The type of the transaction. [1] | `WebTransaction`; `OtherTransaction` | Required |
| [`transactionName`](../attributes-registry/http.md) | string | The name of the transaction. [2] | `WebTransaction/http.route/vets`; `WebTransaction/http.request.method/POST` | Required |
| [`apdex.value`](../attributes-registry/http.md) | double | The satisfying Apdex value. [2] | `0.5` | Required |
| [`apdex.zone`](../attributes-registry/http.md) | double | The Apdex zone. [2] | `S`; `T`; `F` | Required |

### Metric: `apm.service.external.host.duration`

| Name     | Instrument Type | Unit (UCUM) | Description    |
| -------- | --------------- | ----------- | -------------- |
| `apm.service.external.host.duration` | Histogram | `s` | Duration of an external call. |

| Attribute  | Type | Description  | Examples  | Requirement Level |
|---|---|---|---|---|
| [`server.address`](../attributes-registry/error.md) | string | Host identifier of the ["URI origin"](https://www.rfc-editor.org/rfc/rfc9110.html#name-uri-origin) HTTP request is sent to. [1] | `WebTransaction`; `OtherTransaction` | Required |
| [`external.host`](../attributes-registry/error.md) | string | Same as `server.address` | `example.com`; `10.1.2.80`; `/tmp/my.sock` | Required |
| [`metricTimesliceName`](../attributes-registry/http.md) | string | Equivalent timeslice metric name. | `External/{server.address}/all` | Required |

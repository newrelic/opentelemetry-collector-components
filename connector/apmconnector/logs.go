// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmconnector // import "apmconnector"

import (
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func BuildTransactions(td ptrace.Traces) plog.Logs {
	logs := plog.NewLogs()
	for i := 0; i < td.ResourceSpans().Len(); i++ {
		resourceLogs := logs.ResourceLogs().AppendEmpty()
		rs := td.ResourceSpans().At(i)
		rs.Resource().CopyTo(resourceLogs.Resource())
		for j := 0; j < rs.ScopeSpans().Len(); j++ {
			scopeSpan := rs.ScopeSpans().At(j)
			scopeLog := resourceLogs.ScopeLogs().AppendEmpty()
			for k := 0; k < scopeSpan.Spans().Len(); k++ {
				span := scopeSpan.Spans().At(k)
				if span.Kind() != ptrace.SpanKindServer {
					continue
				}
				log := scopeLog.LogRecords().AppendEmpty()
				buildTransaction(log, span)
			}
		}
	}
	return logs
}

func buildTransaction(lr plog.LogRecord, span ptrace.Span) {
	lr.Attributes().PutStr("event.domain", "newrelic.otel_collector")
	lr.Attributes().PutStr("event.name", "Transaction")

	transactionName, transactionType := GetTransactionMetricName(span)
	lr.Attributes().PutStr("transactionType", transactionType.AsString())
	lr.Attributes().PutStr("name", transactionName)

	lr.Attributes().PutStr("trace.id", span.TraceID().String())
	duration := float64((span.EndTimestamp() - span.StartTimestamp()).AsTime().UnixNano()) / 1e9
	lr.Attributes().PutDouble("duration", duration)
	err := span.Status().Code() == ptrace.StatusCodeError
	lr.Attributes().PutBool("error", err)
}

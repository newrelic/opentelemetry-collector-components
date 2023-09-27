// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmconnector

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func TestApdex(t *testing.T) {
	apdex := NewApdex(0.5)
	assert.Equal(t, "S", apdex.GetApdexZone(0.1))
	assert.Equal(t, "S", apdex.GetApdexZone(0.5))
	assert.Equal(t, "T", apdex.GetApdexZone(0.51))
	assert.Equal(t, "T", apdex.GetApdexZone(1.1))
	assert.Equal(t, "T", apdex.GetApdexZone(2.0))
	assert.Equal(t, "F", apdex.GetApdexZone(2.1))
}

func TestGetTransactionMetricNameUnknown(t *testing.T) {
	span := ptrace.NewSpan()
	span.SetKind(ptrace.SpanKindServer)
	span.SetName("Test")

	name, txType := GetTransactionMetricName(span)
	assert.Equal(t, "WebTransaction/Other/Test", name)
	assert.Equal(t, WebTransactionType, txType)
	assert.Equal(t, "Web", txType.AsString())
}

func TestGetTransactionMetricNameRoute(t *testing.T) {
	span := ptrace.NewSpan()
	span.SetKind(ptrace.SpanKindServer)
	span.Attributes().PutStr("http.route", "/users")
	span.Attributes().PutStr("http.method", "GET")

	name, txType := GetTransactionMetricName(span)
	assert.Equal(t, "WebTransaction/http.route/users (GET)", name)
	assert.Equal(t, WebTransactionType, txType)
}

func TestGetTransactionMetricNameUrlPath(t *testing.T) {
	span := ptrace.NewSpan()
	span.SetKind(ptrace.SpanKindServer)
	span.Attributes().PutStr("url.path", "/owners/5")

	name, txType := GetTransactionMetricName(span)
	assert.Equal(t, "WebTransaction/Uri/owners/5", name)
	assert.Equal(t, WebTransactionType, txType)
}

func TestGetTransactionMetricNameHttpTarget(t *testing.T) {
	span := ptrace.NewSpan()
	span.SetKind(ptrace.SpanKindServer)
	span.Attributes().PutStr("http.target", "/owners/5")
	span.Attributes().PutStr("http.request.method", "GET")

	name, txType := GetTransactionMetricName(span)
	assert.Equal(t, "WebTransaction/Uri/owners/5 (GET)", name)
	assert.Equal(t, WebTransactionType, txType)
}

func TestGetOrCreateTransaction(t *testing.T) {
	transactions := NewTransactionsMap(0.5)
	span := ptrace.NewSpan()
	metricMap := NewMetrics()
	resources := pcommon.NewMap()
	metrics := metricMap.GetOrCreateResource(resources)
	transaction, _ := transactions.GetOrCreateTransaction("java", span, metrics, resources)

	transaction.SetRootSpan(span)
	assert.Equal(t, true, transaction.IsRootSet())
	transactions.ProcessTransactions()

	existingTransaction, _ := transactions.GetOrCreateTransaction("java", span, metrics, resources)
	assert.Same(t, transaction, existingTransaction)
	assert.Equal(t, true, existingTransaction.IsRootSet())
}

func TestGetOrCreateTransactionMultipleSpans(t *testing.T) {
	transactions := NewTransactionsMap(0.5)
	span := ptrace.NewSpan()
	span.SetTraceID(pcommon.TraceID{0x01})
	span.SetSpanID(pcommon.SpanID{0x01})
	metricMap := NewMetrics()
	resources := pcommon.NewMap()
	resources.PutStr("service.name", "authentication")
	metrics := metricMap.GetOrCreateResource(resources)
	transaction, _ := transactions.GetOrCreateTransaction("java", span, metrics, resources)

	span = ptrace.NewSpan()
	span.SetTraceID(pcommon.TraceID{0x01})
	span.SetSpanID(pcommon.SpanID{0x02})

	existingTransaction, _ := transactions.GetOrCreateTransaction("java", span, metrics, resources)
	assert.Same(t, transaction, existingTransaction)
}

func TestGetOrCreateTransactionMultipleServices(t *testing.T) {
	transactions := NewTransactionsMap(0.5)
	span := ptrace.NewSpan()
	span.SetTraceID(pcommon.TraceID{0x01})
	span.SetSpanID(pcommon.SpanID{0x01})
	metricMap := NewMetrics()
	resources := pcommon.NewMap()
	resources.PutStr("service.name", "authentication")
	metrics := metricMap.GetOrCreateResource(resources)
	transaction, _ := transactions.GetOrCreateTransaction("java", span, metrics, resources)

	span = ptrace.NewSpan()
	span.SetTraceID(pcommon.TraceID{0x01})
	span.SetSpanID(pcommon.SpanID{0x02})

	resources.PutStr("service.name", "cart")

	existingTransaction, _ := transactions.GetOrCreateTransaction("java", span, metrics, resources)
	assert.NotSame(t, transaction, existingTransaction)
}

func TestGetTransactionMetricNameRpcService(t *testing.T) {
	span := ptrace.NewSpan()
	span.SetKind(ptrace.SpanKindServer)
	span.Attributes().PutStr("rpc.service", "oteldemo.CheckoutService")

	name, txType := GetTransactionMetricName(span)
	assert.Equal(t, "WebTransaction/rpc/oteldemo.CheckoutService", name)
	assert.Equal(t, WebTransactionType, txType)
}

func TestGetTransactionMetricNameRpcServiceMethod(t *testing.T) {
	span := ptrace.NewSpan()
	span.SetKind(ptrace.SpanKindServer)
	span.Attributes().PutStr("rpc.service", "oteldemo.CheckoutService")
	span.Attributes().PutStr("rpc.method", "PlaceOrder")

	name, txType := GetTransactionMetricName(span)
	assert.Equal(t, "WebTransaction/rpc/oteldemo.CheckoutService/PlaceOrder", name)
	assert.Equal(t, WebTransactionType, txType)
}

func TestGetTransactionMetricNameBogusConsumer(t *testing.T) {
	span := ptrace.NewSpan()
	span.SetKind(ptrace.SpanKindConsumer)

	name, txType := GetTransactionMetricName(span)
	assert.Equal(t, "OtherTransaction/Consumer/unknownSystem/unknown/unknown", name)
	assert.Equal(t, OtherTransactionType, txType)
}

func TestGetTransactionMetricNameConsumer(t *testing.T) {
	span := ptrace.NewSpan()
	span.SetKind(ptrace.SpanKindConsumer)

	span.Attributes().PutStr("messaging.system", "kafka")
	span.Attributes().PutStr("messaging.destination.name", "orders")
	span.Attributes().PutStr("messaging.operation", "receive")

	name, txType := GetTransactionMetricName(span)
	assert.Equal(t, "OtherTransaction/Consumer/kafka/orders/receive", name)
	assert.Equal(t, OtherTransactionType, txType)
}

func TestGetTransactionMetricNameProducer(t *testing.T) {
	span := ptrace.NewSpan()
	span.SetKind(ptrace.SpanKindProducer)

	span.Attributes().PutStr("messaging.system", "kafka")
	span.Attributes().PutStr("messaging.destination.name", "orders")
	span.Attributes().PutStr("messaging.operation", "publish")

	// we don't name transactions with producer spans
	_, txType := GetTransactionMetricName(span)
	assert.Equal(t, NullTransactionType, txType)
}

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
	assert.Equal(t, "S", apdex.GetApdexBucket(0.1))
	assert.Equal(t, "S", apdex.GetApdexBucket(0.5))
	assert.Equal(t, "T", apdex.GetApdexBucket(0.51))
	assert.Equal(t, "T", apdex.GetApdexBucket(1.1))
	assert.Equal(t, "T", apdex.GetApdexBucket(2.0))
	assert.Equal(t, "F", apdex.GetApdexBucket(2.1))
}

func TestGetTransactionMetricNameUnknown(t *testing.T) {
	span := ptrace.NewSpan()
	span.SetKind(ptrace.SpanKindServer)

	name, txType := GetTransactionMetricName(span)
	assert.Equal(t, "WebTransaction/Other/unknown", name)
	assert.Equal(t, WebTransactionType, txType)
	assert.Equal(t, "Web", txType.AsString())
}

func TestGetTransactionMetricNameRoute(t *testing.T) {
	span := ptrace.NewSpan()
	span.SetKind(ptrace.SpanKindServer)
	span.Attributes().PutStr("http.route", "/users")

	name, txType := GetTransactionMetricName(span)
	assert.Equal(t, "WebTransaction/http.route/users", name)
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

	name, txType := GetTransactionMetricName(span)
	assert.Equal(t, "WebTransaction/Uri/owners/5", name)
	assert.Equal(t, WebTransactionType, txType)
}

func TestGetOrCreateTransaction(t *testing.T) {
	transactions := NewTransactionsMap(0.5)
	span := ptrace.NewSpan()
	meterProvider := NewMeterProvider()
	metrics := meterProvider.getOrCreateResourceMetrics(pcommon.NewMap())
	transaction, _ := transactions.GetOrCreateTransaction("java", span, metrics)

	transaction.SetRootSpan(span)
	assert.Equal(t, true, transaction.IsRootSet())
	transactions.ProcessTransactions()

	existingTransaction, _ := transactions.GetOrCreateTransaction("java", span, metrics)
	assert.Equal(t, transaction, existingTransaction)
	assert.Equal(t, true, existingTransaction.IsRootSet())
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

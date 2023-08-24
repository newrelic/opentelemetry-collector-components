// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmconnector

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

func TestConvertOneSpanToLogs(t *testing.T) {
	traces := ptrace.NewTraces()
	resourceSpans := traces.ResourceSpans().AppendEmpty()
	resourceSpans.Resource().Attributes().PutStr("service.name", "service")
	resourceSpans.Resource().Attributes().PutStr("instrumentation.provider", "newrelic-opentelemetry")
	scopeSpans := resourceSpans.ScopeSpans().AppendEmpty().Spans()
	attrs := map[string]string{
		"attrKey": "attrValue",
	}
	end := time.Now()
	start := end.Add(-time.Second)
	spanValues := []TestSpan{{Start: start, End: end, Name: "span", Kind: ptrace.SpanKindServer}}
	addSpan(scopeSpans, attrs, spanValues)
	logger, _ := zap.NewDevelopment()
	logs := BuildTransactions(logger, traces)
	assert.Equal(t, 1, logs.LogRecordCount())
}

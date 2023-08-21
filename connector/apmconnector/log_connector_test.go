// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmconnector

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func TestConvertOneSpanToLogs(t *testing.T) {
	traces := ptrace.NewTraces()
	resourceSpans := traces.ResourceSpans().AppendEmpty()
	resourceSpans.Resource().Attributes().PutStr("service.name", "service")
	scopeSpans := resourceSpans.ScopeSpans().AppendEmpty().Spans()
	attrs := map[string]string{
		"attrKey": "attrValue",
	}
	end := time.Now()
	start := end.Add(-time.Second)
	spanValues := []TestSpan{{Start: start, End: end, Name: "span", Kind: ptrace.SpanKindServer}}
	addSpan(scopeSpans, attrs, spanValues)

	logs := BuildTransactions(traces)
	assert.Equal(t, 1, logs.LogRecordCount())
}

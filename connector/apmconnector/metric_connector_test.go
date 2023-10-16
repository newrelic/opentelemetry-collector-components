// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmconnector

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

func TestConvertOneSpanToMetrics(t *testing.T) {
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
	config := Config{ApdexT: 0.5}
	metrics := ConvertTraces(logger, &config, traces)
	assert.Equal(t, 2, metrics.MetricCount())
	rm := metrics.ResourceMetrics().At(0)
	serviceName, _ := rm.Resource().Attributes().Get("service.name")
	assert.Equal(t, "service", serviceName.AsString())
	sm := rm.ScopeMetrics().At(0)

	// TODO: The commented out metrics will be generated when we generate them from spans for languages like Ruby
	// checkSumMetric(t, "apm.service.transaction.apdex", 1, sm.Metrics())
	// checkSumMetric(t, "apm.service.apdex", 1, sm.Metrics())
	checkHistogramMetric(t, "apm.service.overview.web", 1, sm.Metrics())
	// checkHistogramMetric(t, "apm.service.transaction.duration", 1, sm.Metrics())
	checkHistogramMetric(t, "apm.service.transaction.overview", 1, sm.Metrics())
}

func addSpan(spanSlice ptrace.SpanSlice, attributes map[string]string, spanValues []TestSpan) {
	for _, spanValue := range spanValues {
		span := spanSlice.AppendEmpty()
		span.SetName(spanValue.Name)
		span.SetEndTimestamp(pcommon.NewTimestampFromTime(time.Unix(spanValue.End.Unix(), 0)))
		span.SetStartTimestamp(pcommon.NewTimestampFromTime(time.Unix(spanValue.Start.Unix(), 0)))
		span.SetKind(spanValue.Kind)
		for k, v := range attributes {
			span.Attributes().PutStr(k, v)
		}
	}
}

type TestSpan struct {
	Start time.Time
	End   time.Time
	Name  string
	Kind  ptrace.SpanKind
}

func checkHistogramMetric(t *testing.T, name string, value float64, metrics pmetric.MetricSlice) {
	t.Helper()
	found := false

	for i := 0; i < metrics.Len(); i++ {
		m := metrics.At(i)
		if m.Name() == name {
			dp := m.ExponentialHistogram().DataPoints().At(0)
			assert.Equal(t, value, dp.Sum())
			found = true
			break
		}
	}

	if !found {
		assert.Fail(t, fmt.Sprintf("Could not find metric %s", name))
	}
}

func checkSumMetric(t *testing.T, name string, value int64, metrics pmetric.MetricSlice) {
	t.Helper()
	found := false

	for i := 0; i < metrics.Len(); i++ {
		m := metrics.At(i)
		if m.Name() == name {
			dp := m.Sum().DataPoints().At(0)
			assert.Equal(t, value, dp.IntValue())
			found = true
			break
		}
	}

	if !found {
		assert.Fail(t, fmt.Sprintf("Could not find metric %s", name))
	}
}

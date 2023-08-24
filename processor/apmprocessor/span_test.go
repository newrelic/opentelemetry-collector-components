// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmprocessor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processortest"
)

func TestMutateDbSpan(t *testing.T) {
	sink := new(consumertest.TracesSink)

	tp, err := createTestTracesProcessor(sink)
	require.Nil(t, err)
	require.NotNil(t, tp)

	traces := ptrace.NewTraces()
	resourceSpans := traces.ResourceSpans().AppendEmpty()
	resourceSpans.Resource().Attributes().PutStr("service.name", "service")
	resourceSpans.Resource().Attributes().PutStr("telemetry.sdk.name", "opentelemetry")
	scopeSpans := resourceSpans.ScopeSpans().AppendEmpty().Spans()
	attrs := map[string]string{
		"attrKey":      "attrValue",
		"db.statement": "select * from users",
	}
	end := time.Now()
	start := end.Add(-time.Second)
	spanValues := []TestSpan{{Start: start, End: end, Name: "span", Kind: ptrace.SpanKindServer}}
	addSpan(scopeSpans, attrs, spanValues)

	err = tp.ConsumeTraces(context.Background(), traces)
	require.Nil(t, err)
	processedResource := sink.AllTraces()[0].ResourceSpans().At(0)
	processedSpans := processedResource.ScopeSpans().At(0).Spans().At(0)
	dbtable, dbtablePresent := processedSpans.Attributes().Get(DbSQLTableAttributeName)
	assert.True(t, dbtablePresent)
	assert.Equal(t, dbtable.AsString(), "users")

	instrumentationProvider, instrumentationProviderPresent := processedResource.Resource().Attributes().Get("instrumentation.provider")
	assert.True(t, instrumentationProviderPresent)
	assert.Equal(t, instrumentationProvider.AsString(), "newrelic-opentelemetry")
}

func TestDontChangeInstrumentationProvider(t *testing.T) {
	sink := new(consumertest.TracesSink)

	tp, err := createProcessorWithConfig(sink, Config{ChangeInstrumentationProvider: false})
	require.Nil(t, err)
	require.NotNil(t, tp)

	traces := ptrace.NewTraces()
	resourceSpans := traces.ResourceSpans().AppendEmpty()
	resourceSpans.Resource().Attributes().PutStr("service.name", "service")
	resourceSpans.Resource().Attributes().PutStr("telemetry.sdk.name", "opentelemetry")
	scopeSpans := resourceSpans.ScopeSpans().AppendEmpty().Spans()
	attrs := map[string]string{
		"attrKey": "attrValue",
	}
	end := time.Now()
	start := end.Add(-time.Second)
	spanValues := []TestSpan{{Start: start, End: end, Name: "span", Kind: ptrace.SpanKindServer}}
	addSpan(scopeSpans, attrs, spanValues)

	err = tp.ConsumeTraces(context.Background(), traces)
	require.Nil(t, err)
	processedResource := sink.AllTraces()[0].ResourceSpans().At(0)
	_, instrumentationProviderPresent := processedResource.Resource().Attributes().Get("instrumentation.provider")
	assert.False(t, instrumentationProviderPresent)
}

func createTestTracesProcessor(sink consumer.Traces) (processor.Traces, error) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)
	tp, err := factory.CreateTracesProcessor(context.Background(), processortest.NewNopCreateSettings(), oCfg, sink)
	return tp, err
}

func createProcessorWithConfig(sink consumer.Traces, config Config) (processor.Traces, error) {
	factory := NewFactory()
	tp, err := factory.CreateTracesProcessor(context.Background(), processortest.NewNopCreateSettings(), &config, sink)
	return tp, err
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

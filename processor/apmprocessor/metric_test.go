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
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processortest"
)

func TestProcessMetric(t *testing.T) {
	sink := new(consumertest.MetricsSink)
	mp, err := createTestMetricsProcessor(sink)
	require.Nil(t, err)
	require.NotNil(t, mp)

	metrics := pmetric.NewMetrics()
	resourceMetrics := metrics.ResourceMetrics().AppendEmpty()
	resourceMetrics.Resource().Attributes().PutStr("service.name", "service")
	resourceMetrics.Resource().Attributes().PutStr("telemetry.sdk.name", "opentelemetry")
	scopeMetrics := resourceMetrics.ScopeMetrics().AppendEmpty()
	metric := scopeMetrics.Metrics().AppendEmpty()
	metric.SetEmptySum()
	dp := metric.Sum().DataPoints().AppendEmpty()
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	dp.SetIntValue(1)

	err = mp.ConsumeMetrics(context.Background(), metrics)
	require.Nil(t, err)

	provider, providerPresent := sink.AllMetrics()[0].ResourceMetrics().At(0).Resource().Attributes().Get("instrumentation.provider")
	assert.True(t, providerPresent)
	assert.Equal(t, "newrelic-opentelemetry", provider.AsString())
}

func createTestMetricsProcessor(sink consumer.Metrics) (processor.Metrics, error) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)
	mp, err := factory.CreateMetricsProcessor(context.Background(), processortest.NewNopCreateSettings(), oCfg, sink)

	return mp, err
}

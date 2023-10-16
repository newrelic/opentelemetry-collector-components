// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmconnector // import "github.com/newrelic/opentelemetry-collector-components/connector/apmconnector"

import (
	"context"
	"fmt"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type OpenTelemetryMetricToApmMetricConnector struct {
	config *Config
	logger *zap.Logger

	metricsConsumer consumer.Metrics
}

func (c *OpenTelemetryMetricToApmMetricConnector) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (c *OpenTelemetryMetricToApmMetricConnector) ConsumeMetrics(ctx context.Context, td pmetric.Metrics) error {
	metrics := ConvertMetrics(c.logger, c.config, td)
	return c.metricsConsumer.ConsumeMetrics(ctx, metrics)
}

func (c *OpenTelemetryMetricToApmMetricConnector) Start(_ context.Context, _ component.Host) error {
	c.logger.Info("Starting the APM Metric Connector")
	if c.config.ApdexT == 0 {
		c.config.ApdexT = defaultApdexT
	}
	return nil
}

func (c *OpenTelemetryMetricToApmMetricConnector) Shutdown(context.Context) error {
	c.logger.Info("Stopping the APM Metric Connector")
	return nil
}

func ConvertMetrics(logger *zap.Logger, config *Config, md pmetric.Metrics) pmetric.Metrics {
	apdex := NewApdex(config.ApdexT)
	newMetrics := pmetric.NewMetrics()
	attributesFilter := NewAttributeFilter()
	metricMap := NewMetrics()

	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)

		rmNew := pmetric.ResourceMetrics{}
		metrics := &ResourceMetrics{}

		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)

			smNew := pmetric.ScopeMetrics{}

			for k := 0; k < sm.Metrics().Len(); k++ {
				m := sm.Metrics().At(k)

				if m.Name() == "http.server.duration" || m.Name() == "http.server.request.duration" {
					rmNew, smNew, metrics = createResourceAndScopeMetrics(logger, rmNew, attributesFilter, rm, newMetrics, metrics, metricMap, smNew)
					recordTransactionMetrics(logger, m, metrics, apdex, smNew)
				} else if m.Name() == "http.client.duration" || m.Name() == "http.client.request.duration" {
					rmNew, smNew, metrics = createResourceAndScopeMetrics(logger, rmNew, attributesFilter, rm, newMetrics, metrics, metricMap, smNew)
					recordExternalHostDurationMetric(logger, m, smNew)
				}
			}
		}
	}

	if newMetrics.ResourceMetrics().Len() > 0 {
		metricMap.AppendOtelMetrics(newMetrics)
	}

	return newMetrics
}

func recordExternalHostDurationMetric(logger *zap.Logger, m pmetric.Metric, smNew pmetric.ScopeMetrics) {
	newMetric := pmetric.NewMetric()
	newMetric.SetName("apm.service.external.host.duration")
	newMetric.SetDescription("Duration of external calls")
	newMetric.SetUnit(m.Unit())

	switch metricType := m.Type(); metricType {
	case pmetric.MetricTypeHistogram:
		newMetric.SetEmptyHistogram().DataPoints().EnsureCapacity(3)
		for i := 0; i < m.Histogram().DataPoints().Len(); i++ {
			dp := m.Histogram().DataPoints().At(i)
			if serverAddress, hasServerAddress := GetServerAddress(dp.Attributes()); hasServerAddress {
				newDp := newMetric.Histogram().DataPoints().AppendEmpty()
				dp.CopyTo(newDp)
				newDp.Attributes().PutStr("server.address", serverAddress)
				newDp.Attributes().PutStr("external.host", serverAddress)
				newDp.Attributes().PutStr("metricTimesliceName", fmt.Sprintf("External/%s/all", serverAddress))
			}
		}
		newMetric.Histogram().SetAggregationTemporality(m.Histogram().AggregationTemporality())
		newMetric.CopyTo(smNew.Metrics().AppendEmpty())
	case pmetric.MetricTypeExponentialHistogram:
		newMetric.SetEmptyExponentialHistogram().DataPoints().EnsureCapacity(3)
		for i := 0; i < m.ExponentialHistogram().DataPoints().Len(); i++ {
			dp := m.ExponentialHistogram().DataPoints().At(i)
			if serverAddress, hasServerAddress := GetServerAddress(dp.Attributes()); hasServerAddress {
				newDp := newMetric.ExponentialHistogram().DataPoints().AppendEmpty()
				dp.CopyTo(newDp)
				newDp.Attributes().PutStr("server.address", serverAddress)
				newDp.Attributes().PutStr("external.host", serverAddress)
				newDp.Attributes().PutStr("metricTimesliceName", fmt.Sprintf("External/%s/all", serverAddress))
			}
		}
		newMetric.ExponentialHistogram().SetAggregationTemporality(m.ExponentialHistogram().AggregationTemporality())
		newMetric.CopyTo(smNew.Metrics().AppendEmpty())
	default:
		logger.Error("unexpected metric type", zap.String("name", m.Name()), zap.String("type", metricType.String()))
	}
}

func recordTransactionMetrics(logger *zap.Logger, m pmetric.Metric, metrics *ResourceMetrics, apdex Apdex, smNew pmetric.ScopeMetrics) {
	newMetric := pmetric.NewMetric()
	newMetric.SetName("apm.service.transaction.duration")
	newMetric.SetDescription("Duration of the transaction")
	newMetric.SetUnit(m.Unit())

	switch metricType := m.Type(); metricType {
	case pmetric.MetricTypeHistogram:
		newMetric.SetEmptyHistogram().DataPoints().EnsureCapacity(3)
		for i := 0; i < m.Histogram().DataPoints().Len(); i++ {
			dp := m.Histogram().DataPoints().At(i)
			newDp := newMetric.Histogram().DataPoints().AppendEmpty()
			dp.CopyTo(newDp)
			name, txType := GetTransactionMetricNameFromAttributes(dp.Attributes())
			newDp.Attributes().Clear()
			newDp.Attributes().PutStr("transactionType", txType.AsString())
			newDp.Attributes().PutStr("transactionName", name)
			newDp.Attributes().PutStr("metricTimesliceName", name)

			isError := ContainsErrorHTTPStatusCode(dp.Attributes())
			if isError {
				{
					attributes := pcommon.NewMap()
					attributes.PutStr("transactionType", txType.AsString())
					sum := metrics.GetSum("apm.service.error.count", attributes, true, dp.StartTimestamp(), dp.Timestamp())
					sum.Add(int64(newDp.Count()), dp.StartTimestamp(), dp.Timestamp())
				}
				{
					attributes := pcommon.NewMap()
					attributes.PutStr("transactionType", txType.AsString())
					attributes.PutStr("transactionName", name)
					sum := metrics.GetSum("apm.service.transaction.error.count", attributes, true, dp.StartTimestamp(), dp.Timestamp())
					sum.Add(int64(newDp.Count()), dp.StartTimestamp(), dp.Timestamp())
				}

				generateApdexMetrics(apdex, "F", metrics, dp.StartTimestamp(), dp.Timestamp(), int64(dp.Count()), name)
			} else {
				s, t, f := GetApdexFromExplicitHistogramBounds(newDp.ExplicitBounds().AsRaw(), newDp.BucketCounts().AsRaw(), m.Unit(), apdex)

				if s > 0 {
					generateApdexMetrics(apdex, "S", metrics, dp.StartTimestamp(), dp.Timestamp(), int64(s), name)
				}

				if t > 0 {
					generateApdexMetrics(apdex, "T", metrics, dp.StartTimestamp(), dp.Timestamp(), int64(t), name)
				}

				if f > 0 {
					generateApdexMetrics(apdex, "F", metrics, dp.StartTimestamp(), dp.Timestamp(), int64(f), name)
				}
			}
		}
		newMetric.Histogram().SetAggregationTemporality(m.Histogram().AggregationTemporality())
		newMetric.CopyTo(smNew.Metrics().AppendEmpty())
	case pmetric.MetricTypeExponentialHistogram:
		newMetric.SetEmptyExponentialHistogram().DataPoints().EnsureCapacity(3)
		for i := 0; i < m.ExponentialHistogram().DataPoints().Len(); i++ {
			dp := m.ExponentialHistogram().DataPoints().At(i)
			newDp := newMetric.ExponentialHistogram().DataPoints().AppendEmpty()
			dp.CopyTo(newDp)
			name, txType := GetTransactionMetricNameFromAttributes(dp.Attributes())
			newDp.Attributes().Clear()
			newDp.Attributes().PutStr("transactionType", txType.AsString())
			newDp.Attributes().PutStr("transactionName", name)
			newDp.Attributes().PutStr("metricTimesliceName", name)

			isError := ContainsErrorHTTPStatusCode(dp.Attributes())
			if isError {
				{
					attributes := pcommon.NewMap()
					attributes.PutStr("transactionType", txType.AsString())
					sum := metrics.GetSum("apm.service.error.count", attributes, true, dp.StartTimestamp(), dp.Timestamp())
					sum.Add(int64(newDp.Count()), dp.StartTimestamp(), dp.Timestamp())
				}
				{
					attributes := pcommon.NewMap()
					attributes.PutStr("transactionType", txType.AsString())
					attributes.PutStr("transactionName", name)
					sum := metrics.GetSum("apm.service.transaction.error.count", attributes, true, dp.StartTimestamp(), dp.Timestamp())
					sum.Add(int64(newDp.Count()), dp.StartTimestamp(), dp.Timestamp())
				}

				generateApdexMetrics(apdex, "F", metrics, dp.StartTimestamp(), dp.Timestamp(), int64(dp.Count()), name)
			}

			// TODO: Generate apdex metrics for exponential histograms.
		}
		newMetric.ExponentialHistogram().SetAggregationTemporality(m.ExponentialHistogram().AggregationTemporality())
		newMetric.CopyTo(smNew.Metrics().AppendEmpty())
	default:
		logger.Error("unexpected metric type", zap.String("name", m.Name()), zap.String("type", metricType.String()))
	}
}

func createResourceAndScopeMetrics(logger *zap.Logger, rmNew pmetric.ResourceMetrics, attributesFilter *AttributeFilter, rm pmetric.ResourceMetrics, newMetrics pmetric.Metrics, metrics *ResourceMetrics, metricMap Metrics, smNew pmetric.ScopeMetrics) (pmetric.ResourceMetrics, pmetric.ScopeMetrics, *ResourceMetrics) {
	if rmNew == (pmetric.ResourceMetrics{}) {
		resourceAttributes, err := attributesFilter.FilterAttributes(rm.Resource().Attributes())
		if err != nil {
			logger.Error("Could not filter resource attributes", zap.String("error", err.Error()))
		}
		rmNew = newMetrics.ResourceMetrics().AppendEmpty()
		resourceAttributes.CopyTo(rmNew.Resource().Attributes())

		// TODO: should we declare a New Relic specific schema?
		// rmNew.SetSchemaUrl(rm.SchemaUrl())

		metrics = metricMap.GetOrCreateResource(resourceAttributes)
	}

	if smNew == (pmetric.ScopeMetrics{}) {
		smNew = rmNew.ScopeMetrics().AppendEmpty()

		// TODO: do we want any of the scope attributes? or the schema?
		// smNew.SetSchemaUrl(sm.SchemaUrl())
		// sm.Scope().CopyTo(smNew.Scope())
	}

	return rmNew, smNew, metrics
}

func generateApdexMetrics(apdex Apdex, zone string, resourceMetrics *ResourceMetrics, startTimestamp pcommon.Timestamp, timestamp pcommon.Timestamp, count int64, transactionName string) {
	attributes := pcommon.NewMap()
	attributes.PutDouble("apdex.value", apdex.apdexSatisfying)
	attributes.PutStr("apdex.zone", zone)
	attributes.PutStr("transactionType", WebTransactionType.AsString())

	apdexMetric := resourceMetrics.GetSum("apm.service.apdex", attributes, true, startTimestamp, timestamp)
	apdexMetric.Add(count, startTimestamp, timestamp)

	attributes.PutStr("transactionName", transactionName)

	transactionApdexMetric := resourceMetrics.GetSum("apm.service.transaction.apdex", attributes, true, startTimestamp, timestamp)
	transactionApdexMetric.Add(count, startTimestamp, timestamp)
}

// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmconnector // import "github.com/newrelic/opentelemetry-collector-components/connector/apmconnector"

import (
	"context"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

//const defaultApdexT = 0.5

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
	newMetrics := pmetric.NewMetrics()
	attributesFilter := NewAttributeFilter()

	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rs := md.ResourceMetrics().At(i)

		resourceAttributes, err := attributesFilter.FilterAttributes(rs.Resource().Attributes())
		if err != nil {
			logger.Error("Could not filter resource attributes", zap.String("error", err.Error()))
		}
		rms := newMetrics.ResourceMetrics().AppendEmpty()
		resourceAttributes.CopyTo(rms.Resource().Attributes())

		for j := 0; j < rs.ScopeMetrics().Len(); j++ {
			scopeMetric := rs.ScopeMetrics().At(j)
			sm := rms.ScopeMetrics().AppendEmpty()
			scopeMetric.Scope().Attributes().CopyTo(sm.Scope().Attributes())

			for k := 0; k < scopeMetric.Metrics().Len(); k++ {
				metric := scopeMetric.Metrics().At(k)
				if metric.Name() == "http.server.duration" || metric.Name() == "http.server.request.duration" {
					newMetric := pmetric.NewMetric()
					newMetric.SetName("newrelic.apm.service.transaction.duration")
					newMetric.SetDescription("Duration of the transaction")
					newMetric.SetUnit("s")

					switch metricType := metric.Type(); metricType {
					case pmetric.MetricTypeHistogram:
						newMetric.SetEmptyHistogram().DataPoints().EnsureCapacity(3)
						for i := 0; i < metric.Histogram().DataPoints().Len(); i++ {
							srcDp := metric.Histogram().DataPoints().At(i)
							dp := newMetric.Histogram().DataPoints().AppendEmpty()
							srcDp.CopyTo(dp)
							name, _ := GetTransactionMetricNameFromHttpServerDurationAttributes(srcDp.Attributes())
							dp.Attributes().Clear()
							dp.Attributes().PutStr("transactionType", "Web")
							dp.Attributes().PutStr("transactionName", name)
							dp.Attributes().PutStr("metricTimesliceName", name)
						}
						newMetric.Histogram().SetAggregationTemporality(metric.Histogram().AggregationTemporality())
						newMetric.CopyTo(sm.Metrics().AppendEmpty())
					case pmetric.MetricTypeExponentialHistogram:
						newMetric.SetEmptyExponentialHistogram().DataPoints().EnsureCapacity(3)
						for i := 0; i < metric.ExponentialHistogram().DataPoints().Len(); i++ {
							srcDp := metric.ExponentialHistogram().DataPoints().At(i)
							dp := newMetric.ExponentialHistogram().DataPoints().AppendEmpty()
							srcDp.CopyTo(dp)
							name, _ := GetTransactionMetricNameFromHttpServerDurationAttributes(srcDp.Attributes())
							dp.Attributes().Clear()
							dp.Attributes().PutStr("transactionType", "Web")
							dp.Attributes().PutStr("transactionName", name)
							dp.Attributes().PutStr("metricTimesliceName", name)
						}
						newMetric.ExponentialHistogram().SetAggregationTemporality(metric.ExponentialHistogram().AggregationTemporality())
						newMetric.CopyTo(sm.Metrics().AppendEmpty())
					default:
						logger.Error("unexpected metric type", zap.String("name", metric.Name()), zap.String("type", metricType.String()))
					}
				}
			}
		}
	}

	return newMetrics
}

// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmconnector // import "github.com/newrelic/opentelemetry-collector-components/connector/apmconnector"

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

const defaultApdexT = 0.5

type ApmMetricConnector struct {
	config *Config
	logger *zap.Logger

	metricsConsumer consumer.Metrics
}

func (c *ApmMetricConnector) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (c *ApmMetricConnector) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	metrics := ConvertTraces(c.logger, c.config, td)
	return c.metricsConsumer.ConsumeMetrics(ctx, metrics)
}

func (c *ApmMetricConnector) Start(_ context.Context, _ component.Host) error {
	c.logger.Info("Starting the APM Metric Connector")
	if c.config.ApdexT == 0 {
		c.config.ApdexT = defaultApdexT
	}
	return nil
}

func (c *ApmMetricConnector) Shutdown(context.Context) error {
	c.logger.Info("Stopping the APM Metric Connector")
	return nil
}

func ConvertTraces(logger *zap.Logger, config *Config, td ptrace.Traces) pmetric.Metrics {
	attributesFilter := NewAttributeFilter()
	transactions := NewTransactionsMap(config.ApdexT)
	meterProvider := NewMeterProvider()

	for i := 0; i < td.ResourceSpans().Len(); i++ {
		rs := td.ResourceSpans().At(i)
		instrumentationProvider, instrumentationProviderPresent := rs.Resource().Attributes().Get("instrumentation.provider")
		if instrumentationProviderPresent && instrumentationProvider.AsString() != "opentelemetry" {
			logger.Debug("Skipping resource spans", zap.String("instrumentation.provider", instrumentationProvider.AsString()))
			continue
		}

		resourceAttributes, err := attributesFilter.FilterAttributes(rs.Resource().Attributes())
		if err != nil {
			logger.Error("Could not filter resource attributes", zap.String("error", err.Error()))
		}
		resourceMetrics := meterProvider.getOrCreateResourceMetrics(resourceAttributes)

		sdkLanguage := GetSdkLanguage(rs.Resource().Attributes())
		for j := 0; j < rs.ScopeSpans().Len(); j++ {
			scopeSpan := rs.ScopeSpans().At(j)
			for k := 0; k < scopeSpan.Spans().Len(); k++ {
				span := scopeSpan.Spans().At(k)
				if k == 0 {
					if hostName, exists := resourceAttributes.Get("host.name"); exists {
						GenerateInstanceMetric(resourceMetrics, hostName.AsString(), span.EndTimestamp())
					}
				}

				transaction, _ := transactions.GetOrCreateTransaction(sdkLanguage, span, resourceMetrics, rs.Resource().Attributes())

				transaction.AddSpan(span)
			}
		}
	}

	transactions.ProcessTransactions()

	return meterProvider.Metrics
}

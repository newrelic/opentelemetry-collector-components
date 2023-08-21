// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmconnector // import "apmconnector"

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

type ApmTraceConnector struct {
	config    *Config
	logger    *zap.Logger
	sqlparser *SQLParser

	tracesConsumer consumer.Traces
}

func (c *ApmTraceConnector) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

func (c *ApmTraceConnector) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	MutateSpans(c.logger, c.sqlparser, td)
	return c.tracesConsumer.ConsumeTraces(ctx, td)
}

func (c *ApmTraceConnector) Start(_ context.Context, _ component.Host) error {
	c.logger.Info("Starting the New Relic APM Trace Connector")
	if c.config.ApdexT == 0 {
		c.config.ApdexT = defaultApdexT
	}
	return nil
}

func (c *ApmTraceConnector) Shutdown(context.Context) error {
	c.logger.Info("Stopping the APM Trace Connector")
	return nil
}

func MutateSpans(logger *zap.Logger, sqlparser *SQLParser, td ptrace.Traces) {
	for i := 0; i < td.ResourceSpans().Len(); i++ {
		rs := td.ResourceSpans().At(i)
		instrumentationProvider, instrumentationProviderPresent := rs.Resource().Attributes().Get("instrumentation.provider")
		if instrumentationProviderPresent && instrumentationProvider.AsString() != "opentelemetry" {
			logger.Debug("Skipping resource spans", zap.String("instrumentation.provider", instrumentationProvider.AsString()))
			continue
		}

		for j := 0; j < rs.ScopeSpans().Len(); j++ {
			scopeSpan := rs.ScopeSpans().At(j)
			for k := 0; k < scopeSpan.Spans().Len(); k++ {
				span := scopeSpan.Spans().At(k)

				if parsedTable, parsed := sqlparser.ParseDbTableFromSpan(span); parsed {
					span.Attributes().PutStr(DbSQLTableAttributeName, parsedTable)
				}
			}
		}
	}
}

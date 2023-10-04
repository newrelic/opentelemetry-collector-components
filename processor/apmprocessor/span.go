// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmprocessor // import "github.com/newrelic/opentelemetry-collector-components/processor/apmprocessor"

import (
	"context"

	"github.com/newrelic/opentelemetry-collector-components/connector/apmconnector"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

type spanProcessor struct {
	sqlparser *SQLParser
	logger    *zap.Logger
	config    Config
}

func newSpanProcessor(config Config, logger *zap.Logger) (*spanProcessor, error) {
	return &spanProcessor{config: config, sqlparser: NewSQLParser(), logger: logger}, nil
}

func (sp *spanProcessor) processTraces(_ context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	for i := 0; i < td.ResourceSpans().Len(); i++ {
		rs := td.ResourceSpans().At(i)
		instrumentationProvider, instrumentationProviderPresent := rs.Resource().Attributes().Get("instrumentation.provider")
		if instrumentationProviderPresent && instrumentationProvider.AsString() != "opentelemetry" {
			sp.logger.Debug("Skipping resource spans", zap.String("instrumentation.provider", instrumentationProvider.AsString()))
			continue
		}

		setInstrumentationProvider(sp.config, rs.Resource())

		for j := 0; j < rs.ScopeSpans().Len(); j++ {
			scopeSpan := rs.ScopeSpans().At(j)
			for k := 0; k < scopeSpan.Spans().Len(); k++ {
				span := scopeSpan.Spans().At(k)
				if span.Kind() == ptrace.SpanKindServer || span.Kind() == ptrace.SpanKindConsumer {
					transactionName, transactionType := apmconnector.GetTransactionMetricName(span)
					span.Attributes().PutStr("transaction.type", transactionType.AsString())
					span.Attributes().PutStr("transaction.name", transactionName)
				} else if parsedTable, parsed := sp.sqlparser.ParseDbTableFromSpan(span); parsed {
					span.Attributes().PutStr(DbSQLTableAttributeName, parsedTable)
					if dbSystem, dbSystemPresent := span.Attributes().Get(DbSystemAttributeName); dbSystemPresent && dbSystem.AsString() == "redis" {
						if _, dbOperationPresent := span.Attributes().Get(DbOperationAttributeName); !dbOperationPresent {
							span.Attributes().PutStr(DbOperationAttributeName, span.Name())
						}
					}
				}
			}
		}
	}
	return td, nil
}

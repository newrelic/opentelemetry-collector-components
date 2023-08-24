// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmprocessor // import "github.com/newrelic/opentelemetry-collector-components/processor/apmprocessor"

import (
	"context"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

type logProcessor struct {
	logger *zap.Logger
	config Config
}

func newLogProcessor(config Config, logger *zap.Logger) *logProcessor {
	return &logProcessor{config: config, logger: logger}
}

func (lp *logProcessor) processLogs(_ context.Context, ld plog.Logs) (plog.Logs, error) {
	for i := 0; i < ld.ResourceLogs().Len(); i++ {
		rs := ld.ResourceLogs().At(i)
		setInstrumentationProvider(lp.config, rs.Resource())
	}
	return ld, nil
}

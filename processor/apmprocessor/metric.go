// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmprocessor // import "github.com/newrelic/opentelemetry-collector-components/processor/apmprocessor"

import (
	"context"

	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type metricProcessor struct {
	logger *zap.Logger
	config Config
}

func newMetricProcessor(config Config, logger *zap.Logger) *metricProcessor {
	return &metricProcessor{config: config, logger: logger}
}

func (mp *metricProcessor) processMetrics(_ context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rs := md.ResourceMetrics().At(i)
		setInstrumentationProvider(mp.config, rs.Resource())
	}
	return md, nil
}

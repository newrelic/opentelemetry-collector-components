// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmconnector // import "github.com/newrelic/opentelemetry-collector-components/connector/apmconnector"

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.uber.org/zap"
)

func ShouldProcess(logger *zap.Logger, rs pcommon.Resource) bool {
	instrumentationProvider, instrumentationProviderPresent := rs.Attributes().Get("instrumentation.provider")
	if !instrumentationProviderPresent || instrumentationProvider.AsString() != "newrelic-opentelemetry" {
		logger.Debug("Skipping resource spans", zap.String("instrumentation.provider", instrumentationProvider.AsString()))
		return false
	}
	return true
}

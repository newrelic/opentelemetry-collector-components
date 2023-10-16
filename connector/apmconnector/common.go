// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmconnector // import "github.com/newrelic/opentelemetry-collector-components/connector/apmconnector"

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.uber.org/zap"
	"math"
)

func ShouldProcess(logger *zap.Logger, rs pcommon.Resource) bool {
	instrumentationProvider, instrumentationProviderPresent := rs.Attributes().Get("instrumentation.provider")
	if !instrumentationProviderPresent || instrumentationProvider.AsString() != "newrelic-opentelemetry" {
		logger.Debug("Skipping resource spans")
		return false
	}
	return true
}

func ContainsErrorHTTPStatusCode(attributes pcommon.Map) bool {
	statusCodeValue, statusCodeKey := GetFirst(attributes, []string{"http.request.status_code", "http.status_code"})
	if statusCodeKey != "" && statusCodeValue.Type() == pcommon.ValueTypeInt {
		return statusCodeValue.Int() >= 500
	}
	return false
}

func GetApdexFromExplicitHistogramBounds(bounds []float64, bucketCounts []uint64, unit string, apdex Apdex) (uint64, uint64, uint64) {
	if unit != "s" && unit != "ms" {
		return 0, 0, 0
	}

	satisfying := apdex.apdexSatisfying
	tolerating := apdex.apdexTolerating
	if unit == "ms" {
		satisfying = satisfying * 1000
		tolerating = tolerating * 1000
	}

	var s uint64
	var t uint64
	var f uint64

	for i := 0; i < len(bucketCounts); i++ {
		count := bucketCounts[i]

		var upper float64
		if i < len(bucketCounts)-1 {
			upper = bounds[i]
		} else {
			upper = math.Inf(1)
		}

		if upper <= satisfying {
			s += count
		} else if upper <= tolerating {
			t += count
		} else {
			f += count
		}
	}

	return s, t, f
}

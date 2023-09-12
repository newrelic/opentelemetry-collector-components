// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmconnector // import "github.com/newrelic/opentelemetry-collector-components/connector/apmconnector"

import (
	"github.com/lightstep/go-expohisto/structure"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type Histogram struct {
	histogram *structure.Histogram[float64]
}

func NewHistogram() Histogram {
	h := new(structure.Histogram[float64])
	h.Init(structure.NewConfig())
	return Histogram{
		histogram: h,
	}
}

func (h *Histogram) Update(value float64) {
	h.histogram.Update(value)
}

// Copied from https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/connector/spanmetricsconnector/internal/metrics/metrics.go
func (h *Histogram) AddDatapointToHistogram(dp pmetric.ExponentialHistogramDataPoint) {
	agg := h.histogram
	dp.SetCount(agg.Count())
	dp.SetSum(agg.Sum())
	if agg.Count() != 0 {
		dp.SetMin(agg.Min())
		dp.SetMax(agg.Max())
	}

	dp.SetZeroCount(agg.ZeroCount())
	dp.SetScale(agg.Scale())

	for _, half := range []struct {
		inFunc  func() *structure.Buckets
		outFunc func() pmetric.ExponentialHistogramDataPointBuckets
	}{
		{agg.Positive, dp.Positive},
		{agg.Negative, dp.Negative},
	} {
		in := half.inFunc()
		out := half.outFunc()
		out.SetOffset(in.Offset())
		out.BucketCounts().EnsureCapacity(int(in.Len()))

		for i := uint32(0); i < in.Len(); i++ {
			out.BucketCounts().Append(in.At(i))
		}
	}
}

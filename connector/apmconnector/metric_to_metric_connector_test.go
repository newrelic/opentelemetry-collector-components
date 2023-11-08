// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmconnector

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestExponentialHistogramUnitConversion(t *testing.T) {
	originalMetric := pmetric.NewMetric()
	originalMetric.SetName("http.server.duration")
	originalMetric.SetUnit("ms")
	originalMetric.SetEmptyExponentialHistogram()

	eh := originalMetric.ExponentialHistogram()
	eh.SetAggregationTemporality(pmetric.AggregationTemporalityDelta)
	eh.DataPoints().AppendEmpty()

	now := time.Now()
	dp := originalMetric.ExponentialHistogram().DataPoints().At(0)
	dp.SetScale(1)
	dp.SetMin(0.4231)
	dp.SetMax(60.2976)
	dp.SetSum(75.988)
	dp.SetZeroCount(0)
	dp.SetStartTimestamp(pcommon.Timestamp(now.UnixNano()))
	dp.SetTimestamp(pcommon.Timestamp(now.Add(time.Minute).UnixNano()))

	// The boundaries of a given bucket of an exponential histogram is a formula
	// of the histogram's scale and the offset of the initial bucket.
	//   Ihe base is computed as: 2^2^(-scale)
	//   The exclusive lower bound of a bucket at a given index: base^index
	//   The inclusive upper bound of a bucket at a given index: base^(index+1)
	dp.Positive().SetOffset(-3)
	dp.Positive().BucketCounts().FromRaw(make([]uint64, 15))
	dp.Positive().BucketCounts().SetAt(0, 1)  // (0.353553, 0.500000] (index -3)
	dp.Positive().BucketCounts().SetAt(1, 0)  // (0.500000, 0.707107] (index -2)
	dp.Positive().BucketCounts().SetAt(2, 2)  // (0.707107, 1.000000] (index -1)
	dp.Positive().BucketCounts().SetAt(3, 3)  // (1.000000, 1.414214] (index 0)
	dp.Positive().BucketCounts().SetAt(4, 0)  // (1.414214, 2.000000] (index 1)
	dp.Positive().BucketCounts().SetAt(5, 4)  // (2.000000, 2.828427] (index 2)
	dp.Positive().BucketCounts().SetAt(6, 0)  // (2.828427, 4.000000] (index 3)
	dp.Positive().BucketCounts().SetAt(7, 0)  // (4.000000, 5.656854] (index 4)
	dp.Positive().BucketCounts().SetAt(8, 0)  // (5.656854, 8.000000] (index 5)
	dp.Positive().BucketCounts().SetAt(9, 0)  // (8.000000, 11.313708] (index 6)
	dp.Positive().BucketCounts().SetAt(10, 0) // (11.313708, 16.000000] (index 7)
	dp.Positive().BucketCounts().SetAt(11, 0) // (16.000000, 22.627417] (index 8)
	dp.Positive().BucketCounts().SetAt(12, 0) // (22.627417, 32.000000] (index 9)
	dp.Positive().BucketCounts().SetAt(13, 0) // (32.000000, 45.254834] (index 10)
	dp.Positive().BucketCounts().SetAt(14, 1) // (45.254834, 64.000000] (index 11)

	newMetric := pmetric.NewMetric()
	conversionFactor := setUnitAndComputeConversionFactor(newMetric, originalMetric.Unit())
	originalMetric.CopyTo(newMetric)
	newDp := newMetric.ExponentialHistogram().DataPoints().At(0)
	convertUnitsExponentialHistogramDataPoint(newDp, conversionFactor)

	assert.Equal(t, int32(1), newDp.Scale())
	assert.Equal(t, .0004231, newDp.Min())
	assert.Equal(t, .06029760000000001, newDp.Max())
	assert.Equal(t, .075988, newDp.Sum())
	assert.Equal(t, uint64(0), newDp.ZeroCount())
	assert.Equal(t, int32(-23), newDp.Positive().Offset())
	assert.Equal(t, uint64(1), dp.Positive().BucketCounts().At(0))  // (0.000345, 0.000488] (index -23)
	assert.Equal(t, uint64(0), dp.Positive().BucketCounts().At(1))  // (0.000488, 0.000691] (index -22)
	assert.Equal(t, uint64(2), dp.Positive().BucketCounts().At(2))  // (0.000691, 0.000977] (index -21)
	assert.Equal(t, uint64(3), dp.Positive().BucketCounts().At(3))  // (0.000977, 0.001381] (index -20)
	assert.Equal(t, uint64(0), dp.Positive().BucketCounts().At(4))  // (0.001381, 0.001953] (index -19)
	assert.Equal(t, uint64(4), dp.Positive().BucketCounts().At(5))  // (0.001953, 0.002762] (index -18)
	assert.Equal(t, uint64(0), dp.Positive().BucketCounts().At(6))  // (0.002762, 0.003906] (index -17)
	assert.Equal(t, uint64(0), dp.Positive().BucketCounts().At(7))  // (0.003906, 0.005524] (index -16)
	assert.Equal(t, uint64(0), dp.Positive().BucketCounts().At(8))  // (0.005524, 0.007813] (index -15)
	assert.Equal(t, uint64(0), dp.Positive().BucketCounts().At(9))  // (0.007813, 0.011049] (index -14)
	assert.Equal(t, uint64(0), dp.Positive().BucketCounts().At(10)) // (0.011049, 0.015625] (index -13)
	assert.Equal(t, uint64(0), dp.Positive().BucketCounts().At(11)) // (0.015625, 0.022097] (index -12)
	assert.Equal(t, uint64(0), dp.Positive().BucketCounts().At(12)) // (0.022097, 0.031250] (index -11)
	assert.Equal(t, uint64(0), dp.Positive().BucketCounts().At(13)) // (0.031250, 0.044194] (index -10)
	assert.Equal(t, uint64(1), dp.Positive().BucketCounts().At(14)) // (0.044194, 0.062500] (index -9)
}

// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmconnector

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestApdexFromExplicitHistogramBounds(t *testing.T) {
	var tests = []struct {
		name         string
		boundaries   []float64
		bucketCounts []uint64
		unit         string
		apdex        Apdex
		s            uint64
		t            uint64
		f            uint64
	}{
		{
			"Default bounds milliseconds",
			[]float64{0, 5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000},
			[]uint64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			"ms",
			NewApdex(0.5),
			9,
			2,
			5,
		},
		{
			"Default bounds seconds",
			[]float64{0, .005, .01, .025, .05, .075, .1, .25, .5, .750, 1, 2.5, 5, 7.5, 10},
			[]uint64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			"s",
			NewApdex(0.5),
			9,
			2,
			5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			satisfying, tolerating, failing := GetApdexFromExplicitHistogramBounds(tt.boundaries, tt.bucketCounts, tt.unit, tt.apdex)
			assert.Equal(t, tt.s, satisfying)
			assert.Equal(t, tt.t, tolerating)
			assert.Equal(t, tt.f, failing)
		})
	}
}

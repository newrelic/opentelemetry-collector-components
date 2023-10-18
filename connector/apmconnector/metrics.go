// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmconnector // import "github.com/newrelic/opentelemetry-collector-components/connector/apmconnector"

import (
	"crypto"
	"fmt"
	"sort"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// Metrics is a data structure used by the connector while it is
// processing spans. Once the processing is done, the map is converted
// into OTEL metrics
// The map roughly follows the structure of an OTEL resource metrics:
// resource -> scope -> metric -> datapoints

type Metrics map[string]*ResourceMetrics

func NewMetrics() Metrics {
	return make(Metrics)
}

func (metrics *Metrics) AppendOtelMetrics(dest pmetric.Metrics) pmetric.Metrics {
	otelMetrics := dest
	for _, rm := range *metrics {
		resourceMetrics := otelMetrics.ResourceMetrics().AppendEmpty()
		rm.attributes.CopyTo(resourceMetrics.Resource().Attributes())
		for _, sm := range rm.scopeMetrics {
			scopeMetrics := resourceMetrics.ScopeMetrics().AppendEmpty()
			sm.origin.CopyTo(scopeMetrics.Scope())
			for _, m := range sm.metrics {
				addMetricToScope(*m, scopeMetrics)
			}
		}
	}
	return otelMetrics
}

func addMetricToScope(metric Metric, scopeMetrics pmetric.ScopeMetrics) {
	otelMetric := scopeMetrics.Metrics().AppendEmpty()
	otelMetric.SetName(metric.metricName)
	otelMetric.SetUnit(metric.unit)

	if len(metric.histogramDatapoints) > 0 {
		histogram := otelMetric.SetEmptyExponentialHistogram()
		histogram.SetAggregationTemporality(pmetric.AggregationTemporalityDelta)
		otelDatapoints := histogram.DataPoints()
		for _, dp := range metric.histogramDatapoints {
			histoDp := otelDatapoints.AppendEmpty()
			dp.histogram.AddDatapointToHistogram(histoDp)
			histoDp.SetStartTimestamp(dp.startTimestamp)
			histoDp.SetTimestamp(dp.timestamp)
			dp.attributes.CopyTo(histoDp.Attributes())
		}
	}

	if len(metric.sumDatapoints) > 0 {
		sum := otelMetric.SetEmptySum()
		sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
		sum.SetIsMonotonic(false)
		otelDatapoints := sum.DataPoints()
		for _, dp := range metric.sumDatapoints {
			// TODO: This is a little awkward at the moment. IsMonotonic should be declared at the metric level not at the datapoint.
			if dp.isMonotonic {
				sum.SetAggregationTemporality(pmetric.AggregationTemporalityDelta)
				sum.SetIsMonotonic(true)
			}
			sumDp := otelDatapoints.AppendEmpty()
			sumDp.SetTimestamp(dp.timestamp)
			sumDp.SetStartTimestamp(dp.startTimestamp)
			dp.attributes.CopyTo(sumDp.Attributes())
			sumDp.SetIntValue(dp.value)
		}
	}
}

func (metrics *Metrics) GetOrCreateResource(attributes pcommon.Map) *ResourceMetrics {
	key := getKeyFromMap(attributes)
	res, resourcePresent := (*metrics)[key]
	if resourcePresent {
		return res
	}
	res = &ResourceMetrics{
		attributes:   attributes,
		scopeMetrics: make(map[string]*ScopeMetrics),
	}
	(*metrics)[key] = res
	return res
}

type ResourceMetrics struct {
	attributes   pcommon.Map
	scopeMetrics map[string]*ScopeMetrics
}

func (rm *ResourceMetrics) GetOrCreateScope(scope pcommon.InstrumentationScope) *ScopeMetrics {
	key := getKeyFromMap(scope.Attributes())
	scopeMetrics, scopeMetricsPresent := rm.scopeMetrics[key]
	if scopeMetricsPresent {
		return scopeMetrics
	}
	scopeMetrics = &ScopeMetrics{
		origin:  scope,
		metrics: make(map[string]*Metric),
	}
	rm.scopeMetrics[key] = scopeMetrics
	return scopeMetrics
}

func (rm *ResourceMetrics) AddHistogram(metricName string, attributes pcommon.Map, startTimestamp pcommon.Timestamp, endTimestamp pcommon.Timestamp, durationNanos int64) {
	// FIXME - provide a scope?
	scopeMetrics := rm.GetOrCreateScope(pcommon.NewInstrumentationScope())
	metric := scopeMetrics.GetOrCreateMetric(metricName)
	metric.unit = "s"
	metric.AddHistogramDatapoint(attributes, startTimestamp, endTimestamp, NanosToSeconds(durationNanos))
}

func (rm *ResourceMetrics) AddHistogramFromSpan(metricName string, attributes pcommon.Map, span ptrace.Span) {
	rm.AddHistogram(metricName, attributes, span.StartTimestamp(), span.EndTimestamp(), (span.EndTimestamp() - span.StartTimestamp()).AsTime().UnixNano())
}

func (rm *ResourceMetrics) IncrementSum(metricName string, attributes pcommon.Map, startTimestamp pcommon.Timestamp, endTimestamp pcommon.Timestamp) {
	scopeMetrics := rm.GetOrCreateScope(pcommon.NewInstrumentationScope())
	metric := scopeMetrics.GetOrCreateMetric(metricName)
	sum := metric.GetSum(attributes, false, startTimestamp, endTimestamp)
	sum.Add(1, startTimestamp, endTimestamp)
}

func (rm *ResourceMetrics) GetSum(metricName string, attributes pcommon.Map, isMonotonic bool, startTimestamp pcommon.Timestamp, endTimestamp pcommon.Timestamp) SumDatapoint {
	scopeMetrics := rm.GetOrCreateScope(pcommon.NewInstrumentationScope())
	metric := scopeMetrics.GetOrCreateMetric(metricName)
	return metric.GetSum(attributes, isMonotonic, startTimestamp, endTimestamp)
}

func (m *Metric) GetSum(attributes pcommon.Map, isMonotonic bool, startTimestamp pcommon.Timestamp, endTimestamp pcommon.Timestamp) SumDatapoint {
	dp, dpPresent := m.sumDatapoints[getKeyFromMap(attributes)]
	if !dpPresent {
		dp = SumDatapoint{value: 0, attributes: attributes, isMonotonic: isMonotonic, startTimestamp: startTimestamp, timestamp: endTimestamp}
		m.sumDatapoints[getKeyFromMap(attributes)] = dp
	}
	m.sumDatapoints[getKeyFromMap(attributes)] = dp
	return dp
}

func (m *SumDatapoint) Add(value int64, startTimestamp pcommon.Timestamp, endTimestamp pcommon.Timestamp) {
	m.value += value
	if m.startTimestamp.AsTime().After(startTimestamp.AsTime()) {
		m.startTimestamp = startTimestamp
	}
	if m.timestamp.AsTime().Before(endTimestamp.AsTime()) {
		m.timestamp = endTimestamp
	}
}

type ScopeMetrics struct {
	origin  pcommon.InstrumentationScope
	metrics map[string]*Metric
}

func (sm *ScopeMetrics) GetOrCreateMetric(metricName string) *Metric {
	metric, metricPresent := sm.metrics[metricName]
	if metricPresent {
		return metric
	}
	metric = &Metric{
		metricName:          metricName,
		histogramDatapoints: make(map[string]HistogramDatapoint),
		sumDatapoints:       make(map[string]SumDatapoint),
	}
	sm.metrics[metricName] = metric
	return metric
}

type Metric struct {
	histogramDatapoints map[string]HistogramDatapoint
	sumDatapoints       map[string]SumDatapoint
	metricName          string
	unit                string
}

func (m *Metric) AddHistogramDatapoint(attributes pcommon.Map, startTimestamp pcommon.Timestamp, endTimestamp pcommon.Timestamp, value float64) {
	dp, dpPresent := m.histogramDatapoints[getKeyFromMap(attributes)]
	if !dpPresent {
		histogram := NewHistogram()
		dp = HistogramDatapoint{histogram: histogram, attributes: attributes, startTimestamp: startTimestamp, timestamp: endTimestamp}
	}
	dp.histogram.Update(value)
	if dp.startTimestamp.AsTime().After(startTimestamp.AsTime()) {
		dp.startTimestamp = startTimestamp
	}
	if dp.timestamp.AsTime().Before(endTimestamp.AsTime()) {
		dp.timestamp = endTimestamp
	}
	m.histogramDatapoints[getKeyFromMap(attributes)] = dp
}

func (m *Metric) IncrementSumDatapoint(attributes pcommon.Map, startTimestamp pcommon.Timestamp, endTimestamp pcommon.Timestamp) {
	dp, dpPresent := m.sumDatapoints[getKeyFromMap(attributes)]
	if !dpPresent {
		dp = SumDatapoint{value: 0, attributes: attributes, startTimestamp: startTimestamp, timestamp: endTimestamp}
	}
	dp.value++
	if dp.startTimestamp.AsTime().After(startTimestamp.AsTime()) {
		dp.startTimestamp = startTimestamp
	}
	if dp.timestamp.AsTime().Before(endTimestamp.AsTime()) {
		dp.timestamp = endTimestamp
	}
	m.sumDatapoints[getKeyFromMap(attributes)] = dp
}

func NanosToSeconds(nanos int64) float64 {
	return float64(nanos) / 1e9
}

type HistogramDatapoint struct {
	histogram      Histogram
	attributes     pcommon.Map
	startTimestamp pcommon.Timestamp
	timestamp      pcommon.Timestamp
}

type SumDatapoint struct {
	value          int64
	attributes     pcommon.Map
	startTimestamp pcommon.Timestamp
	timestamp      pcommon.Timestamp
	isMonotonic    bool
}

func getKeyFromMap(pMap pcommon.Map) string {
	m := make(map[string]string, pMap.Len())
	pMap.Range(func(k string, v pcommon.Value) bool {
		m[k] = v.AsString()
		return true
	})
	return getKey(m)
}

func getKey(m map[string]string) string {
	// map order is not guaranteed, we need to hash key values in order
	allKeys := make([]string, len(m))
	for k := range m {
		allKeys = append(allKeys, k)
	}
	sort.Strings(allKeys)
	toHash := make([]string, 2*len(m))
	for _, k := range allKeys {
		toHash = append(toHash, k)
		toHash = append(toHash, m[k])
	}
	return hash(toHash)
}

func hash(objs []string) string {
	digester := crypto.MD5.New()
	for _, ob := range objs {
		fmt.Fprint(digester, ob)
	}
	return string(digester.Sum(nil))
}

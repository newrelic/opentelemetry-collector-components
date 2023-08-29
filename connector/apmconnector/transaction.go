// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmconnector // import "github.com/newrelic/opentelemetry-collector-components/connector/apmconnector"

import (
	"fmt"
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type TransactionType string

const (
	DbOperationAttributeName = "db.operation"
	DbSystemAttributeName    = "db.system"
	DbSQLTableAttributeName  = "db.sql.table"
)

const (
	WebTransactionType   TransactionType = "Web"
	OtherTransactionType TransactionType = "Other"
	NullTransactionType  TransactionType = "Skip"
)

func (t TransactionType) AsString() string {
	return string(t)
}

func (t TransactionType) GetOverviewMetricName() string {
	switch t {
	case WebTransactionType:
		return "apm.service.overview.web"
	default:
		return "apm.service.overview.other"
	}
}

type Apdex struct {
	apdexSatisfying float64
	apdexTolerating float64
}

func NewApdex(apdexT float64) Apdex {
	return Apdex{apdexSatisfying: apdexT, apdexTolerating: apdexT * 4}
}

func (apdex Apdex) GetApdexBucket(durationInSeconds float64) string {
	if durationInSeconds <= apdex.apdexSatisfying {
		return "S"
	} else if durationInSeconds <= apdex.apdexTolerating {
		return "T"
	} else {
		return "F"
	}
}

type Transaction struct {
	SdkLanguage         string
	SpanToChildDuration map[string]int64
	resourceMetrics     *ResourceMetrics
	Measurements        map[string]*Measurement
	sqlParser           *SQLParser
	apdex               Apdex
	RootSpan            ptrace.Span
}

type Measurement struct {
	SpanID, MetricName, MetricTimesliceName string
	DurationNanos, ExclusiveDurationNanos   int64
	Attributes                              pcommon.Map
	SegmentNameProvider                     func(TransactionType) string
	// FIXME
	Span ptrace.Span
}

type TransactionsMap struct {
	sqlParser    *SQLParser
	apdex        Apdex
	Transactions map[string]*Transaction
}

func NewTransactionsMap(apdexT float64) *TransactionsMap {
	return &TransactionsMap{Transactions: make(map[string]*Transaction), sqlParser: NewSQLParser(), apdex: NewApdex(apdexT)}
}

func (transactions *TransactionsMap) ProcessTransactions() {
	for _, transaction := range transactions.Transactions {
		// if this returns false, we MAY not have seen all of the spans for a trace
		transaction.ProcessRootSpan()
	}
}

func GetTransactionKey(traceID string, resourceAttributes pcommon.Map) string {
	keys := []string{"host.name", "service.namespace", "service.name", "telemetry.sdk.language", "container.id", "service.instance.id"}
	values := []string{}
	for _, key := range keys {
		if value, exists := resourceAttributes.Get(key); exists {
			values = append(values, value.AsString())
		} else {
			values = append(values, "")
		}
	}
	values = append(values, traceID)
	return strings.Join(values[:], ":")
}

func (transactions *TransactionsMap) GetOrCreateTransaction(sdkLanguage string, span ptrace.Span, resourceMetrics *ResourceMetrics, resourceAttributes pcommon.Map) (*Transaction, string) {
	traceID := span.TraceID().String()
	key := GetTransactionKey(traceID, resourceAttributes)
	transaction, txExists := transactions.Transactions[key]
	if !txExists {
		transaction = &Transaction{SdkLanguage: sdkLanguage, SpanToChildDuration: make(map[string]int64),
			resourceMetrics: resourceMetrics, Measurements: make(map[string]*Measurement), sqlParser: transactions.sqlParser, apdex: transactions.apdex}
		transactions.Transactions[key] = transaction
		//fmt.Printf("Created transaction for: %s   %s\n", traceID, transaction.sdkLanguage)
	}

	return transaction, traceID
}

func (transaction *Transaction) IsRootSet() bool {
	return (ptrace.Span{}) != transaction.RootSpan
}

func (transaction *Transaction) SetRootSpan(span ptrace.Span) bool {
	// favor server span
	if transaction.IsRootSet() && (transaction.RootSpan.Kind() == ptrace.SpanKindServer ||
		transaction.RootSpan.Kind() == ptrace.SpanKindConsumer ||
		transaction.RootSpan.Kind() == ptrace.SpanKindProducer) {
		return false
	}
	transaction.RootSpan = span
	return true
}

func (transaction *Transaction) AddSpan(span ptrace.Span) {
	if span.Kind() == ptrace.SpanKindServer {
		transaction.SetRootSpan(span)
		return
	}
	isRoot := span.ParentSpanID().IsEmpty() && transaction.SetRootSpan(span)
	if !isRoot {
		parentSpanID := span.ParentSpanID().String()
		newDuration := DurationInNanos(span)

		if measurement, exists := transaction.Measurements[parentSpanID]; exists {
			measurement.ExclusiveDurationNanos -= newDuration
		} else {
			transaction.SpanToChildDuration[parentSpanID] += newDuration
		}
	}

	if span.Kind() == ptrace.SpanKindClient {
		// filter out db calls that have no parent (so no transaction)
		if !isRoot {
			if transaction.ProcessClientSpan(span) {
				return
			}
		}
	}
	transaction.ProcessGenericSpan(span)
}

func NewSimpleNameProvider(name string) func(TransactionType) string {
	return func(t TransactionType) string { return name }
}

func (transaction *Transaction) AddMeasurement(measurement *Measurement) {
	transaction.Measurements[measurement.SpanID] = measurement
	measurement.ExclusiveDurationNanos = measurement.ExclusiveTime(transaction)
	if measurement.ExclusiveDurationNanos < 0 {
		// FIXME log this
		measurement.ExclusiveDurationNanos = 0
	}
	measurement.Attributes.PutStr("metricTimesliceName", measurement.MetricTimesliceName)
}

func (transaction *Transaction) ProcessDatabaseSpan(span ptrace.Span) bool {
	dbSystem, dbSystemPresent := span.Attributes().Get(DbSystemAttributeName)
	if !dbSystemPresent {
		return false
	}
	dbOperation, dbOperationPresent := span.Attributes().Get(DbOperationAttributeName)
	if !dbOperationPresent {
		return false
	}
	dbTable, dbTablePresent := span.Attributes().Get(DbSQLTableAttributeName)
	if !dbTablePresent {
		dbTable = pcommon.NewValueStr("unknown")
	}
	attributes := pcommon.NewMap()
	attributes.EnsureCapacity(10)
	attributes.PutStr(DbOperationAttributeName, dbOperation.AsString())
	attributes.PutStr(DbSystemAttributeName, dbSystem.AsString())
	attributes.PutStr(DbSQLTableAttributeName, dbTable.AsString())
	for _, key := range []string{"net.peer.name", "db.name"} {
		if value, exists := span.Attributes().Get(key); exists {
			attributes.PutStr(key, value.AsString())
		}
	}

	timesliceName := fmt.Sprintf("Datastore/statement/%s/%s/%s", dbSystem.AsString(), dbTable.AsString(), dbOperation.AsString())
	measurement := Measurement{SpanID: span.SpanID().String(), MetricName: "apm.service.datastore.operation.duration", Span: span,
		DurationNanos: DurationInNanos(span), Attributes: attributes, SegmentNameProvider: NewSimpleNameProvider(dbSystem.AsString()), MetricTimesliceName: timesliceName}

	transaction.AddMeasurement(&measurement)

	return true
}

func (transaction *Transaction) ProcessExternalSpan(span ptrace.Span) bool {
	serverAddress, serverAddressKey := GetFirst(span.Attributes(), []string{"server.address", "net.peer.name"})
	if serverAddressKey != "" {
		attributes := pcommon.NewMap()
		attributes.PutStr("server.address", serverAddress.AsString())
		// FIXME remove after UI is updated
		attributes.PutStr("external.host", serverAddress.AsString())

		segmentNameProvider := func(t TransactionType) string {
			switch t {
			case WebTransactionType:
				return "Web external"
			default:
				return "Background external"
			}
		}
		timesliceName := fmt.Sprintf("External/%s/all", serverAddress.AsString())
		measurement := Measurement{SpanID: span.SpanID().String(), MetricName: "apm.service.external.host.duration", Span: span,
			DurationNanos: DurationInNanos(span), Attributes: attributes, SegmentNameProvider: segmentNameProvider, MetricTimesliceName: timesliceName}

		transaction.AddMeasurement(&measurement)
		return true
	}
	return false
}

func (transaction *Transaction) ProcessGenericSpan(span ptrace.Span) bool {
	attributes := pcommon.NewMap()
	timesliceName := fmt.Sprintf("Custom/%s", span.Name())
	measurement := Measurement{SpanID: span.SpanID().String(), MetricName: "newrelic.timeslice.value", Span: span,
		DurationNanos: DurationInNanos(span), Attributes: attributes, SegmentNameProvider: NewSimpleNameProvider(transaction.SdkLanguage), MetricTimesliceName: timesliceName}

	transaction.AddMeasurement(&measurement)

	return true
}

func (transaction *Transaction) ProcessClientSpan(span ptrace.Span) bool {
	return transaction.ProcessDatabaseSpan(span) || transaction.ProcessExternalSpan(span)
}

func (transaction *Transaction) ProcessRootSpan() bool {
	if !transaction.IsRootSet() {
		return false
	}
	span := transaction.RootSpan

	transactionName, transactionType := GetTransactionMetricName(span)
	if transactionType == NullTransactionType {
		return true
	}

	err := span.Status().Code() == ptrace.StatusCodeError
	if err {
		transaction.IncrementErrorCount(transactionName, transactionType, span.EndTimestamp())
	}

	transaction.GenerateApdexMetrics(span, err, transactionName, transactionType)

	breakdownBySegment := make(map[string]int64)
	totalBreakdownNanos := int64(0)
	for _, measurement := range transaction.Measurements {
		transaction.ProcessMeasurement(measurement, transactionType, transactionName)
		segmentName := measurement.SegmentNameProvider(transactionType)
		breakdownBySegment[segmentName] += measurement.ExclusiveDurationNanos
		totalBreakdownNanos += measurement.ExclusiveDurationNanos
	}

	remainingNanos := DurationInNanos(span) - totalBreakdownNanos
	if remainingNanos > 0 {
		breakdownBySegment[transaction.SdkLanguage] += remainingNanos
	}

	overviewMetricName := transactionType.GetOverviewMetricName()

	for segment, sum := range breakdownBySegment {
		attributes := pcommon.NewMap()
		attributes.PutStr("segmentName", segment)

		transaction.resourceMetrics.RecordHistogram(overviewMetricName, attributes,
			span.StartTimestamp(), span.EndTimestamp(), sum)
	}

	{
		attributes := pcommon.NewMap()
		attributes.PutStr("transactionType", transactionType.AsString())
		attributes.PutStr("transactionName", transactionName)
		attributes.PutStr("metricTimesliceName", transactionName)

		transaction.resourceMetrics.RecordHistogramFromSpan("apm.service.transaction.duration", attributes, span)

		if remainingNanos > 0 {
			attributes.PutStr("transactionName", transactionName)

			// blame any time not attributed to measurements to the transaction itself
			transaction.resourceMetrics.RecordHistogram("apm.service.transaction.overview", attributes,
				span.StartTimestamp(), span.EndTimestamp(), remainingNanos)
		}
	}

	return true
}

func (transaction *Transaction) GenerateApdexMetrics(span ptrace.Span, err bool, transactionName string, transactionType TransactionType) {
	attributes := pcommon.NewMap()
	attributes.PutDouble("apdex.value", transaction.apdex.apdexSatisfying)
	attributes.PutStr("transactionType", transactionType.AsString())
	if err {
		attributes.PutStr("apdex.bucket", "F")
	} else {
		durationSeconds := NanosToSeconds(DurationInNanos(span))
		attributes.PutStr("apdex.bucket", transaction.apdex.GetApdexBucket(durationSeconds))
	}
	transaction.resourceMetrics.IncrementSum("apm.service.apdex", attributes, span.EndTimestamp())

	txAttributes := pcommon.NewMap()
	attributes.CopyTo(txAttributes)
	txAttributes.PutStr("transactionName", transactionName)
	transaction.resourceMetrics.IncrementSum("apm.service.transaction.apdex", txAttributes, span.EndTimestamp())
}

func (transaction *Transaction) IncrementErrorCount(transactionName string, transactionType TransactionType, timestamp pcommon.Timestamp) {
	{
		attributes := pcommon.NewMap()
		attributes.PutStr("transactionType", transactionType.AsString())
		transaction.resourceMetrics.IncrementSum("apm.service.error.count", attributes, timestamp)
	}
	{
		attributes := pcommon.NewMap()
		attributes.PutStr("transactionName", transactionName)
		attributes.PutStr("transactionType", transactionType.AsString())
		transaction.resourceMetrics.IncrementSum("apm.service.transaction.error.count", attributes, timestamp)
	}
}

func (transaction *Transaction) ProcessMeasurement(measurement *Measurement, transactionType TransactionType, transactionName string) {
	measurement.Attributes.PutStr("transactionType", transactionType.AsString())
	measurement.Attributes.PutStr("scope", transactionName)

	transaction.resourceMetrics.RecordHistogramFromSpan(measurement.MetricName, measurement.Attributes, measurement.Span)

	{
		attributes := pcommon.NewMap()
		measurement.Attributes.CopyTo(attributes)
		// we might not need transactionName here..
		attributes.PutStr("transactionName", transactionName)

		transaction.resourceMetrics.RecordHistogram("apm.service.transaction.overview", attributes,
			measurement.Span.StartTimestamp(), measurement.Span.EndTimestamp(), measurement.ExclusiveDurationNanos)
	}
}

func DurationInNanos(span ptrace.Span) int64 {
	return (span.EndTimestamp() - span.StartTimestamp()).AsTime().UnixNano()
}

func (measurement Measurement) ExclusiveTime(transaction *Transaction) int64 {
	childDurationNanos := transaction.SpanToChildDuration[measurement.SpanID]
	// we no longer need the summed child durations, delete that
	delete(transaction.SpanToChildDuration, measurement.SpanID)
	return measurement.DurationNanos - childDurationNanos
}

func GetTransactionMetricName(span ptrace.Span) (string, TransactionType) {
	if span.Kind() != ptrace.SpanKindServer {
		return "", NullTransactionType
	}
	if rpcService, rpcServicePresent := span.Attributes().Get("rpc.service"); rpcServicePresent {
		if rpcMethod, rpcMethodPresent := span.Attributes().Get("rpc.method"); rpcMethodPresent {
			return fmt.Sprintf("WebTransaction/rpc/%s/%s", rpcService.AsString(), rpcMethod.AsString()), WebTransactionType
		} else {
			return fmt.Sprintf("WebTransaction/rpc/%s", rpcService.AsString()), WebTransactionType
		}
	}
	if httpRoute, routePresent := span.Attributes().Get("http.route"); routePresent {
		return GetWebTransactionMetricName(span, httpRoute.Str(), "http.route")
	}
	if urlPath, _ := GetFirst(span.Attributes(), []string{"url.path", "http.target"}); urlPath.Type() != pcommon.ValueTypeEmpty {
		return GetWebTransactionMetricName(span, urlPath.Str(), "Uri")
	}
	if method, methodPresent := span.Attributes().Get("http.method"); methodPresent {
		return fmt.Sprintf("WebTransaction/http.method/%s", method.Str()), WebTransactionType
	}
	return "WebTransaction/Other/unknown", WebTransactionType
}

func GetWebTransactionMetricName(span ptrace.Span, name, nameType string) (string, TransactionType) {
	if method, methodPresent := span.Attributes().Get("http.method"); methodPresent {
		return fmt.Sprintf("WebTransaction/%s%s (%s)", nameType, name, method.Str()), WebTransactionType
	}
	return fmt.Sprintf("WebTransaction/%s%s", nameType, name), WebTransactionType
}

func GetFirst(attributes pcommon.Map, keys []string) (pcommon.Value, string) {
	for _, key := range keys {
		if value, exists := attributes.Get(key); exists {
			return value, key
		}
	}
	return pcommon.NewValueEmpty(), ""
}

func GetSdkLanguage(attributes pcommon.Map) string {
	sdkLanguage, sdkLanguagePresent := attributes.Get("telemetry.sdk.language")
	if sdkLanguagePresent {
		return sdkLanguage.AsString()
	}
	return "unknown"
}

// Generate the metrc used for the host instances drop down
func GenerateInstanceMetric(resourceMetrics *ResourceMetrics, hostName string, timestamp pcommon.Timestamp) {
	attributes := pcommon.NewMap()
	attributes.PutStr("instanceName", hostName)
	attributes.PutStr("host.displayName", hostName)
	resourceMetrics.IncrementSum("apm.service.instance.count", pcommon.NewMap(), timestamp)
}

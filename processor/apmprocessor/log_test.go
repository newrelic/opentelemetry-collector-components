// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmprocessor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processortest"
)

func TestProcessLog(t *testing.T) {
	sink := new(consumertest.LogsSink)
	lp, err := createTestLogsProcessor(sink)
	require.Nil(t, err)
	require.NotNil(t, lp)

	logs := plog.NewLogs()
	resourceLogs := logs.ResourceLogs().AppendEmpty()
	resourceLogs.Resource().Attributes().PutStr("service.name", "service")
	resourceLogs.Resource().Attributes().PutStr("telemetry.sdk.name", "opentelemetry")
	logRecords := resourceLogs.ScopeLogs().AppendEmpty().LogRecords()
	lr := logRecords.AppendEmpty()
	lr.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))

	err = lp.ConsumeLogs(context.Background(), logs)
	require.Nil(t, err)
	provider, providerPresent := sink.AllLogs()[0].ResourceLogs().At(0).Resource().Attributes().Get("instrumentation.provider")
	assert.True(t, providerPresent)
	assert.Equal(t, "newrelic-opentelemetry", provider.AsString())
}

func createTestLogsProcessor(sink consumer.Logs) (processor.Logs, error) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	oCfg := cfg.(*Config)
	lp, err := factory.CreateLogsProcessor(context.Background(), processortest.NewNopCreateSettings(), oCfg, sink)

	return lp, err
}

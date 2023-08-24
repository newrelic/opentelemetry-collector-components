// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmprocessor // import "github.com/newrelic/opentelemetry-collector-components/processor/apmprocessor"

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

var processorCapabilities = consumer.Capabilities{MutatesData: true}

// FIXME copying this from the metadata/generated_status to be able to build the component externally
const (
	Type             = "newrelicapm"
	TracesStability  = component.StabilityLevelDevelopment
	MetricsStability = component.StabilityLevelDevelopment
	LogsStability    = component.StabilityLevelDevelopment
)

// NewFactory returns a new factory for the Attributes processor.
func NewFactory() processor.Factory {
	return processor.NewFactory(
		Type,
		createDefaultConfig,
		processor.WithTraces(createTracesProcessor, TracesStability),
		processor.WithMetrics(createMetricsProcessor, MetricsStability),
		processor.WithLogs(createLogsProcessor, LogsStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{ChangeInstrumentationProvider: true}
}

func createTracesProcessor(
	ctx context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Traces,
) (processor.Traces, error) {
	oCfg := cfg.(*Config)
	sp, err := newSpanProcessor(*oCfg, set.Logger)
	if err != nil {
		return nil, err
	}
	return processorhelper.NewTracesProcessor(
		ctx,
		set,
		cfg,
		nextConsumer,
		sp.processTraces,
		processorhelper.WithCapabilities(processorCapabilities))
}

func createMetricsProcessor(
	ctx context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (processor.Metrics, error) {
	oCfg := cfg.(*Config)

	metricsProcessor := newMetricProcessor(*oCfg, set.Logger)

	return processorhelper.NewMetricsProcessor(
		ctx,
		set,
		cfg,
		nextConsumer,
		metricsProcessor.processMetrics,
		processorhelper.WithCapabilities(processorCapabilities))
}

func createLogsProcessor(
	ctx context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Logs,
) (processor.Logs, error) {
	oCfg := cfg.(*Config)

	logProcessor := newLogProcessor(*oCfg, set.Logger)

	return processorhelper.NewLogsProcessor(
		ctx,
		set,
		cfg,
		nextConsumer,
		logProcessor.processLogs,
		processorhelper.WithCapabilities(processorCapabilities))
}

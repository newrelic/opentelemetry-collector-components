// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmconnector // import "apmconnector"

//go:generate mdatagen metadata.yaml

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer"
)

const (
	typeStr   = "apm"
	stability = component.StabilityLevelBeta
)

// NewFactory returns a ConnectorFactory.
func NewFactory() connector.Factory {
	return connector.NewFactory(
		typeStr,
		createDefaultConfig,
		connector.WithTracesToMetrics(createTracesToMetrics, stability),
		connector.WithTracesToLogs(createTracesToLogs, stability),
		connector.WithTracesToTraces(createTracesToTraces, stability),
	)
}

// createDefaultConfig creates the default configuration.
func createDefaultConfig() component.Config {
	return &Config{}
}

// createTracesToMetrics creates a traces to metrics connector based on provided config.
func createTracesToMetrics(
	_ context.Context,
	set connector.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (connector.Traces, error) {
	c := cfg.(*Config)

	return &ApmMetricConnector{
		config:          c,
		metricsConsumer: nextConsumer,
		logger:          set.Logger,
	}, nil
}

// createTracesToLogs creates a traces to logs connector based on provided config.
func createTracesToLogs(
	_ context.Context,
	set connector.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Logs,
) (connector.Traces, error) {
	c := cfg.(*Config)

	return &ApmLogConnector{
		config:       c,
		logsConsumer: nextConsumer,
		logger:       set.Logger,
	}, nil
}

// createTracesToTraces creates a traces to traces connector based on provided config.
func createTracesToTraces(
	_ context.Context,
	set connector.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Traces,
) (connector.Traces, error) {
	c := cfg.(*Config)

	return &ApmTraceConnector{
		config:         c,
		tracesConsumer: nextConsumer,
		sqlparser:      NewSQLParser(),
		logger:         set.Logger,
	}, nil
}

// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmconnector // import "github.com/newrelic/opentelemetry-collector-components/connector/apmconnector"

//go:generate mdatagen metadata.yaml

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer"
)

// FIXME copying this from the metadata/generated_status to be able to build the component externally
const (
	Type                      = "newrelicapm"
	TracesToMetricsStability  = component.StabilityLevelDevelopment
	MetricsToMetricsStability = component.StabilityLevelDevelopment
)

// NewFactory returns a ConnectorFactory.
func NewFactory() connector.Factory {
	return connector.NewFactory(
		Type,
		createDefaultConfig,
		connector.WithTracesToMetrics(createTracesToMetrics, TracesToMetricsStability),
		connector.WithMetricsToMetrics(createMetricsToMetrics, MetricsToMetricsStability),
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

// createMetricsToMetrics creates a metrics to metrics connector based on provided config.
func createMetricsToMetrics(
	_ context.Context,
	set connector.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (connector.Metrics, error) {
	c := cfg.(*Config)

	return &OpenTelemetryMetricToApmMetricConnector{
		config:          c,
		metricsConsumer: nextConsumer,
		logger:          set.Logger,
	}, nil
}

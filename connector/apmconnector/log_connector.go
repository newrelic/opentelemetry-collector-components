// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmconnector // import "github.com/newrelic/opentelemetry-collector-components/connector/apmconnector"

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

type ApmLogConnector struct {
	config *Config
	logger *zap.Logger

	logsConsumer consumer.Logs
}

func (c *ApmLogConnector) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (c *ApmLogConnector) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	logs := BuildTransactions(td)
	return c.logsConsumer.ConsumeLogs(ctx, logs)
}

func (c *ApmLogConnector) Start(_ context.Context, _ component.Host) error {
	c.logger.Info("Starting the APM Log Connector")
	if c.config.ApdexT == 0 {
		c.config.ApdexT = 0.5
	}
	return nil
}

func (c *ApmLogConnector) Shutdown(context.Context) error {
	c.logger.Info("Stopping the APM Log Connector")
	return nil
}

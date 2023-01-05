// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package nopreceiver // import "github.com/cristianciutea/opentelemetry-components/receiver/nopreceiver"

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

const (
	typeStr   = "nop_stats"
	stability = component.StabilityLevelDevelopment
)

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, stability))
}

func createDefaultConfig() component.Config {
	return &scraperhelper.ScraperControllerSettings{
		CollectionInterval: 10 * time.Second,
	}
}

func createMetricsReceiver(
	ctx context.Context,
	params receiver.CreateSettings,
	config component.Config,
	consumer consumer.Metrics,
) (receiver.Metrics, error) {
	scConf := config.(*scraperhelper.ScraperControllerSettings)
	dsr, err := newReceiver(*scConf, params, consumer)
	if err != nil {
		return nil, err
	}

	return dsr, nil
}

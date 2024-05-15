// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package nopreceiver // import "github.com/newrelic/opentelemetry-collector-components/receiver/nopreceiver"

import (
	"context"
	"time"

	"github.com/newrelic/opentelemetry-collector-components/receiver/nopreceiver/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

const (
	stability = component.StabilityLevelDevelopment
)

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		metadata.Type,
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, stability))
}

func createDefaultConfig() component.Config {
	return &scraperhelper.ControllerConfig{
		CollectionInterval: 10 * time.Second,
	}
}

func createMetricsReceiver(
	_ context.Context,
	params receiver.CreateSettings,
	config component.Config,
	consumer consumer.Metrics,
) (receiver.Metrics, error) {
	scConf := config.(*scraperhelper.ControllerConfig)
	dsr, err := newReceiver(*scConf, params, consumer)
	if err != nil {
		return nil, err
	}

	return dsr, nil
}

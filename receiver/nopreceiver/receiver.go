// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package nopreceiver

import (
	"context"

	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

func newReceiver(config scraperhelper.ScraperControllerSettings, set receiver.CreateSettings, nextConsumer consumer.Metrics) (receiver.Metrics, error) {
	scrp, err := scraperhelper.NewScraper("nopworld", scrape)
	if err != nil {
		return nil, err
	}
	return scraperhelper.NewScraperControllerReceiver(&config, set, nextConsumer, scraperhelper.AddScraper(scrp))
}

func greeterMetrics(name string) pmetric.Metrics {
	md := pmetric.NewMetrics()

	rs := md.ResourceMetrics().AppendEmpty()

	resourceAttr := rs.Resource().Attributes()
	resourceAttr.PutStr("greeter.name", name)

	ms := rs.ScopeMetrics().AppendEmpty().Metrics()
	m := ms.AppendEmpty()
	m.SetName("nop.requests")
	m.SetUnit("requests")
	m.SetEmptyGauge().DataPoints().AppendEmpty().SetIntValue(1)

	return md
}

func scrape(ctx context.Context) (pmetric.Metrics, error) {
	md := pmetric.NewMetrics()

	greeterMetrics("bob").ResourceMetrics().CopyTo(md.ResourceMetrics())

	return md, nil
}

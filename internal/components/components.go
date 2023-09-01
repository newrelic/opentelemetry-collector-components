// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package components // import "github.com/newrelic/opentelemetry-collector-components/internal/components"

import (
	"github.com/newrelic/opentelemetry-collector-components/connector/apmconnector"
	"github.com/newrelic/opentelemetry-collector-components/processor/apmprocessor"
	"github.com/newrelic/opentelemetry-collector-components/receiver/nopreceiver"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusremotewriteexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver"
)

func Components() (otelcol.Factories, error) {
	var err error
	factories := otelcol.Factories{}
	extensions := []extension.Factory{}
	factories.Extensions, err = extension.MakeFactoryMap(extensions...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	receivers := []receiver.Factory{
		hostmetricsreceiver.NewFactory(),
		otlpreceiver.NewFactory(),
		prometheusreceiver.NewFactory(),
		nopreceiver.NewFactory(),
	}
	factories.Receivers, err = receiver.MakeFactoryMap(receivers...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	exporters := []exporter.Factory{
		otlpexporter.NewFactory(),
		prometheusexporter.NewFactory(),
		prometheusremotewriteexporter.NewFactory(),
	}
	factories.Exporters, err = exporter.MakeFactoryMap(exporters...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	processors := []processor.Factory{
		batchprocessor.NewFactory(),
		apmprocessor.NewFactory(),
	}
	factories.Processors, err = processor.MakeFactoryMap(processors...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	connectors := []connector.Factory{
		apmconnector.NewFactory(),
	}
	factories.Connectors, err = connector.MakeFactoryMap(connectors...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	return factories, nil
}

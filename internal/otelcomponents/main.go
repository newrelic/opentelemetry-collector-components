// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package otelcomponents // import "github.com/newrelic/opentelemetry-collector-components/internal/otelcomponents"

import (
	"fmt"
	"log"

	"github.com/newrelic/opentelemetry-collector-components/internal/otelcomponents/internal/version"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/otelcol"
)

type ComponentsFunc func() (otelcol.Factories, error)

func RunWithComponents(componentsFunc ComponentsFunc) {
	factories, err := componentsFunc()
	if err != nil {
		log.Fatalf("failed to build components: %v", err)
	}

	info := component.BuildInfo{
		Command:     "otelcomponents",
		Description: "New Relic OpenTelemetry Collector Components",
		Version:     version.Version,
	}

	if err = run(otelcol.CollectorSettings{BuildInfo: info, Factories: factories}); err != nil {
		log.Fatal(err)
	}
}

func runInteractive(params otelcol.CollectorSettings) error {
	cmd := otelcol.NewCommand(params)
	if err := cmd.Execute(); err != nil {
		return fmt.Errorf("collector server run finished with error: %w", err)
	}

	return nil
}

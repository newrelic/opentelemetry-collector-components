// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/newrelic/opentelemetry-collector-components/internal/components"
	"github.com/newrelic/opentelemetry-collector-components/internal/otelcomponents"
)

func main() {
	otelcomponents.RunWithComponents(components.Components)
}

// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/cristianciutea/opentelemetry-components/internal/components"
	"github.com/cristianciutea/opentelemetry-components/internal/otelcomponents"
)

func main() {
	otelcomponents.RunWithComponents(components.Components)
}

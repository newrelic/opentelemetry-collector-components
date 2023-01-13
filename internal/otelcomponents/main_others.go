// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build !windows
// +build !windows

package otelcomponents // import "github.com/cristianciutea/opentelemetry-components/internal/otelcomponents"

import "go.opentelemetry.io/collector/otelcol"

func run(params otelcol.CollectorSettings) error {
	return runInteractive(params)
}

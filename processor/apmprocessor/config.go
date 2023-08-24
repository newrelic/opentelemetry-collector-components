// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmprocessor // import "github.com/newrelic/opentelemetry-collector-components/processor/apmprocessor"

import (
	"go.opentelemetry.io/collector/component"
)

type Config struct {
	// If set to true, will set the `instrumentation.provider` attribute to `newrelic-opentelemetry`
	ChangeInstrumentationProvider bool `mapstructure:"change_instrumentation_provider"`
}

var _ component.Config = (*Config)(nil)

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	return nil
}

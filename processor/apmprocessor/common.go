// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmprocessor // import "github.com/newrelic/opentelemetry-collector-components/processor/apmprocessor"

import "go.opentelemetry.io/collector/pdata/pcommon"

func setInstrumentationProvider(config Config, resource pcommon.Resource) {
	if !config.ChangeInstrumentationProvider {
		return
	}
	instrumentationProvider, instrumentationProviderPresent := resource.Attributes().Get("instrumentation.provider")
	if instrumentationProviderPresent {
		if instrumentationProvider.AsString() == "opentelemetry" {
			resource.Attributes().PutStr("instrumentation.provider", "newrelic-opentelemetry")
		}
		return
	}

	sdk, sdkPresent := resource.Attributes().Get("telemetry.sdk.name")
	if sdkPresent && sdk.AsString() == "opentelemetry" {
		resource.Attributes().PutStr("instrumentation.provider", "newrelic-opentelemetry")
	}
}

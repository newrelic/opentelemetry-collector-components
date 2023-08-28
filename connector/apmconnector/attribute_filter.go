// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmconnector // import "github.com/newrelic/opentelemetry-collector-components/connector/apmconnector"

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
)

type AttributeFilter struct {
	attributesToKeep []string
}

func NewAttributeFilter() *AttributeFilter {
	return &AttributeFilter{attributesToKeep: []string{"instrumentation.provider", "os.description", "telemetry.auto.version", "telemetry.sdk.language", "host.name",
		"os.type", "telemetry.sdk.name", "process.runtime.description", "process.runtime.version", "telemetry.sdk.version",
		"host.arch", "service.name", "service.instance.id"}}
}

func (attributeFilter *AttributeFilter) FilterAttributes(from pcommon.Map) (pcommon.Map, error) {
	f := from.AsRaw()
	m := make(map[string]any)
	for _, k := range attributeFilter.attributesToKeep {
		if v, exists := f[k]; exists {
			m[k] = v
		}
	}
	newMap := pcommon.NewMap()
	err := newMap.FromRaw(m)
	if err != nil {
		return newMap, nil
	}
	if hostName, exists := from.Get("host.name"); exists {
		newMap.PutStr("host", hostName.AsString())

		if _, e := newMap.Get("service.instance.id"); !e {
			newMap.PutStr("service.instance.id", hostName.AsString())
		}
	}
	return newMap, nil
}

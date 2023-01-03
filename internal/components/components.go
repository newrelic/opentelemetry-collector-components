// Copyright 2019 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package components // import "github.com/cristianciutea/opentelemetry-components/internal/components"

import (
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/nginxreceiver"
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
		nginxreceiver.NewFactory(),
	}
	factories.Receivers, err = receiver.MakeFactoryMap(receivers...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	exporters := []exporter.Factory{}
	factories.Exporters, err = exporter.MakeFactoryMap(exporters...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	processors := []processor.Factory{}
	factories.Processors, err = processor.MakeFactoryMap(processors...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	return factories, nil
}

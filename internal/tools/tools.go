// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build tools
// +build tools

package tools // import "github.com/cristianciutea/opentelemetry-components/internal/tools"
import (
	_ "github.com/client9/misspell/cmd/misspell"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/google/addlicense"
	_ "github.com/jcchavezs/porto/cmd/porto"
	_ "github.com/open-telemetry/opentelemetry-collector-contrib/cmd/mdatagen"
	_ "github.com/pavius/impi/cmd/impi"
	_ "go.opentelemetry.io/build-tools/checkdoc"
	_ "go.opentelemetry.io/build-tools/crosslink"
	_ "golang.org/x/tools/cmd/goimports"
)

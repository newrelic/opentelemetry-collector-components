// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build windows
// +build windows

package otelcomponents // import "github.com/cristianciutea/opentelemetry-components/internal/otelcomponents"

import (
	"fmt"
	"os"

	"go.opentelemetry.io/collector/otelcol"
	"golang.org/x/sys/windows/svc"
)

func run(params otelcol.CollectorSettings) error {
	if useInteractiveMode, err := checkUseInteractiveMode(); err != nil {
		return err
	} else if useInteractiveMode {
		return runInteractive(params)
	} else {
		return runService(params)
	}
}

func checkUseInteractiveMode() (bool, error) {
	// If environment variable NO_WINDOWS_SERVICE is set with any value other
	// than 0, use interactive mode instead of running as a service. This should
	// be set in case running as a service is not possible or desired even
	// though the current session is not detected to be interactive
	if value, present := os.LookupEnv("NO_WINDOWS_SERVICE"); present && value != "0" {
		return true, nil
	}

	if isInteractiveSession, err := svc.IsAnInteractiveSession(); err != nil {
		return false, fmt.Errorf("failed to determine if we are running in an interactive session %w", err)
	} else {
		return isInteractiveSession, nil
	}
}

func runService(params otelcol.CollectorSettings) error {
	// do not need to supply service name when startup is invoked through Service Control Manager directly
	if err := svc.Run("", otelcol.NewSvcHandler(params)); err != nil {
		return fmt.Errorf("failed to start collector server: %w", err)
	}

	return nil
}

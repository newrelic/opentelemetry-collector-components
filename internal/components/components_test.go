// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build !windows
// +build !windows

package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultComponents(t *testing.T) {
	factories, err := Components()
	assert.NoError(t, err)

	exts := factories.Extensions
	for k, v := range exts {
		assert.Equal(t, k, v.Type())
	}

	recvs := factories.Receivers
	for k, v := range recvs {
		assert.Equal(t, k, v.Type())
	}

	procs := factories.Processors
	for k, v := range procs {
		assert.Equal(t, k, v.Type())
	}

	exps := factories.Exporters
	for k, v := range exps {
		assert.Equal(t, k, v.Type())
	}
}

// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build darwin

package keybase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPausedPrompt(t *testing.T) {
	ctx := newContext(&config{}, log)
	cancel := ctx.PausedPrompt()
	assert.False(t, cancel)
}

func TestRestart(t *testing.T) {
	ctx := newContext(&config{}, log)
	err := ctx.Restart()
	assert.EqualError(t, err, "Unsupported")
}

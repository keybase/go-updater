// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build darwin

package keybase

import (
	"testing"

	"github.com/keybase/go-updater"
	"github.com/stretchr/testify/assert"
)

func TestUpdatePrompt(t *testing.T) {
	ctx := newContext(&testConfigPlatform{}, log)
	_, err := ctx.UpdatePrompt(testUpdate, testOptions, updater.UpdatePromptOptions{})
	assert.Error(t, err, "Unsupported")
}

func TestPausedPrompt(t *testing.T) {
	ctx := newContext(&testConfigPlatform{}, log)
	cancel := ctx.PausedPrompt()
	assert.False(t, cancel)
}

func TestRestart(t *testing.T) {
	ctx := newContext(&config{}, log)
	err := ctx.Restart()
	assert.EqualError(t, err, "Unsupported")
}

// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build windows

package keybase

import (
	"testing"

	"github.com/keybase/go-updater"
	"github.com/stretchr/testify/assert"
)

func TestUpdatePrompt(t *testing.T) {
	ctx := newContext(&testConfigPlatform{}, testLog)
	resp, err := ctx.UpdatePrompt(testUpdate, testOptions, updater.UpdatePromptOptions{})
	assert.NoError(t, err)
	// No response for Windows
	assert.Nil(t, resp)
}

// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build linux

package keybase

import (
	"testing"

	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdatePrompt(t *testing.T) {
	ctx := newContext(&testConfigPlatform{}, testLog)
	resp, err := ctx.UpdatePrompt(testUpdate, testOptions, updater.UpdatePromptOptions{})
	assert.Equal(t, &updater.UpdatePromptResponse{Action: updater.UpdateActionContinue}, resp)
	require.NoError(t, err)
}

func TestPausedPrompt(t *testing.T) {
	ctx := newContext(&testConfigPlatform{}, testLog)
	cancel := ctx.PausedPrompt()
	assert.False(t, cancel)
}

func TestRestart(t *testing.T) {
	ctx := newContext(&config{}, testLog)
	err := ctx.Restart()
	assert.EqualError(t, err, "Unsupported")
}

func TestApplyNoAsset(t *testing.T) {
	ctx := newContext(&testConfigPlatform{}, testLog)
	tmpDir, err := util.MakeTempDir("TestApplyNoAsset.", 0700)
	defer util.RemoveFileAtPath(tmpDir)
	require.NoError(t, err)
	err = ctx.Apply(testUpdate, testOptions, tmpDir)
	require.NoError(t, err)
}

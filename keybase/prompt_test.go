// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/keybase/go-updater"
	"github.com/stretchr/testify/assert"
)

func testPromptWithCommand(t *testing.T, promptCommand string, timeout time.Duration) (*updater.UpdatePromptResponse, error) {
	cfg, _ := testConfig(t)
	ctx := newContext(&cfg, log)
	assert.NotNil(t, ctx)

	update := updater.Update{
		Version:     "1.2.3-400+sha",
		Name:        "Test",
		Description: "Bug fixes",
	}

	updaterOptions := cfg.updaterOptions()

	promptOptions := updater.UpdatePromptOptions{AutoUpdate: false}
	return ctx.updatePrompt(promptCommand, update, updaterOptions, promptOptions, timeout)
}

func TestPromptTimeout(t *testing.T) {
	promptCommand := filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/prompt-sleep.sh")
	resp, err := testPromptWithCommand(t, promptCommand, 10*time.Millisecond)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestPromptInvalidResponse(t *testing.T) {
	promptCommand := filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/prompt-invalid.sh")
	resp, err := testPromptWithCommand(t, promptCommand, time.Second)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestPromptApply(t *testing.T) {
	promptCommand := filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/prompt-apply.sh")
	resp, err := testPromptWithCommand(t, promptCommand, time.Second)
	assert.NoError(t, err)
	if assert.NotNil(t, resp) {
		assert.True(t, resp.AutoUpdate)
		assert.Equal(t, updater.UpdateActionApply, resp.Action)
	}
}

func TestPromptSnooze(t *testing.T) {
	promptCommand := filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/prompt-snooze.sh")
	resp, err := testPromptWithCommand(t, promptCommand, time.Second)
	assert.NoError(t, err)
	if assert.NotNil(t, resp) {
		assert.False(t, resp.AutoUpdate)
		assert.Equal(t, updater.UpdateActionSnooze, resp.Action)
	}
}

func TestPromptCancel(t *testing.T) {
	promptCommand := filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/prompt-cancel.sh")
	resp, err := testPromptWithCommand(t, promptCommand, time.Second)
	assert.NoError(t, err)
	if assert.NotNil(t, resp) {
		assert.False(t, resp.AutoUpdate)
		assert.Equal(t, updater.UpdateActionCancel, resp.Action)
	}
}

func TestPromptNoOutput(t *testing.T) {
	resp, err := testPromptWithCommand(t, "echo", time.Second)
	assert.NoError(t, err)
	if assert.NotNil(t, resp) {
		assert.False(t, resp.AutoUpdate)
		assert.Equal(t, updater.UpdateActionCancel, resp.Action)
	}
}

func TestPromptError(t *testing.T) {
	promptCommand := filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/prompt-error.sh")
	resp, err := testPromptWithCommand(t, promptCommand, time.Second)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func testPausedPromptWithCommand(t *testing.T, promptCommand string, timeout time.Duration) (bool, error) {
	cfg, _ := testConfig(t)
	ctx := newContext(&cfg, log)
	assert.NotNil(t, ctx)
	return ctx.pausedPrompt(promptCommand, timeout)
}

func TestPausedPromptForce(t *testing.T) {
	promptCommand := filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/prompt-paused-force.sh")
	cancel, err := testPausedPromptWithCommand(t, promptCommand, time.Second)
	assert.NoError(t, err)
	assert.False(t, cancel)
}

func TestPausedPromptCancel(t *testing.T) {
	promptCommand := filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/prompt-paused-cancel.sh")
	cancel, err := testPausedPromptWithCommand(t, promptCommand, time.Second)
	assert.NoError(t, err)
	assert.True(t, cancel)
}

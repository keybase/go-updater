// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build windows

package keybase

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testConfigWindows struct {
	testConfigPlatform
}

func (c testConfigWindows) destinationPath() string {
	return filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/windows/prompter/prompter.hta")
}

func TestUpdatePrompt(t *testing.T) {
	ctx := newContext(&testConfigWindows{}, testLog)
	resp, err := ctx.UpdatePrompt(testUpdate, testOptions, updater.UpdatePromptOptions{})
	assert.NoError(t, err)
	assert.Equal(t, &updater.UpdatePromptResponse{Action: updater.UpdateActionContinue}, resp)
}

func TestApplyNoAsset(t *testing.T) {
	ctx := newContext(&testConfigPlatform{}, testLog)
	tmpDir, err := util.MakeTempDir("TestApplyNoAsset.", 0700)
	defer util.RemoveFileAtPath(tmpDir)
	require.NoError(t, err)
	err = ctx.Apply(testUpdate, testOptions, tmpDir)
	require.EqualError(t, err, "No asset")
}

func TestApplyAsset(t *testing.T) {
	ctx := newContext(&testConfigPlatform{}, testLog)
	tmpDir, err := util.MakeTempDir("TestApplyAsset.", 0700)
	defer util.RemoveFileAtPath(tmpDir)
	require.NoError(t, err)

	exePath := filepath.Join(os.Getenv("GOPATH"), "bin", "test.exe")
	localPath := filepath.Join(tmpDir, "test.exe")
	err = util.CopyFile(exePath, localPath, testLog)
	require.NoError(t, err)

	update := updater.Update{
		Asset: &updater.Asset{
			LocalPath: exePath,
		},
	}

	err = ctx.Apply(update, updater.UpdateOptions{}, tmpDir)
	require.NoError(t, err)
}

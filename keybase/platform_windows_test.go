// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build windows

package keybase

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdatePrompt(t *testing.T) {
	outPath := util.TempPath("", "TestUpdatePrompt.")
	defer util.RemoveFileAtPath(outPath)
	promptOptions := updater.UpdatePromptOptions{OutPath: outPath}
	out := `{"action":"apply","autoUpdate":true}` + "\n"

	programPath := filepath.Join(os.Getenv("GOPATH"), "bin", "test.exe")
	args := []string{
		fmt.Sprintf("-out=%s", out),
		fmt.Sprintf("-outPath=%s", outPath),
		"writeToFile"}
	ctx := newContext(&testConfigPlatform{ProgramPath: programPath, Args: args}, testLog)
	resp, err := ctx.UpdatePrompt(testUpdate, testOptions, promptOptions)
	require.NoError(t, err)
	assert.Equal(t, &updater.UpdatePromptResponse{Action: updater.UpdateActionApply, AutoUpdate: true}, resp)
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

const legit64Code = "{65A3A964-3DC3-0100-0000-160621082245}"
const legit86Code = "{65A3A986-3DC3-0100-0000-160621082245}"
const mismatchedCode = "{65A3A986-3DC3-0100-0000-160621082244}"

func testCheckCanBeSlient(t *testing.T, dokan86Code string, dokan64Code string, testcode string) bool {
	result, err := CheckCanBeSilent(dokan86Code, dokan64Code, testLog, func(s string, l Log) bool { return s == testcode })
	require.NoError(t, err)
	return result
}

func TestSearchInstallerLayout(t *testing.T) {
	assert.Equal(t, testCheckCanBeSlient(t, legit86Code, legit64Code, legit64Code), true)
	assert.Equal(t, testCheckCanBeSlient(t, legit86Code, legit64Code, legit86Code), true)
	// This is now true because there's no dokan1.dll
	assert.Equal(t, testCheckCanBeSlient(t, legit86Code, legit64Code, mismatchedCode), true)
}

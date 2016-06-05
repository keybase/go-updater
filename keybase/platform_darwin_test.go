// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build darwin

package keybase

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/process"
	"github.com/keybase/go-updater/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppBundleForPath(t *testing.T) {
	assert.Equal(t, "", appBundleForPath(""))
	assert.Equal(t, "", appBundleForPath("foo"))
	assert.Equal(t, "/Applications/Keybase.app", appBundleForPath("/Applications/Keybase.app"))
	assert.Equal(t, "/Applications/Keybase.app", appBundleForPath("/Applications/Keybase.app/Contents/SharedSupport/bin/keybase"))
	assert.Equal(t, "/Applications/Keybase.app", appBundleForPath("/Applications/Keybase.app/Contents/Resources/Foo.app/Contents/MacOS/Foo"))
	assert.Equal(t, "", appBundleForPath("/Applications/Keybase.ap"))
	assert.Equal(t, "/Applications/Keybase.app", appBundleForPath("/Applications/Keybase.app/"))
}

type testConfigDarwin struct {
	testConfigPlatform
}

func (c testConfigDarwin) destinationPath() string {
	return filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/Test.app")
}

func TestUpdatePrompt(t *testing.T) {
	ctx := newContext(&testConfigPlatform{}, testLog)
	resp, err := ctx.UpdatePrompt(testUpdate, testOptions, updater.UpdatePromptOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestRestart(t *testing.T) {
	ctx := newContext(&testConfigDarwin{}, testLog)
	appPath := ctx.config.destinationPath()

	err := OpenAppDarwin(appPath, testLog)
	defer func() {
		process.TerminateAll(process.NewMatcher(appPath, process.PathPrefix, testLog), 200*time.Millisecond, testLog)
	}()
	require.NoError(t, err)

	// TODO: We don't have watchdog available in tests yet, coming next, so let's
	// test the error that the app was ok, but the services didn't restart.
	err = ctx.restart(20*time.Millisecond, 20*time.Millisecond)
	assert.EqualError(t, err, "There were multiple errors: No process found for Test.app/Contents/SharedSupport/bin/keybase; No process found for Test.app/Contents/SharedSupport/bin/kbfs")
}

func TestOpenDarwin(t *testing.T) {
	appPath := filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/Test.app")
	matcher := process.NewMatcher(appPath, process.PathPrefix, testLog)
	defer process.TerminateAll(matcher, 200*time.Millisecond, testLog)
	err := OpenAppDarwin(appPath, testLog)
	assert.NoError(t, err)
}

func TestOpenDarwinError(t *testing.T) {
	binErr := filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/err.sh")
	appPath := filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/Test.app")
	err := openAppDarwin(binErr, appPath, time.Millisecond, testLog)
	assert.Error(t, err)
}

func TestFindPIDsLaunchd(t *testing.T) {
	procPath := "/sbin/launchd"
	matcher := process.NewMatcher(procPath, process.PathEqual, testLog)
	pids, err := process.FindPIDsWithMatchFn(matcher.Fn(), testLog)
	assert.NoError(t, err)
	t.Logf("Pids: %#v", pids)
	require.True(t, len(pids) >= 1)
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

	zipPath := filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/test.zip")
	localPath := filepath.Join(tmpDir, "test.zip")
	err = util.CopyFile(zipPath, localPath, testLog)
	require.NoError(t, err)

	update := updater.Update{
		Asset: &updater.Asset{
			LocalPath: zipPath,
		},
	}

	options := updater.UpdateOptions{DestinationPath: filepath.Join(os.TempDir(), "test")}

	err = ctx.Apply(update, options, tmpDir)
	require.NoError(t, err)
}

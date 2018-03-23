// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build !windows

package util

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnzipOtherUser checks to make sure that a zip file created from a
// different uid has the current uid after unpacking.
func TestUnzipOtherUser(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unsupported on windows")
	}
	var testZipOtherUserPath = filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/test-uid-503.zip")
	destinationPath := TempPath("", "TestUnzipOtherUser.")
	err := Unzip(testZipOtherUserPath, destinationPath, testLog)
	require.NoError(t, err)

	// Get uid, gid of current user
	currentUser, err := user.Current()
	require.NoError(t, err)
	uid, err := strconv.Atoi(currentUser.Uid)
	require.NoError(t, err)

	fileInfo, err := os.Stat(filepath.Join(destinationPath, "test"))
	require.NoError(t, err)
	fileUID := fileInfo.Sys().(*syscall.Stat_t).Uid
	assert.Equal(t, uid, int(fileUID))
}

// TestUnzipFileModTime checks to make sure after unpacking zip file the file
// modification time is "now" and not the original file time.
func TestUnzipFileModTime(t *testing.T) {
	now := time.Now()
	nowUnix := now.UnixNano()
	t.Logf("Now: %d", nowUnix)
	destinationPath := TempPath("", "TestUnzipFileModTime.")
	err := Unzip(testZipPath, destinationPath, testLog)
	require.NoError(t, err)

	fileInfo, err := os.Stat(filepath.Join(destinationPath, "test"))
	require.NoError(t, err)
	dirMod := fileInfo.ModTime()
	diffDir := nowUnix - dirMod.UnixNano()
	t.Logf("Diff (dir): %d", diffDir)
	assert.True(t, diffDir >= 0, "now=%s, dirtime=%s", now, dirMod)

	fileInfo, err = os.Stat(filepath.Join(destinationPath, "test", "testfile"))
	require.NoError(t, err)
	fileMod := fileInfo.ModTime()
	diffFile := nowUnix - fileMod.UnixNano()
	t.Logf("Diff (file): %d", diffFile)
	assert.True(t, diffFile >= 0, "now=%s, filetime=%s", now, fileMod)
}

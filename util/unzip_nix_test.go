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

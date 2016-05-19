// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package util

import (
	"fmt"
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

// testZipPath is a valid zip file
var testZipPath = filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/test.zip")

// testCorruptedZipPath is a corrupted zip file (flipped a bit)
var testCorruptedZipPath = filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/test-corrupted2.zip")

// testInvalidZipPath is not a valid zip file
var testInvalidZipPath = filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/test-invalid.zip")

func assertFileExists(t *testing.T, path string) {
	t.Logf("Checking %s", path)
	fileExists, err := FileExists(path)
	assert.NoError(t, err)
	assert.True(t, fileExists)
}

func TestUnzipOverValid(t *testing.T) {
	destinationPath := TempPath("", "TestUnzipOver.")
	defer RemoveFileAtPath(destinationPath)

	noCheck := func(sourcePath, destinationPath string) error { return nil }

	err := UnzipOver(testZipPath, "test", destinationPath, noCheck, testLog)
	assert.NoError(t, err)

	dirExists, err := FileExists(destinationPath)
	assert.NoError(t, err)
	assert.True(t, dirExists)

	assertFileExists(t, filepath.Join(destinationPath, "testfile"))
	assertFileExists(t, filepath.Join(destinationPath, "testfolder"))
	assertFileExists(t, filepath.Join(destinationPath, "testfolder", "testsubfolder"))
	assertFileExists(t, filepath.Join(destinationPath, "testfolder", "testsubfolder", "testfile2"))
	assertFileExists(t, filepath.Join(destinationPath, "testfolder", "testlink"))

	// Unzip again over existing path
	err = UnzipOver(testZipPath, "test", destinationPath, noCheck, testLog)
	assert.NoError(t, err)

	dirExists2, err := FileExists(destinationPath)
	assert.NoError(t, err)
	assert.True(t, dirExists2)

	fileExists2, err := FileExists(filepath.Join(destinationPath, "testfile"))
	assert.NoError(t, err)
	assert.True(t, fileExists2)

	// Unzip again over existing path, fail check
	failCheck := func(sourcePath, destinationPath string) error { return fmt.Errorf("Failed check") }
	err = UnzipOver(testZipPath, "test", destinationPath, failCheck, testLog)
	assert.Error(t, err)

	err = unzipOver(testZipPath, destinationPath, testLog)
	assert.NoError(t, err)
}

func TestUnzipOverInvalidPath(t *testing.T) {
	noCheck := func(sourcePath, destinationPath string) error { return nil }
	err := UnzipOver(testZipPath, "test", "", noCheck, testLog)
	assert.Error(t, err)

	destinationPath := TempPath("", "TestUnzipOverInvalidPath.")
	defer RemoveFileAtPath(destinationPath)
	err = UnzipOver("/badfile.zip", "test", destinationPath, noCheck, testLog)
	assert.Error(t, err)

	err = UnzipOver("", "test", destinationPath, noCheck, testLog)
	assert.Error(t, err)

	err = unzipOver("", "", testLog)
	assert.Error(t, err)
}

func TestUnzipOverInvalidZip(t *testing.T) {
	noCheck := func(sourcePath, destinationPath string) error { return nil }
	destinationPath := TempPath("", "TestUnzipOverInvalidZip.")
	defer RemoveFileAtPath(destinationPath)
	err := UnzipOver(testInvalidZipPath, "test", destinationPath, noCheck, testLog)
	t.Logf("Error: %s", err)
	assert.Error(t, err)
}

func TestUnzipOverInvalidContents(t *testing.T) {
	noCheck := func(sourcePath, destinationPath string) error { return nil }
	destinationPath := TempPath("", "TestUnzipOverInvalidContents.")
	defer RemoveFileAtPath(destinationPath)
	err := UnzipOver(testInvalidZipPath, "invalid", destinationPath, noCheck, testLog)
	t.Logf("Error: %s", err)
	assert.Error(t, err)
}

func TestUnzipOverCorrupted(t *testing.T) {
	noCheck := func(sourcePath, destinationPath string) error { return nil }
	destinationPath := TempPath("", "TestUnzipOverCorrupted.")
	defer RemoveFileAtPath(destinationPath)
	err := UnzipOver(testCorruptedZipPath, "test", destinationPath, noCheck, testLog)
	t.Logf("Error: %s", err)
	assert.Error(t, err)
}

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
	now := time.Now().UnixNano()
	destinationPath := TempPath("", "TestUnzipFileModTime.")
	err := Unzip(testZipPath, destinationPath, testLog)
	require.NoError(t, err)

	fileInfo, err := os.Stat(filepath.Join(destinationPath, "test"))
	require.NoError(t, err)
	assert.True(t, now-fileInfo.ModTime().UnixNano() > 0)

	fileInfo, err = os.Stat(filepath.Join(destinationPath, "test", "testfile"))
	require.NoError(t, err)
	assert.True(t, now-fileInfo.ModTime().UnixNano() > 0)
}

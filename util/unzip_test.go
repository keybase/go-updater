// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// testZipPath is a valid zip file
var testZipPath = filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/test.zip")

// testCorruptedZipPath is a corrupted zip file (flipped a bit)
var testCorruptedZipPath = filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/test-corrupted2.zip")

// testInvalidZipPath is not a valid zip file
var testInvalidZipPath = filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/test-invalid.zip")

func TestUnzipOver(t *testing.T) {
	destinationPath, err := TempPath("TestUnzipOver.")
	t.Logf("Destination: %s", destinationPath)
	defer RemoveFileAtPath(destinationPath)
	assert.NoError(t, err)
	err = UnzipOver(testZipPath, destinationPath, log)
	assert.NoError(t, err)

	exists, err := FileExists(destinationPath)
	assert.True(t, exists)
	assert.NoError(t, err)

	fileExists, err := FileExists(filepath.Join(destinationPath, "test", "testfile"))
	assert.True(t, fileExists)
	assert.NoError(t, err)

	// Unzip again over existing path
	err = UnzipOver(testZipPath, destinationPath, log)
	assert.NoError(t, err)

	fileExist2, err := FileExists(destinationPath)
	assert.True(t, fileExist2)
	assert.NoError(t, err)

	fileExists3, err := FileExists(filepath.Join(destinationPath, "test", "testfile"))
	assert.True(t, fileExists3)
	assert.NoError(t, err)
}

func TestUnzipOverInvalidPath(t *testing.T) {
	err := UnzipOver(testZipPath, "", log)
	assert.Error(t, err)

	destinationPath, err := TempPath("TestUnzipOver.")
	assert.NoError(t, err)

	err = UnzipOver("/badfile.zip", destinationPath, log)
	assert.Error(t, err)

	err = UnzipOver("", destinationPath, log)
	assert.Error(t, err)
}

func TestUnzipOverInvalid(t *testing.T) {
	destinationPath, err := TempPath("TestUnzipOverInvalid.")
	assert.NoError(t, err)
	err = UnzipOver(testInvalidZipPath, destinationPath, log)
	t.Logf("Error: %s", err)
	assert.Error(t, err)
}

func TestUnzipOverCorrupted(t *testing.T) {
	destinationPath, err := TempPath("TestUnzipOverCorrupted.")
	assert.NoError(t, err)
	err = UnzipOver(testCorruptedZipPath, destinationPath, log)
	t.Logf("Error: %s", err)
	assert.Error(t, err)
}

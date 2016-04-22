// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package util

import (
	"fmt"
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

func TestUnzipOverValid(t *testing.T) {
	destinationPath := TempPath("", "TestUnzipOver.")
	defer RemoveFileAtPath(destinationPath)

	noCheck := func(sourcePath, destinationPath string) error { return nil }

	err := UnzipOver(testZipPath, "test", destinationPath, noCheck, log)
	assert.NoError(t, err)

	dirExists, err := FileExists(destinationPath)
	assert.NoError(t, err)
	assert.True(t, dirExists)

	fileExists, err := FileExists(filepath.Join(destinationPath, "testfile"))
	assert.NoError(t, err)
	assert.True(t, fileExists)

	// Unzip again over existing path
	err = UnzipOver(testZipPath, "test", destinationPath, noCheck, log)
	assert.NoError(t, err)

	dirExists2, err := FileExists(destinationPath)
	assert.NoError(t, err)
	assert.True(t, dirExists2)

	fileExists2, err := FileExists(filepath.Join(destinationPath, "testfile"))
	assert.NoError(t, err)
	assert.True(t, fileExists2)

	// Unzip again over existing path, fail check
	failCheck := func(sourcePath, destinationPath string) error { return fmt.Errorf("Failed check") }
	// Unzip again over existing path
	err = UnzipOver(testZipPath, "test", destinationPath, failCheck, log)
	assert.Error(t, err)
}

func TestUnzipOverInvalidPath(t *testing.T) {
	noCheck := func(sourcePath, destinationPath string) error { return nil }
	err := UnzipOver(testZipPath, "test", "", noCheck, log)
	assert.Error(t, err)

	destinationPath := TempPath("", "TestUnzipOverInvalidPath.")
	defer RemoveFileAtPath(destinationPath)
	err = UnzipOver("/badfile.zip", "test", destinationPath, noCheck, log)
	assert.Error(t, err)

	err = UnzipOver("", "test", destinationPath, noCheck, log)
	assert.Error(t, err)
}

func TestUnzipOverInvalidZip(t *testing.T) {
	noCheck := func(sourcePath, destinationPath string) error { return nil }
	destinationPath := TempPath("", "TestUnzipOverInvalidZip.")
	defer RemoveFileAtPath(destinationPath)
	err := UnzipOver(testInvalidZipPath, "test", destinationPath, noCheck, log)
	t.Logf("Error: %s", err)
	assert.Error(t, err)
}

func TestUnzipOverInvalidContents(t *testing.T) {
	noCheck := func(sourcePath, destinationPath string) error { return nil }
	destinationPath := TempPath("", "TestUnzipOverInvalidContents.")
	defer RemoveFileAtPath(destinationPath)
	err := UnzipOver(testInvalidZipPath, "invalid", destinationPath, noCheck, log)
	t.Logf("Error: %s", err)
	assert.Error(t, err)
}

func TestUnzipOverCorrupted(t *testing.T) {
	noCheck := func(sourcePath, destinationPath string) error { return nil }
	destinationPath := TempPath("", "TestUnzipOverCorrupted.")
	defer RemoveFileAtPath(destinationPath)
	err := UnzipOver(testCorruptedZipPath, "test", destinationPath, noCheck, log)
	t.Logf("Error: %s", err)
	assert.Error(t, err)
}

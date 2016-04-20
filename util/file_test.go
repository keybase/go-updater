// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/keybase/go-logging"
	"github.com/stretchr/testify/assert"
)

var log = logging.Logger{Module: "test"}

func TestNewFile(t *testing.T) {
	filename := filepath.Join(os.TempDir(), "TestNewFile")
	defer RemoveFileAtPath(filename)

	f := NewFile(filename, []byte("somedata"), 0600)
	err := f.Save(log)
	assert.NoError(t, err)

	fileInfo, err := os.Stat(filename)
	assert.NoError(t, err)
	assert.True(t, 0600 == fileInfo.Mode().Perm())
	assert.False(t, fileInfo.IsDir())
}

func TestMakeParentDirs(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "TestMakeParentDirs", "TestMakeParentDirs2", "TestMakeParentDirs3")
	defer RemoveFileAtPath(dir)

	file := filepath.Join(dir, "testfile")
	defer RemoveFileAtPath(file)

	err := MakeParentDirs(file, 0700)
	assert.NoError(t, err)

	exists, err := FileExists(dir)
	assert.NoError(t, err)
	assert.True(t, exists, "File doesn't exist")

	fileInfo, err := os.Stat(dir)
	assert.NoError(t, err)
	assert.True(t, 0700 == fileInfo.Mode().Perm())
	assert.True(t, fileInfo.IsDir())

	// Test making dir that already exists
	err = MakeParentDirs(file, 0700)
	assert.NoError(t, err)
}

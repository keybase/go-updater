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
	defer os.Remove(filename)

	f := NewFile(filename, []byte("somedata"), 0)
	err := f.Save(log)
	assert.NoError(t, err)
}

func TestMakeParentDirs(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "TestMakeParentDirs")
	defer os.Remove(dir)

	file := filepath.Join(dir, "testfile")
	defer RemoveFileAtPath(file)

	err := MakeParentDirs(file)
	assert.NoError(t, err)

	exists, err := FileExists(dir)
	assert.NoError(t, err)
	assert.True(t, exists, "File doesn't exist")

	// Test making dir that already exists
	err = MakeParentDirs(file)
	assert.NoError(t, err)
}

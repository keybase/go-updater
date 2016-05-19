// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFile(t *testing.T) {
	filename := filepath.Join(os.TempDir(), "TestNewFile")
	defer RemoveFileAtPath(filename)

	f := NewFile(filename, []byte("somedata"), 0600)
	err := f.Save(testLog)
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

	err := MakeParentDirs(file, 0700, testLog)
	assert.NoError(t, err)

	exists, err := FileExists(dir)
	assert.NoError(t, err)
	assert.True(t, exists, "File doesn't exist")

	fileInfo, err := os.Stat(dir)
	assert.NoError(t, err)
	assert.True(t, 0700 == fileInfo.Mode().Perm())
	assert.True(t, fileInfo.IsDir())

	// Test making dir that already exists
	err = MakeParentDirs(file, 0700, testLog)
	assert.NoError(t, err)
}

func TestMakeParentDirsInvalid(t *testing.T) {
	err := MakeParentDirs("\\\\invalid", 0700, testLog)
	assert.EqualError(t, err, "No base directory")
}

func TestTempPathValid(t *testing.T) {
	tempPath := TempPath("", "TempPrefix.")
	t.Logf("Temp path: %s", tempPath)
	assert.True(t, strings.HasPrefix(filepath.Base(tempPath), "TempPrefix."))
	assert.Equal(t, len(filepath.Base(tempPath)), 43)
}

func TestTempPathRandFail(t *testing.T) {
	// Replace rand.Read with a failing read
	defaultRandRead := randRead
	defer func() { randRead = defaultRandRead }()
	randRead = func(b []byte) (int, error) {
		return 0, fmt.Errorf("Test rand failure")
	}

	tempPath := TempPath("", "TempPrefix.")
	t.Logf("Temp path: %s", tempPath)
	assert.True(t, strings.HasPrefix(filepath.Base(tempPath), "TempPrefix."))
	assert.Equal(t, len(filepath.Base(tempPath)), 30)
}

func TestIsDirReal(t *testing.T) {
	ok, err := IsDirReal("/invalid")
	assert.Error(t, err)
	assert.False(t, ok)

	path := os.Getenv("GOPATH")
	ok, err = IsDirReal(path)
	assert.NoError(t, err)
	assert.True(t, ok)

	testFile := filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/test.zip")
	ok, err = IsDirReal(testFile)
	assert.Error(t, err)
	assert.Equal(t, "Path is not a directory", err.Error())
	assert.False(t, ok)

	symLinkPath := TempPath("", "TestIsDirReal")
	err = os.Symlink(os.TempDir(), symLinkPath)
	defer RemoveFileAtPath(symLinkPath)
	assert.NoError(t, err)
	ok, err = IsDirReal(symLinkPath)
	assert.Error(t, err)
	assert.Equal(t, "Path is a symlink", err.Error())
	assert.False(t, ok)
}

func TestMoveFileValid(t *testing.T) {
	destinationPath := filepath.Join(TempPath("", "TestMoveFileDestination"), "TestMoveFileDestinationSubdir")
	defer RemoveFileAtPath(destinationPath)

	sourcePath, err := WriteTempFile("TestMoveFile", []byte("test"), 0600)
	defer RemoveFileAtPath(sourcePath)
	assert.NoError(t, err)

	err = MoveFile(sourcePath, destinationPath, "", testLog)
	assert.NoError(t, err)
	exists, err := FileExists(destinationPath)
	assert.NoError(t, err)
	assert.True(t, exists)
	data, err := ioutil.ReadFile(destinationPath)
	assert.NoError(t, err)
	assert.Equal(t, []byte("test"), data)
	srcExists, err := FileExists(sourcePath)
	assert.NoError(t, err)
	assert.False(t, srcExists)

	// Move again with different source data, and overwrite
	sourcePath2, err := WriteTempFile("TestMoveFile", []byte("test2"), 0600)
	err = MoveFile(sourcePath2, destinationPath, "", testLog)
	assert.NoError(t, err)
	exists, err = FileExists(destinationPath)
	assert.NoError(t, err)
	assert.True(t, exists)
	data2, err := ioutil.ReadFile(destinationPath)
	assert.NoError(t, err)
	assert.Equal(t, []byte("test2"), data2)
	srcExists2, err := FileExists(sourcePath2)
	assert.NoError(t, err)
	assert.False(t, srcExists2)
}

func TestMoveFileDirValid(t *testing.T) {
	destinationPath := filepath.Join(TempPath("", "TestMoveFileDestination"), "TestMoveFileDestinationSubdir")
	defer RemoveFileAtPath(destinationPath)

	sourcePath, err := WriteTempDir("TestMoveDir", 0700)
	defer RemoveFileAtPath(sourcePath)
	assert.NoError(t, err)

	err = MoveFile(sourcePath, destinationPath, "", testLog)
	assert.NoError(t, err)
	exists, err := FileExists(destinationPath)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Move again with different source data, and overwrite
	sourcePath2, err := WriteTempDir("TestMoveDir", 0700)
	err = MoveFile(sourcePath2, destinationPath, "", testLog)
	assert.NoError(t, err)
	exists, err = FileExists(destinationPath)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestMoveFileInvalidSource(t *testing.T) {
	sourcePath := "/invalid"
	destinationPath := TempPath("", "TestMoveFileDestination")
	err := MoveFile(sourcePath, destinationPath, "", testLog)
	assert.Error(t, err)

	exists, err := FileExists(destinationPath)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestMoveFileInvalidDest(t *testing.T) {
	sourcePath := "/invalid"
	destinationPath := TempPath("", "TestMoveFileDestination")
	err := MoveFile(sourcePath, destinationPath, "", testLog)
	assert.Error(t, err)

	exists, err := FileExists(destinationPath)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestCopyFileValid(t *testing.T) {
	destinationPath := filepath.Join(TempPath("", "TestCopyFileDestination"), "TestCopyFileDestinationSubdir")
	defer RemoveFileAtPath(destinationPath)

	sourcePath, err := WriteTempFile("TestCopyFile", []byte("test"), 0600)
	defer RemoveFileAtPath(sourcePath)
	assert.NoError(t, err)

	err = CopyFile(sourcePath, destinationPath, testLog)
	assert.NoError(t, err)
	exists, err := FileExists(destinationPath)
	assert.NoError(t, err)
	assert.True(t, exists)
	data, err := ioutil.ReadFile(destinationPath)
	assert.NoError(t, err)
	assert.Equal(t, []byte("test"), data)

	// Move again with different source data, and overwrite
	sourcePath2, err := WriteTempFile("TestCopyFile", []byte("test2"), 0600)
	err = CopyFile(sourcePath2, destinationPath, testLog)
	assert.NoError(t, err)
	exists, err = FileExists(destinationPath)
	assert.NoError(t, err)
	assert.True(t, exists)
	data2, err := ioutil.ReadFile(destinationPath)
	assert.NoError(t, err)
	assert.Equal(t, []byte("test2"), data2)
}

func TestCopyFileInvalidSource(t *testing.T) {
	sourcePath := "/invalid"
	destinationPath := TempPath("", "TestCopyFileDestination")
	err := CopyFile(sourcePath, destinationPath, testLog)
	assert.Error(t, err)

	exists, err := FileExists(destinationPath)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestCopyFileInvalidDest(t *testing.T) {
	sourcePath := "/invalid"
	destinationPath := TempPath("", "TestCopyFileDestination")
	err := CopyFile(sourcePath, destinationPath, testLog)
	assert.Error(t, err)

	exists, err := FileExists(destinationPath)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestCloseNil(t *testing.T) {
	Close(nil)
}

func TestOpenTempFile(t *testing.T) {
	path, tempFile, err := openTempFile("prefix", "suffix", 0)
	defer Close(tempFile)
	defer RemoveFileAtPath(path)
	require.NoError(t, err)
	require.NotNil(t, tempFile)

	basePath := filepath.Base(path)
	assert.True(t, strings.HasPrefix(basePath, "prefix"))
	assert.True(t, strings.HasSuffix(basePath, "suffix"))
}

func TestFileExists(t *testing.T) {
	exists, err := FileExists("/nope")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestReadFile(t *testing.T) {
	dataIn := []byte("test")
	sourcePath, err := WriteTempFile("TestReadFile", dataIn, 0600)

	dataOut, err := ReadFile(sourcePath)
	require.NoError(t, err)
	assert.Equal(t, dataIn, dataOut)

	_, err = ReadFile("/invalid")
	assert.EqualError(t, err, "open /invalid: no such file or directory")
}

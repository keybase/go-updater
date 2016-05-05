// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/keybase/go-logging"
)

// File uses a safer file API
type File struct {
	name string
	data []byte
	perm os.FileMode
}

// SafeWriter defines a writer that is safer (atomic)
type SafeWriter interface {
	GetFilename() string
	WriteTo(io.Writer) (int64, error)
}

// NewFile returns a File
func NewFile(name string, data []byte, perm os.FileMode) File {
	return File{name, data, perm}
}

// Save file
func (f File) Save(log logging.Logger) error {
	return safeWriteToFile(f, f.perm, log)
}

// GetFilename returns the file name for SafeWriter
func (f File) GetFilename() string {
	return f.name
}

// WriteTo is for SafeWriter
func (f File) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(f.data)
	return int64(n), err
}

// safeWriteToFile to safely write to a file
func safeWriteToFile(t SafeWriter, mode os.FileMode, log logging.Logger) error {
	filename := t.GetFilename()
	if filename == "" {
		return fmt.Errorf("No filename")
	}
	log.Debugf("Writing to %s", filename)
	tempFilename, tempFile, err := openTempFile(filename+"-", "", mode)
	log.Debugf("Temporary file generated: %s", tempFilename)
	if err != nil {
		return err
	}
	_, err = t.WriteTo(tempFile)
	if err != nil {
		log.Errorf("Error writing temporary file %s: %s", tempFilename, err)
		_ = tempFile.Close()
		_ = os.Remove(tempFilename)
		return err
	}
	err = tempFile.Close()
	if err != nil {
		log.Errorf("Error closing temporary file %s: %s", tempFilename, err)
		_ = os.Remove(tempFilename)
		return err
	}
	err = os.Rename(tempFilename, filename)
	if err != nil {
		log.Errorf("Error renaming temporary file %s to %s: %s", tempFilename, filename, err)
		_ = os.Remove(tempFilename)
		return err
	}
	log.Debugf("Wrote to %s", filename)
	return nil
}

// Close closes a file and ignores the error.
// This satisfies lint checks when using with defer and you don't care if there
// is an error, so instead of:
//   defer func() { _ = f.Close() }()
//   defer Close(f)
func Close(f io.Closer) {
	if f == nil {
		return
	}
	_ = f.Close()
}

// RemoveFileAtPath removes a file at path (and any children) ignoring any error.
// This satisfies lint checks when using with defer and you don't care if there
// is an error, so instead of:
//   defer func() { _ = os.Remove(path) }()
//   defer RemoveFileAtPath(path)
func RemoveFileAtPath(path string) {
	_ = os.RemoveAll(path)
}

// openTempFile creates an opened temporary file.
//
//   openTempFile("foo", ".zip", 0755) => "foo.RCG2KUSCGYOO3PCKNWQHBOXBKACOPIKL.zip"
//   openTempFile(path.Join(os.TempDir(), "foo"), "", 0600) => "/tmp/foo.RCG2KUSCGYOO3PCKNWQHBOXBKACOPIKL"
//
func openTempFile(prefix string, suffix string, mode os.FileMode) (string, *os.File, error) {
	filename, err := RandString(prefix, 20)
	if err != nil {
		return "", nil, err
	}
	if suffix != "" {
		filename = filename + suffix
	}
	flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if mode == 0 {
		mode = 0600
	}
	file, err := os.OpenFile(filename, flags, mode)
	return filename, file, err
}

// FileExists returns whether the given file or directory exists or not
func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// MakeParentDirs ensures parent directory exist for path
func MakeParentDirs(path string, mode os.FileMode, log logging.Logger) error {
	// 2nd return value here is filename (not an error), which is not needed
	dir, _ := filepath.Split(path)
	if dir == "" {
		return fmt.Errorf("No base directory")
	}
	exists, err := FileExists(dir)
	if err != nil {
		return err
	}

	if !exists {
		log.Debugf("Creating: %s\n", dir)
		err = os.MkdirAll(dir, mode)
		if err != nil {
			return err
		}
	}
	return nil
}

// TempPath returns a temporary unique file path.
// If for some reason we can't obtain random bytes, we still return a valid
// path, which may not be as unique.
// If tempDir is "", then os.TempDir() is used.
func TempPath(tempDir string, prefix string) string {
	if tempDir == "" {
		tempDir = os.TempDir()
	}
	filename, err := RandString(prefix, 20)
	if err != nil {
		// We had an error getting random bytes, we'll use current nanoseconds
		filename = fmt.Sprintf("%s%d", prefix, time.Now().UnixNano())
	}
	path := filepath.Join(tempDir, filename)
	return path
}

// WriteTempFile creates a unique temp file with data.
//
// For example:
//   WriteTempFile("Test.", byte[]("test data"), 0600)
func WriteTempFile(prefix string, data []byte, mode os.FileMode) (string, error) {
	path := TempPath("", prefix)
	if err := ioutil.WriteFile(path, data, mode); err != nil {
		return "", err
	}
	return path, nil
}

// WriteTempDir creates a unique temp directory.
//
// For example:
//   WriteTempDir("Test.", 0600)
func WriteTempDir(prefix string, mode os.FileMode) (string, error) {
	path := TempPath("", prefix)
	if err := os.MkdirAll(path, mode); err != nil {
		return "", err
	}
	return path, nil
}

// IsDirReal returns true if directory exists and is a real directory (not a symlink).
// If it returns false, an error will be set explaining why.
func IsDirReal(path string) (bool, error) {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return false, err
	}
	// Check if symlink
	if fileInfo.Mode()&os.ModeSymlink != 0 {
		return false, fmt.Errorf("Path is a symlink")
	}
	if !fileInfo.Mode().IsDir() {
		return false, fmt.Errorf("Path is not a directory")
	}
	return true, nil
}

// MoveFile moves a file safely.
// It will create parent directories for destinationPath if they don't exist.
// It will overwrite an existing destinationPath.
func MoveFile(sourcePath string, destinationPath string, log logging.Logger) error {
	if _, statErr := os.Stat(destinationPath); statErr == nil {
		log.Infof("Removing existing destination path: %s", destinationPath)
		if removeErr := os.RemoveAll(destinationPath); removeErr != nil {
			return removeErr
		}
	}

	if err := MakeParentDirs(destinationPath, 0700, log); err != nil {
		return err
	}

	log.Infof("Moving %s to %s", sourcePath, destinationPath)
	// Rename will copy over an existing destination
	return os.Rename(sourcePath, destinationPath)
}

// CopyFile copies a file safely.
// It will create parent directories for destinationPath if they don't exist.
// It will overwrite an existing destinationPath.
func CopyFile(sourcePath string, destinationPath string, log logging.Logger) error {
	log.Infof("Copying %s to %s", sourcePath, destinationPath)
	in, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer Close(in)

	if _, statErr := os.Stat(destinationPath); statErr == nil {
		log.Infof("Removing existing destination path: %s", destinationPath)
		if removeErr := os.RemoveAll(destinationPath); removeErr != nil {
			return removeErr
		}
	}

	if err := MakeParentDirs(destinationPath, 0700, log); err != nil {
		return err
	}

	out, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer Close(out)
	_, err = io.Copy(out, in)
	closeErr := out.Close()
	if err != nil {
		return err
	}
	return closeErr
}

// ReadFile returns data for file at path
func ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer Close(file)
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

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
func Close(f *os.File) {
	if f == nil {
		return
	}
	_ = f.Close()
}

// RemoveFileAtPath removes a file at path and ignores any error.
// This satisfies lint checks when using with defer and you don't care if there
// is an error, so instead of:
//   defer func() { _ = os.Remove(path) }()
//   defer RemoveFileAtPath(path)
func RemoveFileAtPath(path string) {
	_ = os.Remove(path)
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
func MakeParentDirs(path string, mode os.FileMode) error {
	// 2nd return value here is filename (not an error), which is not needed
	dir, _ := filepath.Split(path)
	exists, err := FileExists(dir)
	if err != nil {
		return err
	}

	if !exists {
		fmt.Printf("Creating: %s\n", dir)
		err = os.MkdirAll(dir, mode)
		if err != nil {
			return err
		}
	}
	return nil
}

// TempPath returns a temporary unique file path
func TempPath(prefix string) (string, error) {
	filename, err := RandString(prefix, 20)
	if err != nil {
		return "", err
	}
	path := filepath.Join(os.TempDir(), filename)
	return path, nil
}

// WriteTempFile creates a temp file with data
func WriteTempFile(prefix string, data []byte, mode os.FileMode) (string, error) {
	path, err := TempPath(prefix)
	if err != nil {
		return "", err
	}
	err = ioutil.WriteFile(path, data, mode)
	if err != nil {
		return "", err
	}
	return path, nil
}

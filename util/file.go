// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package util

import (
	"io"
	"os"

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

// safeWriteToFile to safely write to a file. Use mode=0 for default permissions.
func safeWriteToFile(t SafeWriter, mode os.FileMode, log logging.Logger) error {
	fn := t.GetFilename()
	log.Debugf("Writing to %s", fn)
	tmpfn, tmp, err := OpenTempFile(fn, "", mode)
	log.Debugf("Temporary file generated: %s", tmpfn)
	if err != nil {
		return err
	}
	_, err = t.WriteTo(tmp)
	if err != nil {
		log.Errorf("Error writing temporary file %s: %s", tmpfn, err)
		_ = tmp.Close()
		_ = os.Remove(tmpfn)
		return err
	}
	err = tmp.Close()
	if err != nil {
		log.Errorf("Error closing temporary file %s: %s", tmpfn, err)
		_ = os.Remove(tmpfn)
		return err
	}
	err = os.Rename(tmpfn, fn)
	if err != nil {
		log.Errorf("Error renaming temporary file %s to %s: %s", tmpfn, fn, err)
		_ = os.Remove(tmpfn)
		return err
	}
	log.Debugf("Wrote to %s", fn)
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

// OpenTempFile creates an opened temporary file. Use mode=0 for default
// permission (0600).
//
//   OpenTempFile("foo", ".zip", 0755) => "foo.RCG2KUSCGYOO3PCKNWQHBOXBKACOPIKL.zip"
//   OpenTempFile(path.Join(os.TempDir(), "foo"), "", 0) => "/tmp/foo.RCG2KUSCGYOO3PCKNWQHBOXBKACOPIKL"
//
func OpenTempFile(prefix string, suffix string, mode os.FileMode) (string, *os.File, error) {
	if prefix != "" {
		prefix = prefix + "."
	}
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

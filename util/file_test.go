// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/keybase/go-logging"
)

var log = logging.Logger{Module: "test"}

func TestNewFile(t *testing.T) {
	filename := filepath.Join(os.TempDir(), "TestNewFile")
	defer os.Remove(filename)

	f := NewFile(filename, []byte("somedata"), 0)
	err := f.Save(log)
	if err != nil {
		t.Fatal(err)
	}
}

// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build windows

package process

import (
	"testing"

	"github.com/keybase/go-updater/util"
)

func TestTerminateAll(t *testing.T) {
	procPath := procPath(t, "testTerminateAll")
	defer util.RemoveFileAtPath(procPath)
	testTerminateAll(t, procPath, "exit status 1")
}

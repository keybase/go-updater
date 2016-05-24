// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build darwin

package process

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/keybase/go-ps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenDarwin(t *testing.T) {
	appPath := filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/Test.app")
	matcher := NewMatcher(appPath, PathEqual, testLog)
	defer TerminateAll(matcher, 200*time.Millisecond, testLog)
	err := OpenAppDarwin(appPath, testLog)
	assert.NoError(t, err)
}

func TestOpenDarwinError(t *testing.T) {
	binErr := filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/err.sh")
	appPath := filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/Test.app")
	err := openAppDarwin(binErr, appPath, time.Millisecond, testLog)
	assert.Error(t, err)
}

func TestFindPIDsLaunchd(t *testing.T) {
	procPath := "/sbin/launchd"
	matcher := NewMatcher(procPath, PathEqual, testLog)
	pids, err := findPIDsWithFn(ps.Processes, matcher.Fn(), testLog)
	assert.NoError(t, err)
	t.Logf("Pids: %#v", pids)
	require.True(t, len(pids) >= 1)
}

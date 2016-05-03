// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build darwin

package process

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenDarwin(t *testing.T) {
	appPath := filepath.Join(os.Getenv("GOPATH"), "src/github.com/keybase/go-updater/test/Test.app")
	defer TerminateAll(appPath, 200*time.Millisecond, log)

	err := OpenAppDarwin(appPath, log)
	require.NoError(t, err)
}

func TestFindPIDsLaunchd(t *testing.T) {
	pids, err := findPIDs("/sbin/launchd", log)
	assert.NoError(t, err)
	require.Equal(t, 1, len(pids))
	assert.Equal(t, 1, pids[0])
}

// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build darwin

package process

import (
	"os"
	"os/exec"
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
	t.Logf("Pids: %#v", pids)
	require.True(t, len(pids) >= 1)
}

func TestTerminateAll(t *testing.T) {
	start := func() int {
		cmd := exec.Command("sleep", "10")
		err := cmd.Start()
		require.NoError(t, err)
		require.NotNil(t, cmd.Process)
		return cmd.Process.Pid
	}

	pids := []int{}
	pids = append(pids, start())
	pids = append(pids, start())
	TerminateAll("/bin/sleep", time.Millisecond, log)
	assertTerminated(t, pids[0])
	assertTerminated(t, pids[1])
}

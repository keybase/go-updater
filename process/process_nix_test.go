// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build darwin

// TODO: Include linux above after we merge into master

package process

import (
	"os/exec"
	"testing"
	"time"

	"github.com/keybase/go-updater/util"
	"github.com/stretchr/testify/require"
)

func TestFindProcessWait(t *testing.T) {
	procPath := procPath(t)
	cmd := exec.Command(procPath, "10")
	defer cleanupProc(cmd, procPath)
	go func() {
		time.Sleep(10 * time.Millisecond)
		err := cmd.Start()
		require.NoError(t, err)
	}()
	procs, err := FindProcesses(NewMatcher(procPath, PathEqual, testLog), time.Millisecond, 0, testLog)
	require.NoError(t, err)
	require.Equal(t, 0, len(procs))

	// Wait up to second for process to be running
	procs, err = FindProcesses(NewMatcher(procPath, PathEqual, testLog), time.Second, 0, testLog)
	require.NoError(t, err)
	require.True(t, len(procs) == 1)
}

func TestTerminateAll(t *testing.T) {
	procPath := procPath(t)
	defer util.RemoveFileAtPath(procPath)
	start := func() int {
		cmd := exec.Command(procPath, "10")
		err := cmd.Start()
		require.NoError(t, err)
		require.NotNil(t, cmd.Process)
		return cmd.Process.Pid
	}

	pids := []int{}
	pids = append(pids, start())
	pids = append(pids, start())
	matcher := NewMatcher(procPath, PathEqual, testLog)
	TerminateAll(matcher, time.Millisecond, testLog)
	assertTerminated(t, pids[0], "signal: terminated")
	assertTerminated(t, pids[1], "signal: terminated")
}

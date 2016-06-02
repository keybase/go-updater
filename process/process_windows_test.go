// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build windows

package process

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func startProcess(t *testing.T, testCommand string) (string, int, *exec.Cmd) {
	path := filepath.Join(os.Getenv("GOPATH"), "bin", "test.exe")
	cmd := exec.Command(path, testCommand)
	err := cmd.Start()
	require.NoError(t, err)
	require.NotNil(t, cmd.Process)
	return path, cmd.Process.Pid, cmd
}

func TestTerminateAll(t *testing.T) {
	pids := []int{}
	path, pid1, cmd1 := startProcess(t, "sleep")
	defer cleanupProc(cmd1, "")
	_, pid2, cmd2 := startProcess(t, "sleep")
	defer cleanupProc(cmd2, "")
	pids = append(pids, pid1, pid2)
	TerminateAll(NewMatcher(path, PathEqual, testLog), time.Millisecond, testLog)
	assertTerminated(t, pids[0], "exit status 1")
	assertTerminated(t, pids[1], "exit status 1")
}

func TestFindProcessTest(t *testing.T) {
	path, _, cmd := startProcess(t, "sleep")
	defer cleanupProc(cmd, "")
	procs, err := FindProcesses(NewMatcher(path, PathEqual, testLog), 0, 0, testLog)
	require.NoError(t, err)
	// TODO: Fix flakiness where we might have more than 1 process here
	require.Equal(t, len(procs) >= 1)
	//assert.Equal(t, pid, procs[0].Pid())
}

// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package process

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/keybase/go-logging"
	"github.com/keybase/go-ps"
	"github.com/keybase/go-updater/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testLog = logging.Logger{Module: "test"}

var matchAll = func(p ps.Process) bool { return true }

func cleanupProc(cmd *exec.Cmd, procPath string) {
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
	_ = os.Remove(procPath)
}

func procPath(t *testing.T) string {
	// Copy sleep executable to tmp
	procPath := filepath.Join(os.TempDir(), "sleeptest")
	err := util.CopyFile("/bin/sleep", procPath, testLog)
	require.NoError(t, err)
	err = os.Chmod(procPath, 0777)
	require.NoError(t, err)
	// Temp dir might have symlinks in which case we need the eval'ed path
	procPath, err = filepath.EvalSymlinks(procPath)
	require.NoError(t, err)
	return procPath
}

func TestFindPIDsWithFn(t *testing.T) {
	pids, err := findPIDsWithFn(ps.Processes, matchAll, testLog)
	assert.NoError(t, err)
	assert.True(t, len(pids) > 1)

	fn := func() ([]ps.Process, error) {
		return nil, fmt.Errorf("Testing error")
	}
	processes, err := findPIDsWithFn(fn, matchAll, testLog)
	assert.Nil(t, processes)
	assert.Error(t, err)

	fn = func() ([]ps.Process, error) {
		return nil, nil
	}
	processes, err = findPIDsWithFn(fn, matchAll, testLog)
	assert.Equal(t, []int{}, processes)
	assert.NoError(t, err)
}

func TestTerminatePID(t *testing.T) {
	procPath := procPath(t)
	cmd := exec.Command(procPath, "10")
	err := cmd.Start()
	defer cleanupProc(cmd, procPath)
	require.NoError(t, err)
	require.NotNil(t, cmd.Process)

	err = TerminatePID(cmd.Process.Pid, time.Millisecond, testLog)
	assert.NoError(t, err)
}

func assertTerminated(t *testing.T, pid int, stateStr string) {
	process, err := os.FindProcess(pid)
	require.NoError(t, err)
	state, err := process.Wait()
	require.NoError(t, err)
	assert.Equal(t, stateStr, state.String())
}

func TestTerminatePIDInvalid(t *testing.T) {
	err := TerminatePID(-5, time.Millisecond, testLog)
	assert.Error(t, err)
}

func TestTerminateAllFn(t *testing.T) {
	fn := func() ([]ps.Process, error) {
		return nil, fmt.Errorf("Testing error")
	}
	terminateAll(fn, matchAll, time.Millisecond, testLog)

	fn = func() ([]ps.Process, error) {
		return nil, nil
	}
	terminateAll(fn, matchAll, time.Millisecond, testLog)
}

// process returns this test process' path that is running
func process(t *testing.T) (int, string) {
	pid := os.Getpid()
	proc, err := findProcessWithPID(pid)
	require.NoError(t, err)
	path, err := proc.Path()
	require.NoError(t, err)
	return pid, path
}

func TestFindProcessTest(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("Unsupported until we have process path on linux")
	}
	pid, path := process(t)
	procs, err := FindProcesses(NewMatcher(path, PathEqual, testLog), 0, 0, testLog)
	require.NoError(t, err)
	require.True(t, len(procs) == 1)
	assert.Equal(t, pid, procs[0].Pid())
}

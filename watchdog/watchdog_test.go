// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package watchdog

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/keybase/go-logging"
	"github.com/keybase/go-updater/process"
	"github.com/keybase/go-updater/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testLog = &logging.Logger{Module: "test"}

func TestWatchMultiple(t *testing.T) {
	procProgram1 := procProgram(t, "testWatch1", 60)
	procProgram2 := procProgram(t, "testWatch2", 60)
	defer util.RemoveFileAtPath(procProgram1.Path)
	defer util.RemoveFileAtPath(procProgram2.Path)

	delay := 10 * time.Millisecond

	err := Watch([]Program{procProgram1, procProgram2}, delay, testLog)
	require.NoError(t, err)

	matcher1 := process.NewMatcher(procProgram1.Path, process.PathEqual, testLog)
	procs1, err := process.FindProcesses(matcher1, time.Second, 200*time.Millisecond, testLog)
	require.NoError(t, err)
	assert.Equal(t, 1, len(procs1))

	matcher2 := process.NewMatcher(procProgram2.Path, process.PathEqual, testLog)
	procs2, err := process.FindProcesses(matcher2, time.Second, 200*time.Millisecond, testLog)
	require.NoError(t, err)
	assert.Equal(t, 1, len(procs2))
	proc2 := procs2[0]

	err = process.TerminatePID(proc2.Pid(), time.Millisecond, testLog)
	require.NoError(t, err)

	time.Sleep(2 * delay)

	// Check for restart
	procs2After, err := process.FindProcesses(matcher2, time.Second, time.Millisecond, testLog)
	require.NoError(t, err)
	assert.Equal(t, 1, len(procs2After))
}

// TestTerminateBeforeWatch checks to make sure any existing processes are
// terminated before a process is monitored.
func TestTerminateBeforeWatch(t *testing.T) {
	procProgram := procProgram(t, "testTerminateBeforeWatch", 60)
	defer util.RemoveFileAtPath(procProgram.Path)

	matcher := process.NewMatcher(procProgram.Path, process.PathEqual, testLog)

	// Launch program (so we can test it gets terminated on watch)
	err := exec.Command(procProgram.Path, procProgram.Args...).Start()
	require.NoError(t, err)

	procsBefore, err := process.FindProcesses(matcher, time.Second, time.Millisecond, testLog)
	require.NoError(t, err)
	require.Equal(t, 1, len(procsBefore))
	pidBefore := procsBefore[0].Pid()
	t.Logf("Pid before: %d", pidBefore)

	// Start watching
	err = Watch([]Program{procProgram}, 10*time.Millisecond, testLog)
	require.NoError(t, err)

	// Check again, and make sure it's a new process
	procsAfter, err := process.FindProcesses(matcher, time.Second, time.Millisecond, testLog)
	require.NoError(t, err)
	require.Equal(t, 1, len(procsAfter))
	pidAfter := procsAfter[0].Pid()
	t.Logf("Pid after: %d", pidAfter)

	assert.NotEqual(t, pidBefore, pidAfter)
}

func cleanupProc(cmd *exec.Cmd, procPath string) {
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
	_ = os.Remove(procPath)
}

func TestExitOnSuccess(t *testing.T) {
	procProgram := procProgram(t, "testExitOnSuccess", 0) // Don't sleep any
	procProgram.ExitOn = ExitOnSuccess
	defer util.RemoveFileAtPath(procProgram.Path)

	delay := 10 * time.Millisecond
	err := Watch([]Program{procProgram}, delay, testLog)
	require.NoError(t, err)

	matcher := process.NewMatcher(procProgram.Path, process.PathEqual, testLog)
	procsAfter, err := process.FindProcesses(matcher, 300*time.Millisecond, 300*time.Millisecond, testLog)
	require.NoError(t, err)
	assert.Equal(t, 0, len(procsAfter))
}

// procProgram returns a testable unique program at a temporary location that
// will run for the specified number of seconds.
func procProgram(t *testing.T, name string, sleepSeconds int) Program {
	if runtime.GOOS == "windows" {
		return procProgramWindows(t, name, sleepSeconds)
	}
	return procProgramNix(t, name, sleepSeconds)
}

func procProgramNix(t *testing.T, name string, sleepSeconds int) Program {
	// Copy sleep executable to tmp
	procPath := filepath.Join(os.TempDir(), name)
	err := util.CopyFile("/bin/sleep", procPath, testLog)
	require.NoError(t, err)
	err = os.Chmod(procPath, 0777)
	require.NoError(t, err)
	// Temp dir might have symlinks in which case we need the eval'ed path
	procPath, err = filepath.EvalSymlinks(procPath)
	require.NoError(t, err)
	return Program{
		Path: procPath,
		Args: []string{fmt.Sprintf("%d", sleepSeconds)},
	}
}

func procProgramWindows(t *testing.T, name string, sleepSeconds int) Program {
	// Copy executable to tmp
	procPath := filepath.Join(os.TempDir(), name+".exe")

	// Instead of sleep use ping with timeout, http://ss64.com/nt/sleep.html
	err := util.CopyFile(`c:\windows\system32\ping.exe`, procPath, testLog)
	require.NoError(t, err)

	return Program{
		Path: procPath,
		Args: []string{"-n", fmt.Sprintf("%d", sleepSeconds), "127.0.0.1"},
	}
}

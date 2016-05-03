// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package process

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/keybase/go-logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = logging.Logger{Module: "test"}

func TestFindPIDs(t *testing.T) {
	pids, err := findPIDs("", log)
	assert.NoError(t, err)
	assert.True(t, len(pids) > 1)
}

func TestTerminatePid(t *testing.T) {
	cmd := exec.Command("sleep", "10")
	err := cmd.Start()
	require.NoError(t, err)
	require.NotNil(t, cmd.Process)

	err = TerminatePid(cmd.Process.Pid, time.Millisecond, log)
	assert.NoError(t, err)
}

func assertTerminated(t *testing.T, pid int) {
	process, err := os.FindProcess(pid)
	require.NoError(t, err)
	state, err := process.Wait()
	require.NoError(t, err)
	assert.Equal(t, "signal: terminated", state.String())
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

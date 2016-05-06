// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package process

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/keybase/go-logging"
	"github.com/keybase/go-ps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = logging.Logger{Module: "test"}

func TestFindPIDs(t *testing.T) {
	pids, err := findPIDsWithFn(ps.Processes, "", log)
	assert.NoError(t, err)
	assert.True(t, len(pids) > 1)

	fn := func() ([]ps.Process, error) {
		return nil, fmt.Errorf("Testing error")
	}
	processes, err := findPIDsWithFn(fn, "", log)
	assert.Nil(t, processes)
	assert.Error(t, err)

	fn = func() ([]ps.Process, error) {
		return nil, nil
	}
	processes, err = findPIDsWithFn(fn, "", log)
	assert.Nil(t, processes)
	assert.NoError(t, err)
}

func TestTerminatePID(t *testing.T) {
	cmd := exec.Command("sleep", "10")
	err := cmd.Start()
	require.NoError(t, err)
	require.NotNil(t, cmd.Process)

	err = TerminatePID(cmd.Process.Pid, time.Millisecond, log)
	assert.NoError(t, err)
}

func assertTerminated(t *testing.T, pid int) {
	process, err := os.FindProcess(pid)
	require.NoError(t, err)
	state, err := process.Wait()
	require.NoError(t, err)
	assert.Equal(t, "signal: terminated", state.String())
}

func TestTerminatePIDInvalid(t *testing.T) {
	err := TerminatePID(-5, time.Millisecond, log)
	assert.Error(t, err)
}

func TestTerminateAllFn(t *testing.T) {
	fn := func() ([]ps.Process, error) {
		return nil, fmt.Errorf("Testing error")
	}
	terminateAll(fn, "", time.Millisecond, log)

	fn = func() ([]ps.Process, error) {
		return nil, nil
	}
	terminateAll(fn, "", time.Millisecond, log)
}
